import { test, expect } from "@playwright/test";

test("student signup google", async ({ page }) => {
    await page.goto("/"); // goes to http://localhost:3000/

    await page.waitForLoadState("networkidle");

    // Wait for the popup and sign in
    const [popup] = await Promise.all([
        page.waitForEvent("popup"),
        page.getByRole("button", { name: "Continue with Google" }).click(),
    ]);

    // Login via Google Auth Page
    await popup.waitForLoadState("domcontentloaded");

    const insertEmailText = await popup.getByRole("textbox", { name: "Email or phone" });
    await insertEmailText.fill(process.env.TEST_STUDENT_EMAIL);
    await insertEmailText.press("Enter");

    const insertPasswordText = await popup.getByRole("textbox", { name: "Enter your password" });
    await insertPasswordText.fill(process.env.TEST_STUDENT_PASSWORD);
    await insertPasswordText.press("Enter");

    await popup.getByRole("button", { name: "Continue" }).click();
    await popup.waitForEvent("close");

    // Continue registration on main page
    await page.waitForURL(/register\/student/);

    // Upload profile picture
    const addButton = await page.getByRole("button", { name: "+" });
    await addButton.setInputFiles("student.png");

    // Fill out the registration form
    await page.getByRole("button", { name: "Next" }).click();
    await page.getByRole("textbox", { name: /student id/i }).fill("6610101010");
    await page.getByRole("combobox").first().selectOption("Software and Knowledge Engineering");
    await page.getByRole("combobox").nth(1).selectOption("Graduated");

    // Upload document
    const uploadButton = await page.getByRole("button", { name: "Upload Document" });
    await uploadButton.setInputFiles("student.png");

    // Submit the form
    await page.getByRole("button", { name: "Submit" }).click();
    await page.getByRole("button", { name: "Done" }).click();
    await expect(page).toHaveURL(/register\/student/);
});
