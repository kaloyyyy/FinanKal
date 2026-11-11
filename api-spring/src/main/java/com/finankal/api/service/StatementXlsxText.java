package com.finankal.api.service;

import org.apache.poi.EncryptedDocumentException;
import org.apache.poi.ss.usermodel.Cell;
import org.apache.poi.ss.usermodel.CellType;
import org.apache.poi.ss.usermodel.DataFormatter;
import org.apache.poi.ss.usermodel.DateUtil;
import org.apache.poi.ss.usermodel.Row;
import org.apache.poi.ss.usermodel.Sheet;
import org.apache.poi.ss.usermodel.Workbook;
import org.apache.poi.ss.usermodel.WorkbookFactory;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.math.BigDecimal;
import java.math.RoundingMode;
import java.util.HashMap;
import java.util.Locale;
import java.util.Map;
import java.util.regex.Pattern;

/**
 * Turns a statement spreadsheet (UnionBank's "Transaction History" .xlsx) into the text BankStatementParser
 * reads: one line per row, with amounts under a DEBIT / CREDIT header tagged DR / CR like StatementPdfText does.
 */
public final class StatementXlsxText {

    private static final Pattern OUT_HEADER = Pattern.compile("(?i)OUTGOING|DEBITS?");
    private static final Pattern IN_HEADER = Pattern.compile("(?i)INCOMING|CREDITS?");

    private StatementXlsxText() {
    }

    public static String extract(byte[] xlsx, String password) {
        try (Workbook workbook = WorkbookFactory.create(new ByteArrayInputStream(xlsx), password)) {
            StringBuilder text = new StringBuilder();
            DataFormatter formatter = new DataFormatter(Locale.US);
            for (Sheet sheet : workbook) {
                Map<Integer, String> direction = new HashMap<>(); // column → DR / CR
                for (Row row : sheet) {
                    StringBuilder line = new StringBuilder();
                    for (Cell cell : row) {
                        String value;
                        if (cell.getCellType() != CellType.NUMERIC) {
                            value = unwrap(formatter.formatCellValue(cell)).trim();
                            if (OUT_HEADER.matcher(value).matches()) {
                                direction.put(cell.getColumnIndex(), "DR");
                            } else if (IN_HEADER.matcher(value).matches()) {
                                direction.put(cell.getColumnIndex(), "CR");
                            }
                        } else if (DateUtil.isCellDateFormatted(cell)) {
                            value = cell.getLocalDateTimeCellValue().toLocalDate().toString();
                        } else {
                            value = BigDecimal.valueOf(cell.getNumericCellValue()).setScale(2, RoundingMode.HALF_UP).toPlainString()
                                    + direction.getOrDefault(cell.getColumnIndex(), "");
                        }
                        if (!value.isEmpty()) {
                            line.append(value).append(' ');
                        }
                    }
                    if (!line.isEmpty()) {
                        text.append(line.toString().trim()).append('\n');
                    }
                }
            }
            return text.toString();
        } catch (EncryptedDocumentException e) {
            throw new IllegalArgumentException("Statement spreadsheet is password-protected: wrong or missing password");
        } catch (IOException | RuntimeException e) {
            throw new IllegalArgumentException("File is not a readable spreadsheet statement");
        }
    }

    /**
     * UnionBank hard-wraps descriptions inside the cell, mid-word, at 20 characters ("OFF-US ATM WITHDRAWA\nL";
     * 20 after the "To account: " label). Those breaks are removed; the other line ends become spaces.
     * shortcut: a real line of exactly that width is glued to the next one, the file can't tell them apart.
     */
    private static String unwrap(String cell) {
        StringBuilder out = new StringBuilder();
        for (String part : cell.split("\n")) {
            out.append(part);
            if (part.length() != (part.startsWith("To account: ") ? 32 : 20)) {
                out.append(' ');
            }
        }
        return out.toString();
    }
}
