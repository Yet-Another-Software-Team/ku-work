interface ProfileInformation {
    name?: string;
    id: string;
    approved: boolean;
    created: string;
    phone: string;
    photo: {
        id: string;
        createdAt: string;
        updatedAt: string;
        userId: string;
        fileType: string;
        category: string;
    };
    photoId: string;
    birthDate: string;
    aboutMe: string;
    github: string;
    linkedIn: string;
    studentId: string;
    major?: "Software Engineering" | "Computer Science";
    status?: "Graduated" | "Undergraduate" | "Current Student";
    statusFileId: string;
    approvalStatus: string;
    statusFile: {
        id: string;
        createdAt: string;
        updatedAt: string;
        userId: string;
        fileType: string;
        category: string;
    };
    email?: string;
    firstName?: string;
    lastName?: string;
}

interface Profile {
    profile: ProfileInformation;
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

// eslint-disable-next-line @typescript-eslint/no-unused-vars
interface Company {
    id: string;
    createdAt: string;
    email: string;
    phone: string;
    photoId: string;
    bannerId: string;
    about: string;
    site: string;
    address: string;
    city: string;
    country: string;
}

interface CreateJobPost {
    name: string;
    position: string;
    duration: string | undefined;
    description: string | undefined;
    location: string;
    jobType: string | undefined;
    experience: string | undefined;
    minSalary: number | undefined;
    maxSalary: number | undefined;
    open: boolean;
}

interface EditJobPost {
    name: string;
    position: string;
    duration: string;
    description: string;
    location: string;
    jobType: string;
    experience: string;
    minSalary: number;
    maxSalary: number;
    open: boolean;
}

interface JobPost {
    companyName: string;
    id: number;
    createdAt: string;
    name: string;
    companyId: string;
    photoId: string;
    bannerId: string;
    position: string;
    duration: string;
    description: string;
    location: string;
    jobType: string;
    experience: string;
    minSalary: number;
    maxSalary: number;
    approvalStatus: "pending" | "accepted" | "rejected";
    open: boolean;
    accepted?: number;
    rejected?: number;
    pending?: number;
}

interface Job {
    jobs: JobPost[];
}

interface JobApplicationFile {
    id: string;
    createdAt: string;
    userId: string;
    fileType: string;
    category: string;
}

interface JobApplication {
    id: number;
    createdAt: string;
    jobId: number;
    userId: string;
    phone: string;
    email: string;
    status: string;
    username: string;
    files: JobApplicationFile[];
}

const mockJobApplicationData: JobApplication = {
    id: 1,
    createdAt: "2023-10-01T12:00:00Z",
    jobId: 101,
    userId: "12345678-1234-1234-1234-123456789012",
    phone: "012-345-6789",
    email: "abc@ku.th",
    status: "pending",
    username: "John Doe",
    files: [
        {
            id: "file-uuid-1",
            createdAt: "2023-10-01T12:00:00Z",
            userId: "12345678-1234-1234-1234-123456789012",
            fileType: "resume",
            category: "application",
        },
    ],
};

const mockUserData: Profile = {
    profile: {
        name: "John Doe",
        id: "123456",
        approved: true,
        created: "2023-01-01",
        phone: "012-345-6789",
        photo: {
            id: "",
            createdAt: "",
            updatedAt: "",
            userId: "",
            fileType: "",
            category: "",
        },
        photoId: "",
        birthDate: "2003-01-01",
        aboutMe:
            "Hello! I'm John, a passionate software engineering student with a love for coding and problem-solving. I enjoy working on innovative projects and collaborating with others to create impactful solutions. REALLY LONG TEXT TO TEST THE LAYOUT. Hello! I'm John, a passionate software engineering student with a love for coding and problem-solving. I enjoy working on innovative projects and collaborating with others to create impactful solutions.",
        github: "https://github.com",
        linkedIn: "https://linkedin.com/",
        studentId: "6xxxxxxxxx",
        major: "Software Engineering",
        status: "Undergraduate",
        statusFileId: "",
        approvalStatus: "",
        statusFile: {
            id: "",
            createdAt: "",
            updatedAt: "",
            userId: "",
            fileType: "",
            category: "",
        },
    },
};

const multipleMockUserData: ProfileInformation[] = [
    {
        name: "John Doe",
        id: "123456",
        approved: true,
        created: "2023-01-01",
        phone: "012-345-6789",
        photo: {
            id: "",
            createdAt: "",
            updatedAt: "",
            userId: "",
            fileType: "",
            category: "",
        },
        photoId: "",
        birthDate: "2003-01-01",
        aboutMe:
            "Hello! I'm John, a passionate software engineering student with a love for coding and problem-solving. I enjoy working on innovative projects and collaborating with others to create impactful solutions. REALLY LONG TEXT TO TEST THE LAYOUT. Hello! I'm John, a passionate software engineering student with a love for coding and problem-solving. I enjoy working on innovative projects and collaborating with others to create impactful solutions.",
        github: "https://github.com",
        linkedIn: "https://linkedin.com/",
        studentId: "6xxxxxxxxx",
        major: "Software Engineering",
        status: "Undergraduate",
        statusFileId: "",
        approvalStatus: "",
        statusFile: {
            id: "",
            createdAt: "",
            updatedAt: "",
            userId: "",
            fileType: "",
            category: "",
        },
    },
    {
        name: "A Very Very Very Very Very Very Very Long Name To Break Layout",
        id: "9999999999999999999999", // Very long ID
        approved: true,
        created: "2025-12-31", // Future date
        phone: "999-999-999999999999", // Very long phone
        photo: {
            id: "",
            createdAt: "",
            updatedAt: "",
            userId: "",
            fileType: "",
            category: "",
        },
        photoId: "", // Add a value for photoId
        birthDate: "2010-12-31", // Young student
        aboutMe: "Lorem ipsum ".repeat(50), // Extremely long text
        github: "not-a-url", // Invalid URL
        linkedIn: "ftp://invalid.link", // Malformed URL
        studentId: "12345678901234567890", // Very long studentId
        major: "Computer Science",
        status: "Graduated",
        statusFileId: "",
        approvalStatus: "",
        statusFile: {
            id: "",
            createdAt: "",
            updatedAt: "",
            userId: "",
            fileType: "",
            category: "",
        },
    },
    {
        name: "Jane Doe",
        id: "EDGE-123",
        approved: false,
        created: new Date().toISOString().split("T")[0] ?? "", // Today's date
        phone: "+1 (555) 123-4567", // International format
        photo: {
            id: "",
            createdAt: "",
            updatedAt: "",
            userId: "",
            fileType: "",
            category: "",
        },
        photoId: "", // Add a value for photoId
        birthDate: "2000-02-29", // Leap year birthday
        aboutMe: "👩‍💻🔥🚀 Emojis & special chars test — 中文测试 — عربى اختبار",
        github: "https://github.com/jane_doe",
        linkedIn: "https://linkedin.com/in/jane_doe",
        studentId: "EDGE-STU-!@#$%^", // Special characters
        major: "Software Engineering",
        status: "Undergraduate",
        statusFileId: "",
        approvalStatus: "",
        statusFile: {
            id: "",
            createdAt: "",
            updatedAt: "",
            userId: "",
            fileType: "",
            category: "",
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
            createdAt: "2025-08-23T18:17:51.746604+07:00",
            name: "IT City Banna",
            companyId: "2d7f403e-8831-4805-a15b-65d48f6db46e",
            position: "IT Support",
            duration: "forever",
            description: "IT position",
            location: "thailand",
            jobType: "casual",
            experience: "newgrad",
            minSalary: 1,
            maxSalary: 1,
            approvalStatus: "rejected",
            photoId: "305419d1-2d0e-4b0b-9137-f4689e39198d",
            bannerId: "28769ce2-7a40-4ff3-8067-c2b56f926518",
            companyName: "AA",
            open: true,
        },
        {
            id: 2,
            createdAt: "2025-08-20T12:05:21.123456+07:00",
            name: "Software Engineering",
            companyId: "3e5d27c1-98f2-4b6b-a3c9-7e0f32e8f888",
            position: "Frontend Developer",
            duration: "Contract",
            description: "Work on building modern web applications with Vue.js and TypeScript.",
            location: "Thailand",
            jobType: "Contract",
            experience: "Junior",
            minSalary: 30000,
            maxSalary: 45000,
            approvalStatus: "accepted",
            photoId: "305419d1-2d0e-4b0b-9137-f4689e39198d",
            bannerId: "28769ce2-7a40-4ff3-8067-c2b56f926518",
            companyName: "AA",
            open: false,
        },
        {
            id: 3,
            createdAt: "2025-07-15T09:45:00.654321+07:00",
            name: "Finance",
            companyId: "7a1e3f0a-43c7-4f2e-9b19-11234d9abc99",
            position: "Financial Analyst",
            duration: "Permanent",
            description:
                "Analyze financial data and create reports to assist management decisions.",
            location: "Singapore",
            jobType: "Full Time",
            experience: "Mid-level",
            minSalary: 50000,
            maxSalary: 55000,
            approvalStatus: "accepted",
            photoId: "305419d1-2d0e-4b0b-9137-f4689e39198d",
            bannerId: "28769ce2-7a40-4ff3-8067-c2b56f926518",
            companyName: "AA",
            open: true,
        },
    ],
};

export { mockUserData, multipleMockUserData, mockCompanyData, mockJobData, mockJobApplicationData };
export type {
    ProfileInformation,
    Profile,
    CompanyProfile,
    CreateJobPost,
    EditJobPost,
    JobPost,
    Job,
    JobApplication,
};
