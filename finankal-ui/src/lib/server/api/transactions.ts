const API_URL = "http://localhost:8080";

export async function createTransaction(payload: {
    description: string;
    entries: {
        accountId: string;
        amount: number;
        type: "DEBIT" | "CREDIT";
    }[];
}) {
    const response = await fetch(
        `${API_URL}/api/transactions`,
        {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(payload)
        }
    );

    if (!response.ok) {
        const errorText = await response.text();

        throw new Error(
            `Spring API returned ${response.status}: ${errorText}`
        );
    }

    return response.json();
}