export interface CreateLedgerEntryDto {

    accountId: string;

    amount: number;

    type: "DEBIT" | "CREDIT";
}

export interface CreateTransactionRequestDto {

    description: string;

    entries: CreateLedgerEntryDto[];
}