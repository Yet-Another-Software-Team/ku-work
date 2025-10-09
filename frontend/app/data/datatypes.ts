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
        email: string;
        files: { name: string; url: string }[];
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

// eslint-disable-next-line @typescript-eslint/no-unused-vars
interface Company {
    id: string;
    createdAt: string;
    email: string;
    phone: string;
    photoId: string;
    bannerId: string;
    about: string;
    website: string;
    address: string;
    city: string;
    country: string;
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
    approved: boolean;
    open: boolean;
}

interface Job {
    jobs: JobPost[];
}

export type { Profile, CompanyProfile, JobPost, Job };
