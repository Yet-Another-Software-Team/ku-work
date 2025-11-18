<template>
    <div class="pb-12 px-4 sm:px-6 lg:px-8">
        <h1 class="text-3xl font-bold text-gray-800 dark:text-gray-100 mb-6 text-center">
            Google OAuth Login Notice
        </h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 text-center mb-8">
            Last Updated: November 17, 2025
        </p>

        <section class="mb-8">
            <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-4">
                1. What Information We Access
            </h2>
            <p class="text-gray-700 dark:text-gray-300 leading-relaxed mb-4">
                When you choose to log in or sign up using your Google account, we request access to
                your basic profile information from Google. This is designed to make your sign-up
                process faster and more secure.
            </p>
            <p class="text-gray-700 dark:text-gray-300 leading-relaxed">
                The specific information we access includes:
            </p>
            <ul class="list-disc list-inside text-gray-700 dark:text-gray-300 ml-4 my-4 space-y-2">
                <li><strong>Email Address:</strong> To uniquely identify you in our system.</li>
                <li>
                    <strong>Full Name:</strong> To personalize your experience and pre-fill your
                    profile.
                </li>
            </ul>
        </section>

        <section class="mb-8">
            <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-4">
                2. How We Use Your Information
            </h2>
            <p class="text-gray-700 dark:text-gray-300 leading-relaxed mb-4">
                The information obtained from Google is used solely for the following purposes:
            </p>
            <ul class="list-disc list-inside text-gray-700 dark:text-gray-300 ml-4 mb-4 space-y-2">
                <li>
                    <strong>Account Creation:</strong> To automatically create an account for you on
                    our platform without needing a separate password.
                </li>
                <li>
                    <strong>Authentication:</strong> To securely log you into your account each time
                    you visit.
                </li>
                <li>
                    <strong>Profile Setup:</strong> To provide an initial setup for your user
                    profile, which you can edit at any time.
                </li>
            </ul>
        </section>

        <section>
            <h2 class="text-2xl font-bold text-gray-800 dark:text-gray-100 mb-4">
                3. Important Security Notes
            </h2>
            <p class="text-gray-700 dark:text-gray-300 leading-relaxed font-semibold mb-4">
                Your security is our top priority. Please be aware of the following:
            </p>
            <ul class="list-disc list-inside text-gray-700 dark:text-gray-300 ml-4 space-y-2">
                <li>
                    We
                    <strong class="font-bold">never</strong>
                    receive or store your Google account password.
                </li>
                <li>
                    Authentication is handled directly by Google through a secure OAuth 2.0 flow.
                </li>
                <li>
                    You can review and manage the permissions you grant to our application at any
                    time in your
                    <a
                        href="https://myaccount.google.com/permissions"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="text-blue-600 dark:text-blue-400 hover:underline"
                        >Google Account settings</a
                    >.
                </li>
            </ul>
        </section>
    </div>
</template>
