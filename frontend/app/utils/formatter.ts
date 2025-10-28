export function formatSalary(salary: number): string {
    return new Intl.NumberFormat("en", { notation: "compact" }).format(salary);
}

export function formatJobType(type: string): string {
    const typeMap: Record<string, string> = {
        fulltime: "Full Time",
        parttime: "Part Time",
        contract: "Contract",
        casual: "Casual",
        internship: "Internship",
    };
    return typeMap[type.toLowerCase()] || type;
}

export function formatExperience(exp: string): string {
    const expMap: Record<string, string> = {
        newgrad: "New Grad",
        junior: "Junior",
        senior: "Senior",
        manager: "Manager",
        internship: "Internship",
    };
    return expMap[exp.toLowerCase()] || exp;
}
