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
    applied?: boolean;
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
