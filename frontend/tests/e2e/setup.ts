import fs from "fs";

function generateRandomString(length: number): string {
    const characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    let result = "";

    for (let i = 0; i < length; i++) {
        const randomIndex = Math.floor(Math.random() * characters.length);
        result += characters.charAt(randomIndex);
    }

    return result;
}

const csvFile = "frontend/tests/e2e/company-mail.csv";
let uniqueEmail = `testcompany+${generateRandomString(8)}@kuwork.com`;
const uniquePassword = `TestPassword!${generateRandomString(5)}`;
const newLine = `${uniqueEmail},${uniquePassword}\n`;

// Read CSV as text
const content = fs.existsSync(csvFile) ? fs.readFileSync(csvFile, "utf-8") : "";

// Check if line already exists
while (true) {
    if (!content.includes(uniqueEmail)) {
        fs.appendFileSync(csvFile, newLine);
        break;
    } else {
        console.log("Duplicate line. Not added.");
        uniqueEmail = `testcompany+${generateRandomString(8)}@kuwork.com`;
    }
}
