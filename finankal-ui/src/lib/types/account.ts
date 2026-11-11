export interface AccountDto {
    id: string;
    name: string;
    type: string;
    balance: number;
    creditLimit?: number;
    billingDay?: number;
    statementShift?: string;
    createdAt: string;
}
