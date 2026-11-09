package com.finankal.api.service;

import java.math.BigDecimal;
import java.math.MathContext;
import java.time.LocalDate;
import java.time.format.DateTimeFormatter;
import java.time.format.DateTimeFormatterBuilder;
import java.time.temporal.ChronoUnit;
import java.util.ArrayList;
import java.util.List;
import java.util.Locale;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Reads a lender's loan disclosure statement (Truth in Lending Act form) with its amortization schedule.
 * shortcut: only Maya Personal Loan's layout is known; add other lenders' wording when samples turn up.
 */
public final class LoanDisclosureParser {

    public record Installment(LocalDate dueDate, BigDecimal principal, BigDecimal interest, BigDecimal payment) {}

    /** dailyRate is the interest per day on what's owed, worked out from the first installment. */
    public record Loan(String lender, String accountName, String accountNumber, LocalDate bookedOn, BigDecimal amount,
                       BigDecimal deducted, BigDecimal dailyRate, List<Installment> installments) {}

    private static final Pattern TITLE = Pattern.compile("(?i)disclosure statement on loan");
    private static final Pattern MAYA = Pattern.compile("(?i)maya bank");
    private static final Pattern ACCOUNT_TYPE = Pattern.compile("(?i)account type\\s+(.+)");
    private static final Pattern ACCOUNT_NUMBER = Pattern.compile("(?i)account number\\s+(\\d{6,34})");
    private static final Pattern BOOKED = Pattern.compile("(?i)loan booking date\\s+([A-Za-z]{3} \\d{1,2}, \\d{4})");
    private static final Pattern FIRST_DUE = Pattern.compile("(?i)first due date\\s+([A-Za-z]{3} \\d{1,2}, \\d{4})");
    private static final Pattern AMOUNT = Pattern.compile("(?i)loan amount\\s+P\\s*([\\d,]+(?:\\.\\d{2})?)");
    private static final Pattern RECEIVED = Pattern.compile("(?i)amount to receive[^\\n]*?P\\s*([\\d,]+(?:\\.\\d{2})?)\\s*$", Pattern.MULTILINE);
    /** Amortization row: "1 3074.36 4788.00 (7862.36) 116925.64". Other bank charges, when printed, sit before the cash flow. */
    private static final Pattern ROW = Pattern.compile(
            "^(\\d{1,3}) ([\\d,]+\\.\\d{2}) ([\\d,]+\\.\\d{2})(?: [\\d,]+\\.\\d{2})? \\(([\\d,]+\\.\\d{2})\\) ([\\d,]+\\.\\d{2})[ \\t\\r]*$", Pattern.MULTILINE);

    private static final DateTimeFormatter DATE = new DateTimeFormatterBuilder().parseCaseInsensitive()
            .appendPattern("MMM d, yyyy").toFormatter(Locale.ENGLISH);

    private LoanDisclosureParser() {
    }

    /** Whether the text is a loan disclosure statement rather than a bank statement. */
    public static boolean isDisclosure(String text) {
        return TITLE.matcher(text).find();
    }

    public static Loan parse(String text) {
        // Maya's PDF separates each label from its value with a non-breaking space, which \s doesn't match.
        text = text.replace(' ', ' ');
        if (!MAYA.matcher(text).find()) {
            throw new IllegalArgumentException("Only Maya Personal Loan disclosure statements can be imported so far.");
        }
        BigDecimal amount = money(required(AMOUNT, text, "loan amount"));
        BigDecimal received = money(required(RECEIVED, text, "amount to receive"));
        LocalDate booked = LocalDate.parse(required(BOOKED, text, "loan booking date"), DATE);
        // The PDF says "30 days after your previous due date", but its interest is counted to the same day of
        // each month (30- and 31-day periods), so the deadlines fall on that day every month.
        LocalDate firstDue = LocalDate.parse(required(FIRST_DUE, text, "first due date"), DATE);

        List<Installment> installments = new ArrayList<>();
        Matcher row = ROW.matcher(text);
        while (row.find()) {
            int n = Integer.parseInt(row.group(1));
            if (n != installments.size() + 1) {
                throw new IllegalArgumentException("The amortization schedule skips installment " + (installments.size() + 1) + ".");
            }
            installments.add(new Installment(firstDue.plusMonths(n - 1L), money(row.group(2)), money(row.group(3)), money(row.group(4))));
        }
        if (installments.isEmpty()) {
            throw new IllegalArgumentException("No amortization schedule was found in this disclosure statement.");
        }
        BigDecimal principal = installments.stream().map(Installment::principal).reduce(BigDecimal.ZERO, BigDecimal::add);
        if (principal.compareTo(amount) != 0) {
            throw new IllegalArgumentException("The schedule's principal adds up to ₱" + principal.toPlainString()
                    + ", not the loan amount ₱" + amount.toPlainString() + ". The PDF may have been misread.");
        }

        // First period: interest = amount × rate × days from booking to the first due date.
        long days = ChronoUnit.DAYS.between(booked, firstDue);
        BigDecimal dailyRate = installments.get(0).interest()
                .divide(amount.multiply(BigDecimal.valueOf(days)), new MathContext(6))
                .stripTrailingZeros();

        Matcher type = ACCOUNT_TYPE.matcher(text);
        String accountName = type.find() ? type.group(1).trim() : "Maya Personal Loan";
        return new Loan("Maya", accountName, required(ACCOUNT_NUMBER, text, "account number"), booked, amount,
                amount.subtract(received), dailyRate, installments);
    }

    private static String required(Pattern pattern, String text, String what) {
        Matcher m = pattern.matcher(text);
        if (!m.find()) {
            throw new IllegalArgumentException("Couldn't find the " + what + " in this disclosure statement.");
        }
        return m.group(1);
    }

    private static BigDecimal money(String value) {
        return new BigDecimal(value.replace(",", ""));
    }
}
