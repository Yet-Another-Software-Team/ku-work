interface Profile {
    profile: {
        name: string;
        id: string;
        approved: boolean;
        created: string;
        phone: string;
        photo: string;
        birthDate: string;
        aboutMe: string;
        github: string;
        linkedIn: string;
        studentId: string;
        major: "Software Engineering" | "Computer Science";
        status: "Graduated" | "Undergraduate";
        statusFile: string;
    };
}

interface CompanyProfile {
    profile: {
        name: string;
        id: string;
        approved: boolean;
        phone: string;
        logo: string;
        banner: string;
        aboutUs: string;
        website: string;
        contact: string;
        address: string;
    };
}

interface JobApplication {
    id: string;
    createdAt: string;
    name: string;
    companyId: string;
    position: string;
    duration: string;
    description: string;
    location: string;
    jobType: string;
    experienceType: string;
    minSalary: number;
    maxSalary: number;
    approved: boolean;
    logo: string;
    open?: boolean;
    pending?: number;
    accepted?: number;
    rejected?: number;
}

interface Job {
    jobs: JobApplication[];
}

const mockUserData: Profile = {
    profile: {
        name: "John Doe",
        id: "123456",
        approved: true,
        created: "2023-01-01",
        phone: "012-345-6789",
        photo: "",
        birthDate: "2003-01-01",
        aboutMe:
            "Hello! I'm John, a passionate software engineering student with a love for coding and problem-solving. I enjoy working on innovative projects and collaborating with others to create impactful solutions. REALLY LONG TEXT TO TEST THE LAYOUT. Hello! I'm John, a passionate software engineering student with a love for coding and problem-solving. I enjoy working on innovative projects and collaborating with others to create impactful solutions.",
        github: "https://github.com",
        linkedIn: "https://linkedin.com/",
        studentId: "6xxxxxxxxx",
        major: "Software Engineering",
        status: "Undergraduate",
        statusFile: "https://example.com/status.pdf",
    },
};

const multipleMockUserData: Profile[] = [
    {
        profile: {
            name: "John Doe",
            id: "123456",
            approved: true,
            created: "2023-01-01",
            phone: "012-345-6789",
            photo: "",
            birthDate: "2003-01-01",
            aboutMe:
                "Hello! I'm John, a passionate software engineering student with a love for coding and problem-solving. I enjoy working on innovative projects and collaborating with others to create impactful solutions. REALLY LONG TEXT TO TEST THE LAYOUT. Hello! I'm John, a passionate software engineering student with a love for coding and problem-solving. I enjoy working on innovative projects and collaborating with others to create impactful solutions.",
            github: "https://github.com",
            linkedIn: "https://linkedin.com/",
            studentId: "6xxxxxxxxx",
            major: "Software Engineering",
            status: "Undergraduate",
            statusFile: "https://example.com/status.pdf",
        },
    },
    {
        profile: {
            name: "A Very Very Very Very Very Very Very Long Name To Break Layout",
            id: "9999999999999999999999", // Very long ID
            approved: true,
            created: "2025-12-31", // Future date
            phone: "999-999-999999999999", // Very long phone
            photo: "/images/background.png", // Placeholder image
            birthDate: "2010-12-31", // Young student
            aboutMe: "Lorem ipsum ".repeat(50), // Extremely long text
            github: "not-a-url", // Invalid URL
            linkedIn: "ftp://invalid.link", // Malformed URL
            studentId: "12345678901234567890", // Very long studentId
            major: "Computer Science",
            status: "Graduated",
            statusFile: "https://example.com/very/long/path/to/status/file/status.pdf",
        },
    },
    {
        profile: {
            name: "Jane Doe",
            id: "EDGE-123",
            approved: false,
            created: new Date().toISOString().split("T")[0] ?? "", // Today's date
            phone: "+1 (555) 123-4567", // International format
            photo: "broken-link.jpg", // Broken image link
            birthDate: "2000-02-29", // Leap year birthday
            aboutMe: "👩‍💻🔥🚀 Emojis & special chars test — 中文测试 — عربى اختبار",
            github: "https://github.com/jane_doe",
            linkedIn: "https://linkedin.com/in/jane_doe",
            studentId: "EDGE-STU-!@#$%^", // Special characters
            major: "Software Engineering",
            status: "Undergraduate",
            statusFile: "javascript:alert('XSS')", // Malicious-looking link
        },
    },
];

const mockCompanyData: CompanyProfile = {
    profile: {
        name: "TechNova Solutions Co., Ltd.",
        id: "COMP-2025-001",
        approved: true,
        phone: "+66 2 123 4567",
        logo: "https://placehold.co/200x200?text=Logo",
        banner: "https://placehold.co/1200x300?text=Company+Banner",
        aboutUs:
            "TechNova Solutions is a leading provider of innovative IT and data analytics solutions. We specialize in cloud computing, business intelligence, and AI-driven platforms to help companies scale efficiently in the digital era.",
        website: "https://www.technova.co.th",
        contact: "contact@technova.co.th",
        address: "99/9 Rama IX Road, Huai Khwang, Bangkok 10310, Thailand",
    },
};

const mockJobData: Job = {
    jobs: [
        {
            id: "1",
            createdAt: "2025-08-23T18:17:51.746604+07:00",
            name: "IT City Banna",
            companyId: "2d7f403e-8831-4805-a15b-65d48f6db46e",
            position: "IT Support",
            duration: "forever",
            description: "IT position",
            location: "thailand",
            jobType: "casual",
            experienceType: "newgrad",
            minSalary: 1,
            maxSalary: 1,
            approved: false,
            logo: "",
        },
        {
            id: "2",
            createdAt: "2025-08-20T12:05:21.123456+07:00",
            name: "Software Engineering",
            companyId: "3e5d27c1-98f2-4b6b-a3c9-7e0f32e8f888",
            position: "Frontend Developer",
            duration: "Contract",
            description: "Work on building modern web applications with Vue.js and TypeScript.",
            location: "Thailand",
            jobType: "Contract",
            experienceType: "Junior",
            minSalary: 30000,
            maxSalary: 45000,
            approved: true,
            logo: "",
        },
        {
            id: "3",
            createdAt: "2025-07-15T09:45:00.654321+07:00",
            name: "Finance",
            companyId: "7a1e3f0a-43c7-4f2e-9b19-11234d9abc99",
            position: "Financial Analyst",
            duration: "Permanent",
            description:
                "Analyze financial data and create reports to assist management decisions.",
            location: "Singapore",
            jobType: "Full Time",
            experienceType: "Mid-level",
            minSalary: 50000,
            maxSalary: 70000,
            approved: true,
            logo: "",
        },
    ],
};

export { mockUserData, multipleMockUserData, mockCompanyData, mockJobData };
export type { Profile, CompanyProfile, JobApplication, Job };
