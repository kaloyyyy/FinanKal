package com.finankal.api.service;

import com.finankal.api.service.BankStatementParser.Bank;
import com.finankal.api.service.BankStatementParser.Line;
import com.finankal.api.service.BankStatementParser.Result;
import org.junit.jupiter.api.Test;

import java.math.BigDecimal;
import java.time.LocalDate;

import static org.junit.jupiter.api.Assertions.*;

class BankStatementParserTest {

    @Test
    void balanceColumnDecidesDirection() {
        Result r = BankStatementParser.parse("""
                Transaction History 2026
                Date and Time Description Reference No. Debit Credit Balance
                Beginning Balance PHP 5,000.00
                2026-01-15 10:23 AM Payment to Jollibee 1234567890 250.00 4,750.00
                2026-01-16 09:00 AM Received from Juan 9876543210 1,000.00 5,750.00
                Ending Balance 5,750.00
                """, Bank.GCASH);

        assertEquals(2, r.lines().size());
        Line spend = r.lines().get(0);
        assertEquals(LocalDate.of(2026, 1, 15), spend.date());
        assertEquals("Payment to Jollibee 1234567890", spend.description());
        assertEquals(new BigDecimal("250.00"), spend.amount());
        assertTrue(spend.debit());
        assertFalse(r.lines().get(1).debit());
        assertTrue(r.skipped().isEmpty());
        assertEquals(new BigDecimal("5000.00"), r.openingBalance());
        assertEquals(new BigDecimal("5750.00"), r.closingBalance());
    }

    @Test
    void explicitSignsAndMarkers() {
        Result r = BankStatementParser.parse("""
                Statement period 2026
                15 Jan 2026 Grab ride -₱180.00
                Jan 20 ATM WITHDRAWAL 2,000.00 DR
                01/22/26 SALARY 30,000.00CR
                """, Bank.BPI);

        assertEquals(3, r.lines().size());
        assertTrue(r.lines().get(0).debit());
        assertEquals(new BigDecimal("180.00"), r.lines().get(0).amount());
        assertEquals(LocalDate.of(2026, 1, 20), r.lines().get(1).date());
        assertTrue(r.lines().get(1).debit());
        assertFalse(r.lines().get(2).debit());
    }

    @Test
    void unknownDirectionIsSkippedNotGuessed() {
        Result r = BankStatementParser.parse("""
                01/22/2026 Something 99.00
                Ref 12 not a date 10.00
                """, Bank.BDO);

        assertTrue(r.lines().isEmpty());
        assertEquals(1, r.skipped().size());
    }

    /** Text as StatementPdfText produces it for a MariBank savings statement (names made up). */
    @Test
    void mariBankSavingsStatement() {
        Result r = BankStatementParser.parse("""
                BANK STATEMENT
                01 JAN 2027
                ACCOUNT SUMMARY
                01 DEC 2026 to 31 DEC 2026
                ACCOUNT STARTING BALANCE TOTAL OUTGOING TOTAL INCOMING ENDING BALANCE
                SAVINGS 5,000.00 350.00 1,000.00 5,650.00
                SAVINGS - TRANSACTION DETAILS
                DATE TRANSACTION OUTGOING (PHP) INCOMING (PHP)
                Interest
                DEC 1.50CR
                Net Interest
                Juan Dela Cruz
                01 DEC 250.00DR
                Transfer
                Maria Clara Santos
                02 DEC Reyes 1,000.00CR
                Transfer
                GoEXTRA99
                15 DEC 98.01DR
                Top Up - Data
                page 1 of 2
                BANK STATEMENT
                01 JAN 2027
                DATE TRANSACTION OUTGOING (PHP) INCOMING (PHP)
                Overseas ATM Cash
                31 DEC Withdrawal 1.99DR
                Cash Withdrawal
                SAVINGS - INTEREST & TAX DETAILS*
                01 DEC 5,000.00 0.45 -0.09 0.36
                """, Bank.MARIBANK);

        assertEquals(5, r.lines().size());
        // Interest has a month but no day; it lands on the month's last day.
        assertEquals(new Line(LocalDate.of(2026, 12, 31), "Interest (Net Interest)", new BigDecimal("1.50"), false), r.lines().get(0));
        assertEquals(new Line(LocalDate.of(2026, 12, 1), "Juan Dela Cruz (Transfer)", new BigDecimal("250.00"), true), r.lines().get(1));
        assertEquals(new Line(LocalDate.of(2026, 12, 2), "Maria Clara Santos Reyes (Transfer)", new BigDecimal("1000.00"), false), r.lines().get(2));
        assertEquals(new Line(LocalDate.of(2026, 12, 15), "GoEXTRA99 (Top Up - Data)", new BigDecimal("98.01"), true), r.lines().get(3));
        assertEquals(new Line(LocalDate.of(2026, 12, 31), "Overseas ATM Cash Withdrawal (Cash Withdrawal)", new BigDecimal("1.99"), true), r.lines().get(4));
        assertTrue(r.skipped().isEmpty());
        assertEquals(new BigDecimal("5000.00"), r.openingBalance());
        assertEquals(new BigDecimal("5650.00"), r.closingBalance());
        assertEquals(LocalDate.of(2026, 12, 1), r.periodStart());
        assertEquals(LocalDate.of(2026, 12, 31), r.periodEnd());
    }

    /** MariBank's header (number made up); a later "GCash" in a transaction must not win. */
    @Test
    void bankAndAccountNumberComeFromTheHeader() {
        String text = """
                BANK STATEMENT
                01 NOV 2026
                JUAN DELA CRUZ Contact Us
                MARIBANK ACCOUNT: 12345678901 Call (+632) 8000 0000
                ACCOUNT SUMMARY
                ACCOUNT STARTING BALANCE TOTAL OUTGOING TOTAL INCOMING ENDING BALANCE
                SAVINGS 5,000.00 350.00 1,000.00 5,650.00
                Transfer to GCash
                01 OCT 250.00DR
                """;
        assertEquals(Bank.MARIBANK, BankStatementParser.detectBank(text));
        assertEquals("12345678901", BankStatementParser.accountNumber(text));

        assertEquals("1234567890", BankStatementParser.accountNumber("Account No.: 1234-5678-90"));
        assertNull(BankStatementParser.accountNumber("Account balances up to PHP 1,000,000"));
        assertNull(BankStatementParser.detectBank("Some other bank"));
    }

    /**
     * UnionBank transaction history as StatementPdfText produces it (names and refs made up): newest first,
     * Debit/Credit columns not tagged, so directions come from the balance; descriptions wrap below.
     */
    @Test
    void unionBankHistoryNewestFirst() {
        String text = """
                Page 1 of 1
                Juan Dela Cruz
                Account Number     1234 5678 9012
                TRANSACTION HISTORY AS OF NOVEMBER 09, 2026
                Date Check No. Ref. No. Description Debit Credit Balance
                09/29/26 UB100001 Sent to Juan Dela Cruz SEA PHP 19,010.34 PHP 0.00
                998095
                09/29/26 S20000001 Payroll PHP 18,495.43 PHP 19,010.34
                09/25/26 UB100002 Received from JUAN DELA CRUZ PHP 514.91 PHP 514.91
                FLORES BPI 919246
                For billing concerns, you may contact our 24-Hour Customer Service at +632 8841-8600 or send us your concern via our Mailbox-Support Feature.
                """;
        assertEquals(Bank.UNIONBANK, BankStatementParser.detectBank(text));
        assertEquals("123456789012", BankStatementParser.accountNumber(text));

        Result r = BankStatementParser.parse(text, Bank.UNIONBANK);
        assertEquals(3, r.lines().size());
        assertEquals(new Line(LocalDate.of(2026, 9, 29), "UB100001 Sent to Juan Dela Cruz SEA 998095", new BigDecimal("19010.34"), true), r.lines().get(0));
        assertFalse(r.lines().get(1).debit());
        // Oldest row: balance equals amount, so it was money in and the account started at zero.
        assertEquals(new Line(LocalDate.of(2026, 9, 25), "UB100002 Received from JUAN DELA CRUZ FLORES BPI 919246", new BigDecimal("514.91"), false), r.lines().get(2));
        assertTrue(r.skipped().isEmpty());
        assertEquals(new BigDecimal("0.00"), r.openingBalance());
        assertEquals(new BigDecimal("0.00"), r.closingBalance());
    }

    /** The oldest row empties the account: money in would mean it started below zero, so it's money out. */
    @Test
    void unionBankOldestRowThatEmptiesTheAccountIsMoneyOut() {
        Result r = BankStatementParser.parse("""
                Date Check No. Ref. No. Description Debit Credit Balance
                11/02/26 UB100003 Received from JUAN DELA CRUZ PHP 3,586.83 PHP 3,586.83
                11/02/26 UB100004 BILLS PAYMENT PHP 18,912.67 PHP 0.00
                """, Bank.UNIONBANK);

        assertTrue(r.skipped().isEmpty());
        assertEquals(new Line(LocalDate.of(2026, 11, 2), "UB100004 BILLS PAYMENT", new BigDecimal("18912.67"), true), r.lines().get(1));
        assertFalse(r.lines().get(0).debit());
        assertEquals(new BigDecimal("18912.67"), r.openingBalance());
    }

    /** Maya Wallet as StatementPdfText produces it (name, number and refs made up). */
    @Test
    void mayaWallet() {
        String text = """
                1 of 2
                JUAN DELA CRUZ Account number: +639170000000
                @juan Coverage period: 1 Sep 2026 - 30 Sep 2026
                Account summary
                Starting balance + Total incoming - Total outgoing Ending balance
                PHP 915.00 PHP 7,000.00 PHP 7,862.36 PHP 52.64
                Transaction details
                Date & Time Transaction type & details Reference ID Amount Running balance
                Beginning Balance PHP 915.00
                1 Sep 2026 Received from bank AAAA00000001 PHP 7,000.00 PHP 7,915.00
                11:08:52 PM Credit via InstaPay
                2 of 2
                JUAN DELA CRUZ Account number: +639170000000
                1 Sep 2026 Paid bill AAAA00000002 -PHP 7,862.36 PHP 52.64
                11:09:22 PM Pay with PayMaya Wallet Bills Pay
                Ending Balance PHP 52.64
                ** Nothing follows **
                """;
        assertEquals(Bank.MAYA, BankStatementParser.detectBank(text));
        assertEquals("639170000000", BankStatementParser.accountNumber(text));
        assertEquals("Maya Wallet", BankStatementParser.accountLabel(Bank.MAYA, "639170000000"));
        assertNull(BankStatementParser.unsupported(text, Bank.MAYA));

        Result r = BankStatementParser.parse(text, Bank.MAYA);
        assertEquals(new Line(LocalDate.of(2026, 9, 1), "Received from bank AAAA00000001 (Credit via InstaPay)", new BigDecimal("7000.00"), false), r.lines().get(0));
        assertEquals(new Line(LocalDate.of(2026, 9, 1), "Paid bill AAAA00000002 (Pay with PayMaya Wallet Bills Pay)", new BigDecimal("7862.36"), true), r.lines().get(1));
        assertEquals(2, r.lines().size());
        assertEquals(new BigDecimal("915.00"), r.openingBalance());
        assertEquals(new BigDecimal("52.64"), r.closingBalance());
        assertEquals(LocalDate.of(2026, 9, 30), r.periodEnd());
    }

    @Test
    void mayaCreditStatementsAreRefused() {
        assertNotNull(BankStatementParser.unsupported("Maya Easy Credit\nStatement of Account", Bank.MAYA));
        assertNotNull(BankStatementParser.unsupported("Maya Personal Loan\nStatement", Bank.MAYA));
        assertEquals("Maya Savings", BankStatementParser.accountLabel(Bank.MAYA, "1234567890"));
    }

    /** BDO "Account Activities" export as StatementPdfText produces it (numbers and refs made up). */
    @Test
    void bdoAccountActivities() {
        String text = """
                Dear customer, Exported date range
                Oct 22, 2025
                The transactions are listed below for the current account BBAN 001234567890 Oct 20, 2026
                Account number: 001234567890 Available balance PHP100.08
                Date Counterparty Description Amount
                Oct 20, 2026 Sent via InstaPay POB IBFT BN-20261020- -PHP44,000.00
                UBP ****0000 30895000 IBTW
                Oct 20, 2026 Received from IBTD 000004555282 PHP44,000.00
                Other Bank
                Page 1 of 2
                Account number 001234567890 Available balance PHP100.08
                Date Counterparty Description Amount
                Mar 4, 2026 W/D FR SAV BDO W/D FR SAV BDO/008332 -PHP25,000.00
                /008332 33RVERBNKSFL BDO
                33RVERBNKSFL
                BDO
                Mar 4, 2026 W/D FR SAV BDO W/D FR SAV BDO/008332 -PHP25,000.00
                Page 2 of 2
                """;
        assertEquals(Bank.BDO, BankStatementParser.detectBank(text));
        assertEquals("001234567890", BankStatementParser.accountNumber(text, Bank.BDO));

        Result r = BankStatementParser.parse(text, Bank.BDO);
        assertEquals(4, r.lines().size());
        assertEquals(new Line(LocalDate.of(2026, 10, 20), "Sent via InstaPay POB IBFT BN-20261020- UBP ****0000 30895000 IBTW",
                new BigDecimal("44000.00"), true), r.lines().get(0));
        assertEquals(new Line(LocalDate.of(2026, 10, 20), "Received from IBTD 000004555282 Other Bank",
                new BigDecimal("44000.00"), false), r.lines().get(1));
        assertEquals("Cash", BankStatementParser.category(Bank.BDO, r.lines().get(2)));
        assertEquals(r.lines().get(2), r.lines().get(3));
        assertTrue(r.skipped().isEmpty());
    }

    /** GCash lists failed loads as debits and doesn't list their reversals; the running balance is followed. */
    @Test
    void gcashFollowsRunningBalance() {
        String text = """
                GCash Transaction History
                2026-01-16 to 2026-01-16
                Date and Time Description Reference No. Debit Credit Balance
                STARTING BALANCE 2107.80
                2026-01-16 04:30 PM Buy Load Transaction for 09170000001 0000970170394 11.00DR 2096.80
                2026-01-16 05:12 PM Buy Load Transaction for 09170000002 0000970328603 11.00DR 2085.80
                2026-01-16 05:13 PM Buy Load Transaction for 09170000002 0000970317473 11.00DR 2085.80
                2026-01-16 05:19 PM Buy Load Transaction for 09170000002 0000970379561 12.00DR 2084.80
                ENDING BALANCE 2084.80
                """;
        Result r = BankStatementParser.parse(text, Bank.GCASH);

        LocalDate day = LocalDate.of(2026, 1, 16);
        assertEquals(1, r.skipped().size()); // 05:13: the balance didn't move
        assertTrue(r.skipped().get(0).contains("0000970317473"));
        assertEquals(4, r.lines().size());
        assertEquals(new Line(day, "Buy Load Transaction for 09170000002 0000970379561", new BigDecimal("12.00"), true), r.lines().get(2));
        assertEquals(new Line(day, BankStatementParser.GCASH_CORRECTION, new BigDecimal("11.00"), false), r.lines().get(3));

        // The rows now add up to what GCash says it holds.
        BigDecimal end = r.openingBalance();
        for (Line line : r.lines()) {
            end = line.debit() ? end.subtract(line.amount()) : end.add(line.amount());
        }
        assertEquals(r.closingBalance(), end);
    }

    /** GCash transaction history as StatementPdfText produces it (numbers and refs made up). */
    @Test
    void gcashHistory() {
        String text = """
                GCash Transaction History
                2026-09-10 to 2026-11-08
                Date and Time Description Reference No. Debit Credit Balance
                STARTING BALANCE 1253.56
                2026-09-10 07:46 AM Transfer from 09170000000 to 09180000000 9044879765879 250.00DR 1003.56
                2026-09-23 12:21 AM Received GCash from MariBank with account ending in 0000 and
                invno:20260923LAUIPHM2XXXB000000000000001 9045339782233 150.00CR 1153.56
                2026-09-23 12:21 AM Payment to Angkas 0000911838547 133.00DR 1020.56
                ENDING BALANCE 1020.56
                Total Debit 383.00DR
                Total Credit 150.00CR
                """;
        assertEquals(Bank.GCASH, BankStatementParser.detectBank(text));
        assertNull(BankStatementParser.accountNumber(text, Bank.GCASH));

        Result r = BankStatementParser.parse(text, Bank.GCASH);
        assertEquals(3, r.lines().size());
        assertEquals(new Line(LocalDate.of(2026, 9, 23),
                "Received GCash from MariBank with account ending in 0000 and invno:20260923LAUIPHM2XXXB000000000000001 9045339782233",
                new BigDecimal("150.00"), false), r.lines().get(1));
        assertEquals(new Line(LocalDate.of(2026, 9, 23), "Payment to Angkas 0000911838547", new BigDecimal("133.00"), true), r.lines().get(2));
        assertEquals(new BigDecimal("1253.56"), r.openingBalance());
        assertEquals(new BigDecimal("1020.56"), r.closingBalance());
        assertEquals(LocalDate.of(2026, 9, 10), r.periodStart());
        assertEquals(LocalDate.of(2026, 11, 8), r.periodEnd());
        assertTrue(r.skipped().isEmpty());
    }

    /** Maya Savings as StatementPdfText produces it (name, number and refs made up). */
    @Test
    void mayaSavings() {
        String text = """
                Statement of Account
                Statement date: 2 January 2027
                Consumer Savings
                0.28 PHP
                1234 5678 0005 7,503.50 PHP
                - 5,000.28 PHP
                JUAN DELA CRUZ 3.08 PHP
                Transactions
                1-31 Dec 2026
                7 Dec, 04:35 PM Wallet Transfer 5,000.00 5,000.28
                Reference ID: 1e85d4205b1d
                7 Dec, 04:35 PM Transfer is successfully sent to BancNet - 4,990.28 10.00
                Reference ID: 2741f4f84b62
                7 Dec, 04:35 PM Fee Applied - 10.00 0.00
                23 Dec, 03:49 AM Interest Applied (at 3.0% p.a.) 0.19 0.19
                Page 1 of 2
                Maya Savings (0005) Statement date: 2 January 2027
                23 Dec, 03:49 AM Base Interest Withholding Tax - 0.04 0.15
                Maya Bank, Inc.
                """;
        assertEquals(Bank.MAYA, BankStatementParser.detectBank(text));
        assertEquals("123456780005", BankStatementParser.accountNumber(text, Bank.MAYA));
        assertEquals("Maya Savings", BankStatementParser.accountLabel(Bank.MAYA, "123456780005"));

        Result r = BankStatementParser.parse(text, Bank.MAYA);
        assertEquals(new Line(LocalDate.of(2026, 12, 7), "Wallet Transfer", new BigDecimal("5000.00"), false), r.lines().get(0));
        assertEquals(new Line(LocalDate.of(2026, 12, 7), "Transfer is successfully sent to BancNet", new BigDecimal("4990.28"), true), r.lines().get(1));
        assertEquals(new Line(LocalDate.of(2026, 12, 23), "Base Interest Withholding Tax", new BigDecimal("0.04"), true), r.lines().get(4));
        assertEquals(5, r.lines().size());
        assertEquals(new BigDecimal("0.28"), r.openingBalance());
        assertEquals(new BigDecimal("0.15"), r.closingBalance());
        assertEquals(LocalDate.of(2026, 12, 1), r.periodStart());
        assertEquals(LocalDate.of(2026, 12, 31), r.periodEnd());
    }

    @Test
    void unionBankPayrollIsSalary() {
        LocalDate day = LocalDate.of(2026, 10, 14);
        BigDecimal amount = new BigDecimal("19155.05");
        assertEquals("Salary Income", BankStatementParser.category(Bank.UNIONBANK, new Line(day, "S55093995 Dela Cruz, Juan", amount, false)));
        // Transfers in, ATM withdrawals and other banks keep the default.
        assertEquals("", BankStatementParser.category(Bank.UNIONBANK, new Line(day, "UB100001 Received from JUAN DELA CRUZ SEA 1", amount, false)));
        assertEquals("", BankStatementParser.category(Bank.UNIONBANK, new Line(day, "S57762549 OFF-US ATM WITHDRAWAL", amount, true)));
        assertEquals("", BankStatementParser.category(Bank.MARIBANK, new Line(day, "S55093995 Dela Cruz, Juan", amount, false)));
    }

    /** UnionBank credit card statement as StatementPdfText produces it (name, card and refs made up). */
    @Test
    void unionBankCreditCard() {
        String text = """
                Statement of Account Available Points*
                Statement Date: October 23, 2026 6,062 PTS
                Rewards Platinum Mastercard Credit Card available points.
                C A R D  N U M B E R Summary
                * * * *  * * * *  * * * *  9 8 7 6
                Previous Balance
                NAME PHP 1,000.00
                J U A N  D E L A  C R U Z
                Due Date Minimum Amount Due Total Amount Due
                November 09, 2026 PHP 100.00 PHP 2,550.00
                Transactions
                DATE DESCRIPTION AMOUNT
                Oct 23, 2026 Instalment Interest (06/12) [1000001] PHP 50.00
                Sep 28, 2026 PayDirect - Rent Fees - Landlord PHP 2,000.00
                Sep 28, 2026 PayDirect Processing Fee PHP 500.00
                Sep 25, 2026 Payment via UB Online UB100001 PHP -1,000.00
                """;
        assertEquals(Bank.UNIONBANK_CARD, BankStatementParser.detectBank(text));
        assertEquals("9876", BankStatementParser.accountNumber(text, Bank.UNIONBANK_CARD));

        Result r = BankStatementParser.parse(text, Bank.UNIONBANK_CARD);
        assertTrue(r.skipped().isEmpty());
        assertEquals(4, r.lines().size());
        // Charges are money out; the payment is money in.
        assertTrue(r.lines().get(0).debit());
        assertEquals(new Line(LocalDate.of(2026, 9, 25), "Payment via UB Online UB100001", new BigDecimal("1000.00"), false), r.lines().get(3));
        // Balances are owed, so negative; the period is the month up to the statement date.
        assertEquals(new BigDecimal("-1000.00"), r.openingBalance());
        assertEquals(new BigDecimal("-2550.00"), r.closingBalance());
        assertEquals(LocalDate.of(2026, 9, 24), r.periodStart());
        assertEquals(LocalDate.of(2026, 10, 23), r.periodEnd());

        BankStatementParser.CardSummary card = BankStatementParser.cardSummary(text);
        assertEquals(LocalDate.of(2026, 11, 9), card.dueDate());
        assertEquals(new BigDecimal("100.00"), card.minimumDue());

        // Interest and the bank's own fee are categorised; rent paid through the card is ordinary spending.
        assertEquals("Loan Interest", BankStatementParser.category(Bank.UNIONBANK_CARD, r.lines().get(0)));
        assertEquals("", BankStatementParser.category(Bank.UNIONBANK_CARD, r.lines().get(1)));
        assertEquals("Loan Fees", BankStatementParser.category(Bank.UNIONBANK_CARD, r.lines().get(2)));
    }

    /** UnionBank credit card transaction history: two dates per row, two cards, no balances (names and numbers made up). */
    @Test
    void unionBankCreditCardHistory() {
        String text = """
                Page 1 of 2
                Juan Dela Cruz
                Account Number  1000 2000 3000 400
                Principal Cardholder (Juan Dela Cruz)
                Card Number: 1234 56** **** 1111
                TRANSACTION HISTORY AS OF NOVEMBER 09, 2026
                Transaction Date Posting Date Transaction Details Amount
                11/07/26 11/07/26 KFC EDSA CENTRAL,HIGHWAY HILLS,PHL PHP 295.00
                11/03/26 11/03/26 Payment via UB Online UB100001 -PHP 2,754.98
                For billing concerns, you may contact our 24-Hour Customer Service at +632 8841-8600 or send us your concern via our Mailbox-Support Feature.
                Page 2 of 2
                Virtual Card (Juan Dela Cruz)
                Card Number: 1234 56** **** 2222
                TRANSACTION HISTORY AS OF NOVEMBER 09, 2026
                Transaction Date Posting Date Transaction Details Amount
                """;
        assertEquals(Bank.UNIONBANK_CARD, BankStatementParser.detectBank(text));
        // The virtual card is the number the monthly statement prints, so both land in one account.
        assertEquals("2222", BankStatementParser.accountNumber(text, Bank.UNIONBANK_CARD));

        Result r = BankStatementParser.parse(text, Bank.UNIONBANK_CARD);
        assertTrue(r.skipped().isEmpty());
        assertEquals(new Line(LocalDate.of(2026, 11, 7), "KFC EDSA CENTRAL,HIGHWAY HILLS,PHL", new BigDecimal("295.00"), true), r.lines().get(0));
        assertEquals(new Line(LocalDate.of(2026, 11, 3), "Payment via UB Online UB100001", new BigDecimal("2754.98"), false), r.lines().get(1));
        assertNull(r.openingBalance());
        assertNull(r.closingBalance());
    }

    @Test
    void atmWithdrawalsGoToCash() {
        LocalDate day = LocalDate.of(2026, 10, 20);
        BigDecimal amount = new BigDecimal("10000.00");
        assertEquals("Cash", BankStatementParser.category(Bank.UNIONBANK, new Line(day, "S57762240 OFF-US ATM WITHDRAWAL", amount, true)));
        assertEquals("Cash", BankStatementParser.category(Bank.MARIBANK, new Line(day, "Overseas ATM Cash Withdrawal (Cash Withdrawal)", amount, true)));
        assertEquals("Cash", BankStatementParser.category(Bank.MAYA, new Line(day, "Withdrawal from ARANETA CENTER 2", amount, true)));
        // Not withdrawals: money coming in, a card's cash advance, ordinary spending.
        assertEquals("", BankStatementParser.category(Bank.UNIONBANK, new Line(day, "S57762240 OFF-US ATM WITHDRAWAL", amount, false)));
        assertEquals("", BankStatementParser.category(Bank.UNIONBANK_CARD, new Line(day, "ATM CASH WITHDRAWAL", amount, true)));
        assertEquals("", BankStatementParser.category(Bank.MARIBANK, new Line(day, "Nica Ong (Transfer)", amount, true)));
    }

    @Test
    void periodCrossingNewYearDatesJanuaryRowsInTheEndYear() {
        Result r = BankStatementParser.parse("""
                15 DEC 2026 to 14 JAN 2027
                Juan Dela Cruz
                20 DEC 100.00DR
                Transfer
                Juan Dela Cruz
                03 JAN 100.00DR
                Transfer
                """, Bank.MARIBANK);

        assertEquals(LocalDate.of(2026, 12, 20), r.lines().get(0).date());
        assertEquals(LocalDate.of(2027, 1, 3), r.lines().get(1).date());
    }
}
