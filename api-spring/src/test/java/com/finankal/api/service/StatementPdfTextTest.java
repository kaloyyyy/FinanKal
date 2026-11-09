package com.finankal.api.service;

import org.apache.pdfbox.pdmodel.PDDocument;
import org.apache.pdfbox.pdmodel.PDPage;
import org.apache.pdfbox.pdmodel.PDPageContentStream;
import org.apache.pdfbox.pdmodel.font.PDType1Font;
import org.apache.pdfbox.pdmodel.font.Standard14Fonts;
import org.junit.jupiter.api.Test;

import java.io.ByteArrayOutputStream;
import java.io.IOException;

import static org.junit.jupiter.api.Assertions.*;

class StatementPdfTextTest {

    /** Same column geometry as a MariBank statement: right-aligned amounts under OUTGOING / INCOMING. */
    @Test
    void amountsAreTaggedByTheColumnTheySitIn() throws IOException {
        try (PDDocument doc = new PDDocument()) {
            PDPage page = new PDPage();
            doc.addPage(page);
            PDType1Font font = new PDType1Font(Standard14Fonts.FontName.HELVETICA);

            try (PDPageContentStream cs = new PDPageContentStream(doc, page)) {
                text(cs, font, 50, 700, "DATE");
                text(cs, font, 364, 700, "OUTGOING (PHP)");
                text(cs, font, 485, 700, "INCOMING (PHP)");
                text(cs, font, 50, 680, "01 OCT");
                text(cs, font, 397, 680, "250.00");
                text(cs, font, 50, 660, "02 OCT");
                text(cs, font, 510, 660, "1,000.00");
            }

            ByteArrayOutputStream out = new ByteArrayOutputStream();
            doc.save(out);
            String text = StatementPdfText.extract(out.toByteArray(), null);

            assertTrue(text.contains("250.00DR"), text);
            assertTrue(text.contains("1,000.00CR"), text);
            assertFalse(text.contains("(PHP)DR") || text.contains("(PHP)CR"), text);
        }
    }

    @Test
    void notAPdfIsAValidationError() {
        assertThrows(IllegalArgumentException.class, () -> StatementPdfText.extract("hello".getBytes(), null));
    }

    private static void text(PDPageContentStream cs, PDType1Font font, float x, float y, String s) throws IOException {
        cs.beginText();
        cs.setFont(font, 10);
        cs.newLineAtOffset(x, y);
        cs.showText(s);
        cs.endText();
    }
}
