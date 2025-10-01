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

interface JobPost {
    id: number;
    createdAt: string;
    name: string;
    companyId: string;
    company: {
        id: string;
        User: {
            ID: string;
            CreatedAt: string;
            UpdatedAt: string;
            DeletedAt: string | null;
            Username: string;
        };
        createdAt: string;
        email: string;
        phone: string;
        photoId: string;
        bannerId: string;
        address: string;
        city: string;
        country: string;
    };
    position: string;
    duration: string;
    description: string;
    location: string;
    jobType: string;
    experienceType: string;
    minSalary: number;
    maxSalary: number;
    approved: boolean;
    open: boolean;
}

interface Job {
    jobs: JobPost[];
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
            id: 1,
            createdAt: "2025-10-01T22:13:47.233059+07:00",
            name: "Software Engineer",
            companyId: "734fc1e6-34b5-4810-b139-ce575b1a52c6",
            company: {
                id: "734fc1e6-34b5-4810-b139-ce575b1a52c6",
                User: {
                    ID: "734fc1e6-34b5-4810-b139-ce575b1a52c6",
                    CreatedAt: "2025-10-01T22:06:51.157972+07:00",
                    UpdatedAt: "2025-10-01T22:06:51.157972+07:00",
                    DeletedAt: null,
                    Username: "AA",
                },
                createdAt: "2025-10-01T22:06:52.152089+07:00",
                email: "AAA@AAA.AAA",
                phone: "+6699999999999",
                photoId: "305419d1-2d0e-4b0b-9137-f4689e39198d",
                bannerId: "28769ce2-7a40-4ff3-8067-c2b56f926518",
                address: "That St.",
                city: "Quebec",
                country: "Canada",
            },
            position: "Software Engineer",
            duration: "6 month",
            description: "Do Software Engineering related works 😊",
            location: "New York",
            jobType: "fulltime",
            experienceType: "junior",
            minSalary: 50000,
            maxSalary: 55000,
            approved: true,
            open: true,
        },
    ],
};

export { mockUserData, multipleMockUserData, mockCompanyData, mockJobData };
export type { Profile, CompanyProfile, JobPost, Job };
