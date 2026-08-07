export interface LedgerEntryDto {

    accountId: string;

    amount: number;

    type: "DEBIT" | "CREDIT";

    description?: string;

    transactionId?: string;

    createdAt?: string;
}