# KU-Work Frontend

A modern Vue.js frontend application built with Nuxt 3, providing a user-friendly interface for the KU-Work user management system.

## Prerequisites

- **Node.js** 18+
- **Bun**
- **Backend API** running on http://localhost:8000 (default)

## Installation

### 1. Install Dependencies

Using Bun:

```bash
bun install
```

### 2. Environment Configuration

Copy the sample environment file and configure your API settings:

```bash
# Linux/Unix/MacOS
cp sample.env .env

# Windows
copy sample.env .env

# Edit .env with your configuration
```

Edit `.env` with your backend API URL:

```env
API_BASE_URL=http://localhost:8000
GOOGLE_CLIENT_ID=your_google_client_id_here
TURNSTILE_CLIENT_TOKEN=your_turnstile_client_token
```

## Development

### Start Development Server

```bash
bun run dev
```

The application will be available at `http://localhost:3000`

## Terms & Policies (TXT/PDF Support)

To support legal documents (Terms of Service, Privacy Policy, role supplements, Google Sign‑In Notice) in both `.txt` and `.pdf` formats: Configuration, role mapping, and dynamic TXT/PDF resolution are centralized in `app/config/terms.ts`.

1. Location  
   Place files in `public/terms/` (e.g., `public/terms/ku_work_core_terms.txt` or `public/terms/ku_work_core_terms.pdf`).

2. Naming Conventions  
   - Core Terms: `ku_work_core_terms.{txt|pdf}`  
   - Student Supplement: `ku_work_terms.{txt|pdf}`  
   - Company Supplement: `company_terms.{txt|pdf}`  
   - Privacy Policy: `privacy_policy.{txt|pdf}`  
   - Google OAuth Notice: `google_oauth_notice.{txt|pdf}`  
   Use lowercase with underscores for consistency.

3. Using PDFs  
   - If you replace a `.txt` with a `.pdf`, keep the same base name.  
   - In `app/config/terms.ts`, set `preferredExtension` to `"pdf"` for that document so the UI loads the PDF by default.  
   - PDFs are embedded in a modal (toolbar hidden); `.txt` files render inline with wrapping.

4. Adding a New Document  
   - Create the file in `public/terms/`.  
   - Add an entry to `TERMS_DOCUMENTS` in `app/config/terms.ts` (and set `preferredExtension` accordingly).  
   - If it must be accepted for a role, add the document key to the role in `ROLE_REQUIRED_DOC_KEYS`.  
   Example:  
   ```ts
   {
     key: "export-policy",
     title: "Data Export Policy",
     baseName: "export_policy",
     preferredExtension: "pdf",
     required: false,
     version: "1.0.0",
     category: "other",
   }
   // If required for a role, map it in ROLE_REQUIRED_DOC_KEYS:
   // e.g., student: ["ku-work-core-terms", "ku-work-student-terms", "export-policy"]
   ```

5. Versioning & Hashes (Optional)  
   - Add a version line at the top: `Version: 1.0.0`  
   - For audit logging later, compute a hash (e.g., SHA256) at build time or on first fetch and store alongside user consent.

6. Fallback Behavior  
   - If a PDF cannot render (browser plugin disabled), the UI shows a link allowing the user to open it in a new tab.  
   - Text files are rendered verbatim in a `<pre>` block with `whitespace-pre-wrap`.

7. Acceptance Flow  
   - Student: must accept Core + Student supplement.  
   - Company: must accept Core + Privacy Policy + Company supplement.  
   - Viewer (Google Sign‑In): implicitly accepts Core & Privacy Policy; explicit notice is the Google OAuth Notice.  
   - Add any new required doc by inserting it into `app/config/terms.ts` (TERMS_DOCUMENTS) and mapping its key in `ROLE_REQUIRED_DOC_KEYS` for the role. Set `required` appropriately.

8. Changing Formats Safely  
   - Replace `*.txt` with `*.pdf` and keep identical base name to avoid path updates.  
   - Clear browser cache if testing stale documents.

9. Do Not  
   - Rename keys in `app/config/terms.ts` (TERMS_DOCUMENTS) without migrating stored acceptance records (if you later implement backend consent logging).  
   - Mix different versions under the same filename—always increment version inside the document.

10. Recommended Next Steps (Future)  
   - Implement backend endpoint `POST /consents` to record `{ userId, key, version, hash, acceptedAt }`.  
   - Add automated diff alert if a document changes without version bump.

This section ensures legal documents remain flexible and easy to maintain in either plain text or PDF formats.

### Available Scripts

| Command            | Description                             |
| ------------------ | --------------------------------------- |
| `bun run dev`      | Start development server                |
| `bun run build`    | Build for production                    |
| `bun run preview`  | Preview production build                |
| `bun run generate` | Generate static site                    |
| `bun lint`         | Check project file with eslint          |
| `bun lint:fix`     | Fix project file formatting with eslint |
| `bun format`       | Check project file format with prettier |
| `bun format:fix`   | Fix project file format with prettier   |

## Production Deployment

### Build Application

```bash
bun run build
```

### Preview Production Build

```bash
bun run preview
```

### Static Generation (Optional)

For static hosting:

```bash
bun run generate
```
