<template>
    <div>
        <h1 class="text-3xl text-primary-800 dark:text-primary font-bold mb-4">Profile</h1>
        <section></section>
    </div>
</template>

<script setup lang="ts">
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
        phone: "123-456-7890",
        photo: "https://example.com/photo.jpg",
        birthDate: "2000-01-01",
        aboutMe: "I am a passionate software developer.",
        github: "https://github.com/johndoe",
        linkedIn: "https://linkedin.com/in/johndoe",
        studentId: "S12345678",
        major: "Software Engineering",
        status: "Undergraduate",
        statusFile: "https://example.com/status.pdf"
    }
};
</script>
