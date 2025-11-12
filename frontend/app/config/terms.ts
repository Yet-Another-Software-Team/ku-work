/**
 * Centralized Terms & Policies configuration for KU Work.
 *
 * Supports both .txt and .pdf documents. Components can rely on this single source
 * to determine which documents must be presented and accepted for each role.
 *
 * Migration to PDF:
 * - Replace the `.txt` file with a `.pdf` using the same base name.
 * - Set `preferredExtension` to "pdf" for that document (optional).
 * - Components can call `resolveDocumentSrc(doc)` to automatically prefer the PDF
 *   if it exists, otherwise fall back to `.txt`.
 *
 * Acceptance & Auditing:
 * - Each document includes an optional `version` field; bump it whenever the
 *   content changes materially.
 * - A `contentHash` can be populated at runtime using `computeDocumentHash`
 *   for audit logging (e.g., POST /consents).
 */

export type TermsCategory = "core" | "privacy" | "student" | "company" | "oauth" | "other";

export interface TermsDocument {
    /**
     * Unique stable key used for acceptance tracking
     */
    key: string;
    /**
     * Human-readable title shown in UI
     */
    title: string;
    /**
     * Base filename without extension (e.g., 'ku_work_core_terms')
     */
    baseName: string;
    /**
     * Explicit src override. If omitted, will be resolved from baseName + extension.
     * Provide full path starting with /terms/ if used.
     */
    src?: string;
    /**
     * Preferred extension if both .pdf and .txt are present. Defaults to 'txt'.
     */
    preferredExtension?: "txt" | "pdf";
    /**
     * Whether the document must be accepted for the user role to proceed.
     */
    required: boolean;
    /**
     * Semantic version or date string (e.g., '1.0.0' or '2025-01-18')
     */
    version?: string;
    /**
     * Category for filtering/grouping in UI
     */
    category?: TermsCategory;
    /**
     * Runtime-populated SHA-256 or similar hash of fetched content (for audit)
     */
    contentHash?: string;
    /**
     * Optional description / tooltip text
     */
    description?: string;
}

/**
 * Canonical list of all terms/policy documents.
 * NOTE: Keep keys stable—changing a key requires migration of stored consents.
 */
export const TERMS_DOCUMENTS: TermsDocument[] = [
    {
        key: "ku-work-core-terms",
        title: "KU Work Core Terms of Use and Privacy Notice",
        baseName: "ku_work_core_terms",
        preferredExtension: "pdf",
        required: true,
        version: "1.0.0",
        category: "core",
        description: "Global baseline terms & privacy notice applying to all user roles.",
    },
    {
        key: "ku-work-student-terms",
        title: "Student Terms Supplement",
        baseName: "ku_work_terms",
        preferredExtension: "pdf",
        required: true,
        version: "1.0.0",
        category: "student",
        description: "Role-specific student/KU supplement to the core terms & privacy notice.",
    },
    {
        key: "company-terms",
        title: "Company Terms Supplement",
        baseName: "company_terms",
        preferredExtension: "pdf",
        required: true,
        version: "1.0.0",
        category: "company",
        description: "Company-specific supplement detailing obligations & candidate data handling.",
    },
    {
        key: "privacy-policy",
        title: "Privacy Policy",
        baseName: "privacy_policy",
        preferredExtension: "pdf",
        required: true,
        version: "1.0.0",
        category: "privacy",
        description: "Concise overview; full legal detail in Core & supplements.",
    },
    {
        key: "google-oauth-notice",
        title: "Google Sign‑In Notice & Consent",
        baseName: "google_oauth_notice",
        preferredExtension: "pdf",
        required: true,
        version: "1.0.0",
        category: "oauth",
        description: "Explains Google OAuth data exchange & consent for viewer pre-registration.",
    },
];

/**
 * User roles recognized by the consent system.
 */
export type UserRole = "viewer" | "student" | "company" | "admin";

/**
 * Mapping of required document keys by role.
 * Viewer implicitly acknowledges core + privacy + oauth; Student/Company must explicitly accept.
 */
const ROLE_REQUIRED_DOC_KEYS: Record<UserRole, string[]> = {
    viewer: ["ku-work-core-terms", "privacy-policy", "google-oauth-notice"],
    student: ["ku-work-core-terms", "ku-work-student-terms"],
    company: ["ku-work-core-terms", "privacy-policy", "company-terms"],
    admin: ["ku-work-core-terms", "privacy-policy"], // adjust if admin needs supplements
};

/**
 * Resolve a document by its key.
 */
export function getDocumentByKey(key: string): TermsDocument | undefined {
    return TERMS_DOCUMENTS.find((d) => d.key === key);
}

/**
 * Get all documents relevant for a given role (required first, then optional).
 * Optional documents (required=false) can be appended for display if added later.
 */
export function getDocumentsForRole(role: UserRole): TermsDocument[] {
    const requiredKeys = new Set(ROLE_REQUIRED_DOC_KEYS[role] || []);
    const required = TERMS_DOCUMENTS.filter((d) => requiredKeys.has(d.key));
    const optional = TERMS_DOCUMENTS.filter(
        (d) => !requiredKeys.has(d.key) && d.required === false
    );
    return [...required, ...optional];
}

/**
 * Return only required docs for a role.
 */
export function getRequiredDocumentsForRole(role: UserRole): TermsDocument[] {
    const keys = new Set(ROLE_REQUIRED_DOC_KEYS[role] || []);
    return TERMS_DOCUMENTS.filter((d) => keys.has(d.key));
}

/**
 * Determine final src path for a document:
 * - If doc.src explicitly provided, use it.
 * - Else attempt preferred extension first; consumer can optionally verify existence.
 */
export function buildDocumentSrc(doc: TermsDocument): string {
    if (doc.src) return doc.src;
    const ext = doc.preferredExtension || "txt";
    return `/terms/${doc.baseName}.${ext}`;
}

/**
 * Helper: is document a PDF (based on chosen src).
 */
export function isPdf(doc: TermsDocument): boolean {
    const src = buildDocumentSrc(doc).toLowerCase();
    return src.endsWith(".pdf");
}

/**
 * Attempt to resolve to a PDF if present, falling back to TXT.
 * This performs HEAD requests sequentially. Use sparingly to avoid latency;
 * for bulk resolution, pre-warm or rely on naming conventions.
 */
export async function resolveDocumentSrc(doc: TermsDocument): Promise<string> {
    if (doc.src) return doc.src;
    const pdf = `/terms/${doc.baseName}.pdf`;
    const txt = `/terms/${doc.baseName}.txt`;

    // Try PDF first
    try {
        const headPdf = await fetch(pdf, { method: "HEAD" });
        if (headPdf.ok) return pdf;
    } catch {
        /* ignore */
    }

    // Fallback TXT
    return txt;
}

/**
 * Compute SHA-256 hash of the document's textual content (for audit logging).
 * If the document is a PDF, this will hash the binary content (ArrayBuffer).
 */
export async function computeDocumentHash(doc: TermsDocument): Promise<string | null> {
    try {
        const src = await resolveDocumentSrc(doc);
        const res = await fetch(src);
        if (!res.ok) return null;

        let data: BufferSource;
        if (src.toLowerCase().endsWith(".pdf")) {
            data = await res.arrayBuffer();
        } else {
            const text = await res.text();
            data = new TextEncoder().encode(text);
        }

        const digest = await crypto.subtle.digest("SHA-256", data);
        const bytes = Array.from(new Uint8Array(digest));
        return bytes.map((b) => b.toString(16).padStart(2, "0")).join("");
    } catch {
        return null;
    }
}

/**
 * Convenience loader:
 * Fetch content (string for txt; binary base64 for pdf) for unified downstream handling.
 */
export async function fetchDocumentContent(
    doc: TermsDocument
): Promise<{ content: string; isPdf: boolean } | null> {
    try {
        const src = await resolveDocumentSrc(doc);
        const isPdfFlag = src.toLowerCase().endsWith(".pdf");
        const res = await fetch(src);
        if (!res.ok) return null;

        if (isPdfFlag) {
            const buf = await res.arrayBuffer();
            // Convert binary to base64 for potential inline display fallback
            const bytes = new Uint8Array(buf);
            let binary = "";
            bytes.forEach((b) => (binary += String.fromCharCode(b)));
            const base64 = btoa(binary);
            return { content: base64, isPdf: true };
        } else {
            const text = await res.text();
            return { content: text, isPdf: false };
        }
    } catch {
        return null;
    }
}

/**
 * Normalize acceptance payload before sending to backend audit endpoint.
 */
export interface AcceptanceRecord {
    key: string;
    version?: string;
    hash?: string;
    acceptedAt: string; // ISO timestamp
    roleContext?: UserRole;
}

export function buildAcceptanceRecord(
    doc: TermsDocument,
    role?: UserRole,
    hash?: string
): AcceptanceRecord {
    return {
        key: doc.key,
        version: doc.version,
        hash,
        acceptedAt: new Date().toISOString(),
        roleContext: role,
    };
}

/**
 * Export a keyed map for quick access in components.
 */
export const TERMS_MAP: Record<string, TermsDocument> = TERMS_DOCUMENTS.reduce<
    Record<string, TermsDocument>
>((acc, d) => {
    acc[d.key] = d;
    return acc;
}, {});
