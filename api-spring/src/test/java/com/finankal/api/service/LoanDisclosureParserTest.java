package com.finankal.api.service;

import org.junit.jupiter.api.Test;

import java.math.BigDecimal;
import java.time.LocalDate;

import static org.junit.jupiter.api.Assertions.*;

class LoanDisclosureParserTest {

    // A Maya Personal Loan disclosure as StatementPdfText extracts it, cut to a 3-month schedule. Labels and
    // values are separated by a space and a non-breaking space ( ), as in the real PDF.
    private static final String MAYA = """
            Maya Bank, Inc.
            DISCLOSURE STATEMENT ON LOAN/CREDIT TRANSACTION
            ACCOUNT NUMBER  663000000001
            LOAN BOOKING DATE  Apr 01, 2026
            ACCOUNT TYPE  Maya Personal Loan
            1. LOAN AMOUNT  P10,000
            a. Documentary Stamp Tax (DST) on Amount to Receive  P _75_
            3. AMOUNT TO RECEIVE (Loan Amount less Fees and Taxes/Deductions Collected)  P9,925
            a. First Due Date  May 01, 2026
            b. Next Due Date  30 days after your previous due date
            Installment Principal Interest Other Bank Cash Flows Outstanding
            0 0.00 10,000 10,000
            1 3000.00 399.00 (3399.00) 7000.00
            2 3110.39 288.61 (3399.00) 3889.61
            3 3889.61 155.19 (4044.80) 0.00
            """;

    @Test
    void readsTheLoanAndItsMonthlyDeadlines() {
        assertTrue(LoanDisclosureParser.isDisclosure(MAYA));
        LoanDisclosureParser.Loan loan = LoanDisclosureParser.parse(MAYA);

        assertEquals("Maya Personal Loan", loan.accountName());
        assertEquals("663000000001", loan.accountNumber());
        assertEquals(LocalDate.of(2026, 4, 1), loan.bookedOn());
        assertEquals(new BigDecimal("10000"), loan.amount());
        assertEquals(new BigDecimal("75"), loan.deducted());
        // 399 = 10,000 × rate × 30 days
        assertEquals(new BigDecimal("0.00133"), loan.dailyRate());
        assertEquals(3, loan.installments().size());
        assertEquals(LocalDate.of(2026, 6, 1), loan.installments().get(1).dueDate());
        assertEquals(new BigDecimal("4044.80"), loan.installments().get(2).payment());
    }

    @Test
    void refusesAScheduleThatDoesntAddUpToTheLoan() {
        String misread = MAYA.replace("2 3110.39 288.61 (3399.00) 3889.61\n", "");
        assertThrows(IllegalArgumentException.class, () -> LoanDisclosureParser.parse(misread));
    }
}
