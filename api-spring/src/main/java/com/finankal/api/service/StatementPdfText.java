package com.finankal.api.service;

import org.apache.pdfbox.Loader;
import org.apache.pdfbox.pdmodel.PDDocument;
import org.apache.pdfbox.pdmodel.PDPage;
import org.apache.pdfbox.pdmodel.encryption.InvalidPasswordException;
import org.apache.pdfbox.text.PDFTextStripper;
import org.apache.pdfbox.text.TextPosition;

import java.io.IOException;
import java.util.List;
import java.util.regex.Pattern;

/** Extracts a statement PDF's text in reading order, with column-only directions made explicit. */
public final class StatementPdfText {

    private StatementPdfText() {
    }

    public static String extract(byte[] pdf, String password) {
        try (PDDocument document = Loader.loadPDF(pdf, password == null ? "" : password)) {
            PDFTextStripper stripper = new ColumnTaggingStripper();
            stripper.setSortByPosition(true);
            return stripper.getText(document);
        } catch (InvalidPasswordException e) {
            throw new IllegalArgumentException("Statement PDF is password-protected: wrong or missing password");
        } catch (IOException e) {
            throw new IllegalArgumentException("File is not a readable PDF statement");
        }
    }

    /**
     * Statements like MariBank's put outgoing and incoming amounts in separate columns with no sign,
     * so the direction only exists as a position on the page. This appends DR / CR to every amount
     * that sits under the page's OUTGOING / INCOMING (or DEBIT / CREDIT) header.
     */
    private static final class ColumnTaggingStripper extends PDFTextStripper {

        private static final Pattern OUT_HEADER = Pattern.compile("(?i)(OUTGOING|DEBITS?)( \\(PHP\\))?");
        private static final Pattern IN_HEADER = Pattern.compile("(?i)(INCOMING|CREDITS?)( \\(PHP\\))?");
        private static final Pattern AMOUNT = Pattern.compile("\\(?-?[\\d,]+\\.\\d{2}\\)?-?");

        private float[] outColumn; // {startX, endX, y}
        private float[] inColumn;
        private boolean columnsLocked;

        ColumnTaggingStripper() throws IOException {
        }

        @Override
        protected void startPage(PDPage page) throws IOException {
            outColumn = null;
            inColumn = null;
            columnsLocked = false;
            super.startPage(page);
        }

        @Override
        protected void writeString(String text, List<TextPosition> positions) throws IOException {
            if (positions.isEmpty()) {
                super.writeString(text, positions);
                return;
            }
            String word = text.trim();
            float[] span = span(positions);

            if (!columnsLocked) {
                if (OUT_HEADER.matcher(word).matches()) {
                    outColumn = span;
                } else if (IN_HEADER.matcher(word).matches()) {
                    inColumn = span;
                }
                // Only a header pair on one line counts, so a lone "Credit" in a description can't move a column.
                columnsLocked = outColumn != null && inColumn != null && Math.abs(outColumn[2] - inColumn[2]) < 2;
            } else if (AMOUNT.matcher(word).matches()) {
                if (overlaps(span, outColumn)) {
                    text = text + "DR";
                } else if (overlaps(span, inColumn)) {
                    text = text + "CR";
                }
            }
            super.writeString(text, positions);
        }

        private static float[] span(List<TextPosition> positions) {
            TextPosition first = positions.get(0);
            TextPosition last = positions.get(positions.size() - 1);
            return new float[]{first.getXDirAdj(), last.getXDirAdj() + last.getWidthDirAdj(), first.getYDirAdj()};
        }

        private static boolean overlaps(float[] amount, float[] column) {
            return amount[0] <= column[1] && amount[1] >= column[0];
        }
    }
}
