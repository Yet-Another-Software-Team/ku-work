import { test, expect } from "@playwright/test";

// setup(() => {
//     // Any setup steps if necessary
//     const config = useRuntimeConfig();
//     const testCompanyEmail = config.public.testCompanyEmail;
// });

test("company signup google", async ({ page }) => {
    await page.goto("/"); // http://localhost:3000/

    await page.waitForLoadState("networkidle");

    // Navigate to company signup page
    await page.getByText("Company").click();
    await page.getByRole("link", { name: "Sign Up" }).click();

    // First step - fill out the form
    await page.getByRole("textbox", { name: "Enter your Company Name (" }).fill("TestCompany");
    await page
        .getByRole("textbox", { name: "Enter your Company Email" })
        .fill("testCompany@gmail.com");
    await page
        .getByRole("textbox", { name: "Enter your Company Website" })
        .fill("https://test.web.com");
    await page
        .getByRole("textbox", { name: "Enter your Company Website" })
        .fill("https://test.web.com");
    await page.getByRole("textbox", { name: "Enter your Password" }).fill("password");
    await page.getByRole("textbox", { name: "Repeat your Password" }).fill("password");
    await page.getByRole("textbox", { name: "Enter International Phone" }).fill("+66 1112");
    await page.getByRole("textbox", { name: "Enter your Address" }).fill("Test Address");
    await page.getByRole("textbox", { name: "Enter your City" }).fill("Test City");
    await page.getByRole("textbox", { name: "Enter your Country" }).fill("Test Country");
    await page.getByRole("button", { name: "Next" }).click();

    // Second step - upload images
    await page.getByRole("button", { name: "+" }).first().click();
    await page.getByRole("button", { name: "+" }).first().setInputFiles("test-avatar.png");
    await page.getByRole("button", { name: "+" }).click();
    await page.getByRole("button", { name: "+" }).setInputFiles("test-business.png");
    await page.getByRole("button", { name: "Submit" }).click();
    await page.getByRole("button", { name: "Done" }).click();

    await expect(page).toHaveURL(/dashboard/);
});

test("company login", async ({ page }) => {
    await page.goto("/");

    await page.waitForLoadState("networkidle");
});

test("student signup google", async ({ page }) => {
    await page.goto("/");

    await page.waitForLoadState("networkidle");

    // Wait for the popup and sign in
    const [popup] = await Promise.all([
        page.waitForEvent("popup"),
        page.getByRole("button", { name: "Continue with Google" }).click(),
    ]);

    // Login via Google Auth Page
    await popup.waitForLoadState("domcontentloaded");

    const insertEmailText = await popup.getByRole("textbox", { name: "Email or phone" });
    await insertEmailText.fill("WhenWillThisWork" + "@gmail.com");
    await insertEmailText.press("Enter");

    const insertPasswordText = await popup.getByRole("textbox", { name: "Enter your password" });
    await insertPasswordText.fill("ThisIsThePassword");
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
