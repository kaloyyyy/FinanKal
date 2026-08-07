const API_URL = "http://localhost:8080";

export async function getUserFinancialSummary(
    userId: string
) {
    const [
        totalCreditResponse,
        totalDebitResponse,
        netWorthResponse
    ] = await Promise.all([
        fetch(
            `${API_URL}/api/users/${userId}/total-credit`
        ),

        fetch(
            `${API_URL}/api/users/${userId}/total-debit`
        ),

        fetch(
            `${API_URL}/api/users/${userId}/net-worth`
        )
    ]);


    if (
        !totalCreditResponse.ok ||
        !totalDebitResponse.ok ||
        !netWorthResponse.ok
    ) {
        throw new Error(
            "Failed to fetch financial summary"
        );
    }


    const [
        totalCredit,
        totalDebit,
        netWorth
    ] = await Promise.all([
        totalCreditResponse.json(),
        totalDebitResponse.json(),
        netWorthResponse.json()
    ]);

    console.log(totalCredit,totalDebit, netWorth);
    return {
        // Money received / assets side
        totalCredit,

        // Money owed / liabilities side
        totalDebit,

        // Assets - liabilities
        netWorth
    };
}