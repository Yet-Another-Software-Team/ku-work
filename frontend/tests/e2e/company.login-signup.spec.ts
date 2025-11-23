import { test, expect } from "@playwright/test";

test("company signup mail", async ({ page }) => {
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
