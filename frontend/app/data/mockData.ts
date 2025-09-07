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

const mockData: Profile = {
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

const multipleMockData: Profile[] = [
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

export { mockData, multipleMockData };
