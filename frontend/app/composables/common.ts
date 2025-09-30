export const logout = async () => {
    try {
        await $fetch("/api/logout");
        localStorage.removeItem("token");
        localStorage.removeItem("username");
        localStorage.removeItem("role");
        localStorage.removeItem("isRegistered");
        navigateTo("/", { replace: true });
    } catch (error) {
        console.error("Error during logout:", error);
    } finally {
        navigateTo("/", { replace: true });
    }
};
