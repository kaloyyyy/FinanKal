export interface CreateLedgerEntryDto {

    accountId: string;

    amount: number;

    type: "DEBIT" | "CREDIT";
}

export interface CreateTransactionRequestDto {

    description: string;

    /** yyyy-MM-dd; omitted = now */
    date?: string;

    entries: CreateLedgerEntryDto[];
}