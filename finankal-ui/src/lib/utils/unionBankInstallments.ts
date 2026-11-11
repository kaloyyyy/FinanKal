export type UnionBankInstallmentRow = { number: number; principal: number; interest: number };
export type UnionBankInstallmentPlan = {
    id: string;
    cardId: string;
    name: string;
    reference: string;
    balance: number | null;
    monthlyInstallment: number;
    term: number;
    processingFee: number | null;
    rows: UnionBankInstallmentRow[];
    nextNumber: number;
    nextDueDate: string;
};
