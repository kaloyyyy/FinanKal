package com.finankal.api.service;

import com.finankal.api.service.BankStatementParser.Bank;
import com.finankal.api.service.BankStatementParser.Line;
import com.finankal.api.service.BankStatementParser.Result;
import org.apache.poi.poifs.crypt.EncryptionInfo;
import org.apache.poi.poifs.crypt.EncryptionMode;
import org.apache.poi.poifs.crypt.Encryptor;
import org.apache.poi.poifs.filesystem.POIFSFileSystem;
import org.apache.poi.ss.usermodel.Row;
import org.apache.poi.ss.usermodel.Sheet;
import org.apache.poi.xssf.usermodel.XSSFWorkbook;
import org.junit.jupiter.api.Test;

import java.io.ByteArrayOutputStream;
import java.io.OutputStream;
import java.math.BigDecimal;
import java.time.LocalDate;

import static org.junit.jupiter.api.Assertions.*;

class StatementXlsxTextTest {

    /** Same layout as UnionBank's password-protected "Transaction History" export (name and numbers made up). */
    @Test
    void unionBankHistoryReadsLikeItsPdf() throws Exception {
        byte[] file;
        try (XSSFWorkbook workbook = new XSSFWorkbook(); POIFSFileSystem fs = new POIFSFileSystem()) {
            Sheet sheet = workbook.createSheet();
            row(sheet, 0, "ACCOUNT NAME", "Juan Dela Cruz");
            row(sheet, 1, "ACCOUNT NUMBER:", "123456789012");
            row(sheet, 3, "DATE", "CHECK NO", "REFERENCE NO", "DESCRIPTION", "CURRENCY", "DEBIT", "CREDIT", "ENDING BALANCE");
            row(sheet, 4, "01/29/2026", null, "S43669828", "DELACRUZPAYROLLJAN31\n2026 \n", "PHP", null, 16706.05, 18935.60);
            row(sheet, 5, "01/29/2026", null, "UB358005",
                    "ONLINE FUND TRANSFER \nTo account: Union Bank of the Ph\nilippines\n109000000000\n", "PHP", 285.0, null, 2229.55);
            row(sheet, 6, "01/25/2026", null, "S42045859", "OFF-US ATM WITHDRAWA\nL \n", "PHP", 500.0, null, 2514.55);

            Encryptor encryptor = new EncryptionInfo(EncryptionMode.agile).getEncryptor();
            encryptor.confirmPassword("123");
            try (OutputStream out = encryptor.getDataStream(fs)) {
                workbook.write(out);
            }
            ByteArrayOutputStream bytes = new ByteArrayOutputStream();
            fs.writeFilesystem(bytes);
            file = bytes.toByteArray();
        }

        assertThrows(IllegalArgumentException.class, () -> StatementXlsxText.extract(file, "wrong"));

        String text = StatementXlsxText.extract(file, "123");
        assertEquals(Bank.UNIONBANK, BankStatementParser.detectBank(text));
        assertEquals("123456789012", BankStatementParser.accountNumber(text, Bank.UNIONBANK));

        Result r = BankStatementParser.parse(text, Bank.UNIONBANK);
        assertTrue(r.skipped().isEmpty(), text);
        assertEquals(3, r.lines().size(), text);
        assertEquals(new Line(LocalDate.of(2026, 1, 29), "S43669828 DELACRUZPAYROLLJAN312026", new BigDecimal("16706.05"), false), r.lines().get(0));
        assertEquals(new Line(LocalDate.of(2026, 1, 29),
                "UB358005 ONLINE FUND TRANSFER To account: Union Bank of the Philippines 109000000000", new BigDecimal("285.00"), true), r.lines().get(1));
        assertEquals(new Line(LocalDate.of(2026, 1, 25), "S42045859 OFF-US ATM WITHDRAWAL", new BigDecimal("500.00"), true), r.lines().get(2));
        assertEquals(new BigDecimal("3014.55"), r.openingBalance());
        assertEquals(new BigDecimal("18935.60"), r.closingBalance());

        assertEquals("Salary Income", BankStatementParser.category(Bank.UNIONBANK, r.lines().get(0)));
        assertEquals("Cash", BankStatementParser.category(Bank.UNIONBANK, r.lines().get(2)));
    }

    /** UnionBank credit card statement .xlsx: labelled summary rows, then DATE / DESCRIPTION / CURRENCY / AMOUNT (made-up values). */
    @Test
    void unionBankCardSheetReadsLikeItsPdf() throws Exception {
        byte[] file;
        try (XSSFWorkbook workbook = new XSSFWorkbook(); ByteArrayOutputStream out = new ByteArrayOutputStream()) {
            Sheet sheet = workbook.createSheet();
            row(sheet, 1, null, "ACCOUNT NAME", null, "JUAN DELA CRUZ");
            row(sheet, 2, null, "CARD NUMBER", null, "5292471506951135");
            row(sheet, 3, null, "STATEMENT BALANCE", "PHP", 1500.0);
            row(sheet, 4, null, "MINIMUM AMOUNT DUE", "PHP", 500.0);
            row(sheet, 5, null, "PAYMENT DUE DATE", null, "Feb 09, 2026");
            row(sheet, 6, null, "STATEMENT DATE", null, "Jan 23, 2026");
            row(sheet, 7, null, "PREVIOUS BALANCE", "PHP", 0.0);
            row(sheet, 10, null, "TOTAL AMOUNT DUE", "PHP", 1500.0);
            row(sheet, 13, "DATE", "DESCRIPTION", "CURRENCY", "AMOUNT");
            row(sheet, 14, "01/16/2026", "Quasi-Cash Fee", "PHP", 25.75);
            row(sheet, 15, "01/15/2026", "Payment via UB Online UB379796", "PHP", -1000.0);
            row(sheet, 16, "01/14/2026", "SHOPEE PH, MANDALUYONG, PHL", "PHP", 2474.25);
            workbook.write(out);
            file = out.toByteArray();
        }

        String text = StatementXlsxText.extract(file, null);
        assertEquals(Bank.UNIONBANK_CARD, BankStatementParser.detectBank(text), text);
        assertEquals("1135", BankStatementParser.accountNumber(text, Bank.UNIONBANK_CARD));

        BankStatementParser.CardSummary card = BankStatementParser.cardSummary(text);
        assertEquals(LocalDate.of(2026, 1, 23), card.statementDate());
        assertEquals(LocalDate.of(2026, 2, 9), card.dueDate());
        assertEquals(new BigDecimal("1500.00"), card.totalDue());
        assertEquals(new BigDecimal("500.00"), card.minimumDue());

        Result r = BankStatementParser.parse(text, Bank.UNIONBANK_CARD);
        assertTrue(r.skipped().isEmpty(), text);
        assertEquals(3, r.lines().size(), text);
        assertEquals(new Line(LocalDate.of(2026, 1, 16), "Quasi-Cash Fee", new BigDecimal("25.75"), true), r.lines().get(0));
        assertFalse(r.lines().get(1).debit());
        assertEquals("Loan Fees", BankStatementParser.category(Bank.UNIONBANK_CARD, r.lines().get(0)));
    }

    private static void row(Sheet sheet, int index, Object... cells) {
        Row row = sheet.createRow(index);
        for (int i = 0; i < cells.length; i++) {
            if (cells[i] instanceof Double number) {
                row.createCell(i).setCellValue(number);
            } else if (cells[i] != null) {
                row.createCell(i).setCellValue((String) cells[i]);
            }
        }
    }
}
