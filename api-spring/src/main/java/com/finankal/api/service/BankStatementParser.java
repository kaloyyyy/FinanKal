package com.finankal.api.service;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.format.DateTimeFormatter;
import java.time.format.DateTimeFormatterBuilder;
import java.time.format.DateTimeParseException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Locale;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Turns the extracted text of a bank / e-wallet statement into dated entries.
 *
 * One parser covers every supported bank: a transaction row is any line that starts with a date
 * and ends with one or more amounts (amount, optionally followed by the running balance).
 * Direction comes from an explicit sign / DR / CR marker, else from whether the balance went down.
 * Rows that can't be classified are returned in {@code skipped} so nothing is silently guessed.
 *
 * Statements with separate Outgoing/Incoming (Debit/Credit) columns have no sign in the text; the
 * PDF extractor tags those amounts DR/CR from their column position before they reach this parser.
 */
public final class BankStatementParser {

    public enum Bank {
        MARIBANK("MariBank", "maribank|seabank"),
        // BDO's "Account Activities" export names the bank only in transactions; its header has "Exported date range".
        BDO("BDO", "bdo unibank|\\bbdo\\b|exported date range"),
        MAYA("Maya", "maya bank|maya savings|maya philippines|maya wallet"),
        GCASH("GCash", "gcash"),
        GOTYME("GoTyme", "gotyme"),
        BPI("BPI", "bank of the philippine islands|\\bbpi\\b"),
        // UnionBank's transaction history shows its name only in the logo image, so its table header
        // (printed above the rows, where a "BPI" transfer could otherwise come first) identifies it.
        UNIONBANK("UnionBank", "unionbank|union bank of the philippines|check no\\. ref\\. no\\."),
        // UnionBank credit card statement of account: its first line, and the card's product name.
        // Card statement of account (first line, product name) or card transaction history (its table header).
        UNIONBANK_CARD("UnionBank Credit Card",
                "statement of account available points|rewards platinum mastercard|transaction date posting date transaction details", true);

        public final String label;
        /** A credit card statement: balances are owed, a positive amount is a charge and a negative one a payment. */
        public final boolean card;
        private final Pattern mark;

        Bank(String label, String mark) {
            this(label, mark, false);
        }

        Bank(String label, String mark, boolean card) {
            this.label = label;
            this.card = card;
            this.mark = Pattern.compile(mark, Pattern.CASE_INSENSITIVE);
        }
    }

    // UnionBank credit card summary: "Statement Date: October 23, 2026", "Previous Balance … PHP 14,485.15",
    // and "Due Date Minimum Amount Due Total Amount Due" over "November 09, 2026 PHP 643.11 PHP 21,247.65".
    private static final Pattern CARD_STATEMENT_DATE = Pattern.compile("(?i)statement date:\\s*([A-Za-z]+ \\d{1,2}, \\d{4})");
    private static final Pattern CARD_PREVIOUS = Pattern.compile("(?i)previous balance\\s+(?:\\S+\\s+)?PHP\\s*(-?[\\d,]+\\.\\d{2})");
    private static final Pattern CARD_DUE = Pattern.compile(
            "(?i)total amount due\\s+([A-Za-z]+ \\d{1,2}, \\d{4})\\s+PHP\\s*([\\d,]+\\.\\d{2})\\s+PHP\\s*(-?[\\d,]+\\.\\d{2})");
    /** In a card transaction history: "Virtual Card (…)" then "Card Number: 5292 47** **** 1135". */
    private static final Pattern CARD_VIRTUAL = Pattern.compile("(?i)virtual card[^\\n]*\\R\\s*card number:\\s*[\\d* ]*?(\\d{4})\\s*$", Pattern.MULTILINE);
    /** The card's last four digits, printed as "* * * *  * * * *  * * * *  1 1 3 5". */
    private static final Pattern CARD_LAST4 = Pattern.compile("\\*[\\s*]*\\*\\s+((?:\\d\\s?){4})(?!\\s?\\d)");

    /** What a credit card statement says is owed and when, or null when the text isn't a card statement. */
    public record CardSummary(LocalDate statementDate, BigDecimal previousBalance, BigDecimal totalDue,
                              BigDecimal minimumDue, LocalDate dueDate) {}

    public static CardSummary cardSummary(String text) {
        Matcher date = CARD_STATEMENT_DATE.matcher(text);
        Matcher previous = CARD_PREVIOUS.matcher(text);
        Matcher due = CARD_DUE.matcher(text);
        if (!date.find() || !due.find()) {
            return null;
        }
        DateTimeFormatter longDate = fmt("MMMM d, yyyy");
        return new CardSummary(
                LocalDate.parse(date.group(1), longDate),
                previous.find() ? money(previous.group(1)) : null,
                money(due.group(3)),
                money(due.group(2)),
                LocalDate.parse(due.group(1), longDate));
    }

    private static BigDecimal money(String amount) {
        return new BigDecimal(amount.replace(",", ""));
    }

    /**
     * The bank named earliest in the text, or null. The header comes first, so a bank that only shows up
     * in a transaction ("Transfer to GCash") doesn't win.
     * shortcut: only MariBank's header is checked against a real statement; add the others' wording when samples turn up.
     */
    public static Bank detectBank(String text) {
        Bank found = null;
        int at = Integer.MAX_VALUE;
        for (Bank bank : Bank.values()) {
            Matcher m = bank.mark.matcher(text);
            if (m.find() && m.start() < at) {
                found = bank;
                at = m.start();
            }
        }
        return found;
    }

    /** "MARIBANK ACCOUNT: 1234…", "Account No. 1234-5678-90", "GCash Number: 0917…", "Account number: +63915…". */
    private static final Pattern ACCOUNT_NUMBER = Pattern.compile(
            "\\b(?:account|acct|gcash|mobile)\\s*(?:no\\.?|number|#)?\\s*[:.]?\\s*\\+?(\\d{2,}(?:[ -]\\d{2,})*)",
            Pattern.CASE_INSENSITIVE);

    /** Maya statements that are money owed rather than held; their layout and rules aren't known yet. */
    private static final Pattern MAYA_CREDIT = Pattern.compile("(?i)easy credit|personal loan");

    /**
     * Why this statement can't be imported, or null when it can.
     * shortcut: only the first 800 characters (the statement's title area) are checked, so a wallet
     * payment to Easy Credit further down doesn't count; revisit with real Easy Credit / loan samples.
     */
    public static String unsupported(String text, Bank bank) {
        if (bank == Bank.MAYA) {
            Matcher m = MAYA_CREDIT.matcher(text.substring(0, Math.min(800, text.length())));
            if (m.find()) {
                return "Maya " + (m.group().equalsIgnoreCase("easy credit") ? "Easy Credit" : "Personal Loan")
                        + " statements can't be imported yet: they're money owed, not money held. Maya Wallet and Maya Bank statements work.";
            }
        }
        return null;
    }

    // "OFF-US ATM WITHDRAWAL" (UnionBank), "Overseas ATM Cash Withdrawal" (MariBank), "Withdrawal from ARANETA CENTER 2" (Maya),
    // "W/D FR SAV BDO" (BDO, at the branch).
    private static final Pattern CASH_WITHDRAWAL = Pattern.compile("(?i)\\batm\\b.*withdrawal|cash withdrawal|^withdrawal from\\b|^w/d fr sav\\b");

    // Card charges that are the bank's own: interest, and its named fees. "Rent Fees" paid to a landlord isn't one.
    private static final Pattern CARD_INTEREST = Pattern.compile("(?i)\\binterest\\b|finance charge");
    private static final Pattern CARD_FEE = Pattern.compile("(?i)\\b(processing|annual|membership|late(?: payment)?|service|cash advance|overlimit)\\s+fee\\b");

    /** UnionBank payroll credit: an S-reference and the employee's name as "Surname, Given names". */
    private static final Pattern UNIONBANK_PAYROLL = Pattern.compile("^S\\d+ [A-Za-z][A-Za-z .'-]*, [A-Za-z][A-Za-z .'-]*$");

    /**
     * The income / expense account a row belongs to, or "" for Incoming / Outgoing Transfers.
     * shortcut: one built-in rule (UnionBank payroll → Salary Income); move to user-editable rules when a second one is needed.
     */
    public static String category(Bank bank, Line line) {
        if (bank == Bank.UNIONBANK && !line.debit() && UNIONBANK_PAYROLL.matcher(line.description()).matches()) {
            return "Salary Income";
        }
        // Cash taken out at an ATM or counter goes into the user's Cash account, not spending.
        if (!bank.card && line.debit() && CASH_WITHDRAWAL.matcher(line.description()).find()) {
            return "Cash";
        }
        if (bank.card && line.debit()) {
            if (CARD_INTEREST.matcher(line.description()).find()) return "Loan Interest";
            if (CARD_FEE.matcher(line.description()).find()) return "Loan Fees";
        }
        return "";
    }

    // "Bills Payment for Maya Personal Loan" (Maya Wallet).
    private static final Pattern MAYA_LOAN_PAYMENT = Pattern.compile("(?i)\\bfor maya personal loan\\b");

    /** The loan account a payment goes into ("Maya Personal Loan"), or null when the row doesn't pay a known loan. */
    public static String loanPaidBy(Bank bank, Line line) {
        return bank == Bank.MAYA && line.debit() && MAYA_LOAN_PAYMENT.matcher(line.description()).find() ? "Maya Personal Loan" : null;
    }

    /** Name for an account created from a statement: Maya's mobile-number account is the wallet, any other its bank savings. */
    public static String accountLabel(Bank bank, String accountNumber) {
        boolean mobile = accountNumber != null && accountNumber.matches("(63|0)9\\d{9}");
        return bank == Bank.MAYA ? (mobile ? "Maya Wallet" : "Maya Savings") : bank.label;
    }

    /** Maya Savings prints its number unlabelled, as the first thing on the line under "Consumer Savings": "8083 0917 0005". */
    private static final Pattern MAYA_SAVINGS_NUMBER = Pattern.compile("(?m)^(\\d{4} \\d{4} \\d{4})(?!\\d)");

    /** The statement's account number as digits only, or null when none is printed. */
    public static String accountNumber(String text, Bank bank) {
        if (bank != null && bank.card) {
            // A transaction history lists every card on the account; the virtual card's number is the one the
            // monthly statement prints, so it's the one that keeps both in the same account.
            // shortcut: assumes the statement always prints the virtual card; revisit if a statement shows the principal card.
            Matcher virtual = CARD_VIRTUAL.matcher(text);
            if (virtual.find()) {
                return virtual.group(1);
            }
            Matcher m = CARD_LAST4.matcher(text);
            return m.find() ? m.group(1).replaceAll("\\D", "") : null;
        }
        String number = accountNumber(text);
        if (number == null && bank == Bank.MAYA) {
            Matcher m = MAYA_SAVINGS_NUMBER.matcher(text);
            return m.find() ? m.group(1).replace(" ", "") : null;
        }
        return number;
    }

    public static String accountNumber(String text) {
        Matcher m = ACCOUNT_NUMBER.matcher(text);
        while (m.find()) {
            String digits = m.group(1).replaceAll("\\D", "");
            if (digits.length() >= 6 && digits.length() <= 34) {
                return digits;
            }
        }
        return null;
    }

    public record Line(LocalDate date, String description, BigDecimal amount, boolean debit) {}

    /** Balances and period dates are null when the statement doesn't print them. */
    public record Result(List<Line> lines, List<String> skipped,
                         BigDecimal openingBalance, BigDecimal closingBalance,
                         LocalDate periodStart, LocalDate periodEnd) {}

    private record DateFormat(Pattern pattern, DateTimeFormatter formatter, boolean hasYear) {}

    private record Amount(BigDecimal signed, Boolean debit) {}

    private record Dated(LocalDate date, String rest) {}

    /** A row before its direction is settled; debit is null until the running balance decides it. */
    private record Row(LocalDate date, String description, BigDecimal amount, Boolean debit, BigDecimal balance, String raw) {
        Row describedAs(String text) {
            return new Row(date, text, amount, debit, balance, raw);
        }
    }

    /** Year for dates printed without one: the period's start year, or its end year for dates before the start (Dec–Jan statements). */
    private record YearRule(int year, LocalDate periodStart, LocalDate periodEnd, int endYear) {}

    private static final List<DateFormat> DATE_FORMATS = List.of(
            new DateFormat(Pattern.compile("^(\\d{4}-\\d{2}-\\d{2})"), fmt("yyyy-MM-dd"), true),
            new DateFormat(Pattern.compile("^(\\d{1,2}/\\d{1,2}/\\d{4})"), fmt("M/d/yyyy"), true),
            new DateFormat(Pattern.compile("^(\\d{1,2}/\\d{1,2}/\\d{2})(?!\\d)"), fmt("M/d/yy"), true),
            new DateFormat(Pattern.compile("^(\\d{1,2} [A-Za-z]{3} \\d{4})"), fmt("d MMM yyyy"), true),
            new DateFormat(Pattern.compile("^(\\d{1,2}-[A-Za-z]{3}-\\d{4})"), fmt("d-MMM-yyyy"), true),
            new DateFormat(Pattern.compile("^([A-Za-z]{3} \\d{1,2},? \\d{4})"), fmt("MMM d yyyy"), true),
            new DateFormat(Pattern.compile("^([A-Za-z]{3} \\d{1,2})(?![\\d,.])"), fmt("MMM d yyyy"), false),
            new DateFormat(Pattern.compile("^(\\d{1,2} [A-Za-z]{3})(?![A-Za-z])"), fmt("d MMM yyyy"), false)
    );

    private static final Pattern PERIOD = Pattern.compile(
            "(\\d{1,2}\\s+[A-Za-z]{3}\\s+20\\d{2})\\s+(?:to|-)\\s+(\\d{1,2}\\s+[A-Za-z]{3}\\s+20\\d{2})",
            Pattern.CASE_INSENSITIVE);
    private static final Pattern PAGE_FOOTER = Pattern.compile("(?i)page \\d+ of \\d+");
    private static final Pattern PERIOD_ISO = Pattern.compile("(20\\d{2}-[01]\\d-[0-3]\\d) to (20\\d{2}-[01]\\d-[0-3]\\d)");
    private static final Pattern LONE_AMOUNT =Pattern.compile("-?[\\d,]+\\.\\d{2}");
    private static final Pattern PERIOD_DAYS = Pattern.compile("(?m)^(\\d{1,2})-(\\d{1,2}) ([A-Za-z]{3} 20\\d{2})$");
    /** Sections after the transaction list whose dated rows are not transactions (e.g. MariBank's daily interest table). */
    private static final Pattern STOP = Pattern.compile("interest & tax details");
    /** MariBank's monthly interest row has a month but no day ("OCT 21.82"). */
    private static final Pattern MONTH_ONLY = Pattern.compile("^([A-Za-z]{3}) (?=\\(?-?[\\d,]+\\.\\d{2})");

    private static final Pattern TIME = Pattern.compile("^\\d{1,2}:\\d{2}(?::\\d{2})?\\s*(?:[AaPp][Mm])?");
    // "PHP44,000.00" (BDO) has no space before the amount.
    private static final Pattern CURRENCY = Pattern.compile("(?i)\\bPHP(?![A-Za-z])|₱");
    private static final Pattern NEGATIVE_CURRENCY = Pattern.compile("(?i)-\\s*(?:PHP|₱)\\s*(?=\\d)");
    private static final Pattern AMOUNT = Pattern.compile(
            "^(\\()?(-)?((?:\\d{1,3}(?:,\\d{3})+|\\d+)\\.\\d{2})\\)?(-)?(DR|CR)?$",
            Pattern.CASE_INSENSITIVE);
    private static final Pattern YEAR = Pattern.compile("\\b(20\\d{2})\\b");
    private static final Pattern OPENING = Pattern.compile(
            "beginning balance|opening balance|previous balance|starting balance|balance forward|balance brought forward");
    private static final Pattern CLOSING = Pattern.compile(
            "ending balance|closing balance|available balance|total debit|total credit|^total\\b");

    private BankStatementParser() {
    }

    public static Result parse(String text, Bank bank) {
        YearRule years = yearRule(text);
        // MariBank prints the counterparty on the line above the date/amount and the type on the line below.
        boolean nameAboveTypeBelow = bank == Bank.MARIBANK;
        // UnionBank and BDO wrap long descriptions onto the line below the row.
        boolean wrapsBelow = bank == Bank.UNIONBANK || bank == Bank.BDO;
        // Maya puts the time and the transaction's details on the line below ("11:08:52 PM Credit via InstaPay").
        boolean detailBelow = bank == Bank.MAYA;
        // GCash wraps a long description and prints the row's amounts on the wrapped line ("… ending in 2043 and" / "invno:… 150.00 282.56").
        boolean amountsBelow = bank == Bank.GCASH;
        String held = null;

        List<Row> rows = new ArrayList<>();
        List<String> skipped = new ArrayList<>();
        BigDecimal openingBalance = null;
        BigDecimal closingBalance = null;
        boolean summaryHeader = false;
        String lineAbove = "";
        int awaitingType = -1;

        for (String raw : text.split("\\R")) {
            // "-PHP 7,862.36" (Maya) keeps its sign: "-7,862.36".
            String line = CURRENCY.matcher(NEGATIVE_CURRENCY.matcher(raw).replaceAll("-")).replaceAll("").replaceAll("\\s+", " ").trim();
            if (detailBelow) {
                // Maya Savings: "- 4,990.28" → "-4,990.28".
                line = line.replaceAll("(^| )- (?=\\d)", "$1-");
            }
            if (line.isEmpty()) {
                continue;
            }
            boolean joined = false;
            if (held != null) {
                // Only an undated line can be the held row's wrapped part; join once, so a dated title doesn't swallow the page.
                if (leadingDate(line, years) == null) {
                    line = held + " " + line;
                    joined = true;
                }
                held = null;
            }
            String lower = line.toLowerCase(Locale.ROOT);
            if (PAGE_FOOTER.matcher(line).matches()) {
                awaitingType = -1; // not a wrapped description
                continue;
            }
            if (STOP.matcher(lower).find()) {
                break;
            }

            // Summary table ("STARTING BALANCE ... ENDING BALANCE" header, then a row of amounts):
            // the first amount is the opening balance and the last is the closing balance.
            if (summaryHeader) {
                List<Amount> amounts = trailingAmounts(new ArrayList<>(List.of(line.split(" "))));
                if (amounts.size() >= 2) {
                    summaryHeader = false;
                    if (openingBalance == null) {
                        openingBalance = amounts.get(0).signed();
                        closingBalance = amounts.get(amounts.size() - 1).signed();
                    }
                }
                continue; // e.g. the "(PHP) (PHP)" unit row under the header
            }

            if (OPENING.matcher(lower).find()) {
                List<Amount> amounts = trailingAmounts(new ArrayList<>(List.of(line.split(" "))));
                if (!amounts.isEmpty()) {
                    if (openingBalance == null) {
                        openingBalance = amounts.get(amounts.size() - 1).signed();
                    }
                } else if (lower.contains("ending balance")) {
                    summaryHeader = true;
                }
                continue;
            }
            if (CLOSING.matcher(lower).find()) {
                if (closingBalance == null && (lower.contains("ending balance") || lower.contains("closing balance"))) {
                    List<Amount> amounts = trailingAmounts(new ArrayList<>(List.of(line.split(" "))));
                    if (!amounts.isEmpty()) {
                        closingBalance = amounts.get(amounts.size() - 1).signed();
                    }
                }
                continue;
            }
            // Maya Savings' summary starts with the opening balance, unlabelled, on a line of its own.
            if (detailBelow && openingBalance == null && rows.isEmpty() && LONE_AMOUNT.matcher(line).matches()) {
                openingBalance = new BigDecimal(line.replace(",", ""));
                continue;
            }

            Dated dated = leadingDate(line, years);
            Matcher monthOnly = MONTH_ONLY.matcher(line);
            if (dated == null && nameAboveTypeBelow && monthOnly.find()) {
                dated = monthEnd(monthOnly.group(1), line.substring(monthOnly.end()), years);
            }
            if (dated == null) {
                // header, footer, or wrapped description line
                Matcher time = TIME.matcher(line);
                boolean timeDetail = detailBelow && time.find();
                if (awaitingType >= 0 && (nameAboveTypeBelow || timeDetail
                        || (wrapsBelow && line.length() <= 40))) { // a footer sentence is longer
                    Row row = rows.get(awaitingType);
                    String extra = timeDetail ? line.substring(time.end()).trim() : line;
                    rows.set(awaitingType, row.describedAs(
                            extra.isEmpty() ? row.description()
                            : nameAboveTypeBelow || timeDetail ? row.description() + " (" + extra + ")"
                            : row.description() + " " + extra));
                } else {
                    lineAbove = line;
                }
                awaitingType = -1;
                continue;
            }
            awaitingType = -1;
            String above = lineAbove;
            lineAbove = "";
            String rest = dated.rest();
            if (bank.card) {
                // Card transaction histories print the posting date after the transaction date: drop it.
                Dated posted = leadingDate(rest, years);
                if (posted != null) {
                    rest = posted.rest();
                }
            }
            List<String> tokens = new ArrayList<>(List.of(rest.split(" ")));
            List<Amount> amounts = trailingAmounts(tokens);
            if (amounts.isEmpty()) {
                if (amountsBelow && !joined) {
                    held = line;
                }
                continue;
            }

            Amount first = amounts.get(0);
            BigDecimal balance = amounts.size() >= 2 ? amounts.get(amounts.size() - 1).signed() : null;
            // On a card statement a positive amount is a charge (money out: you owe more), a negative one a payment.
            Boolean debit = bank.card ? Boolean.valueOf(first.signed().signum() > 0) : first.debit();
            if (debit == null && balance == null && bank == Bank.BDO) {
                debit = false; // BDO prints money out as "-PHP…" and money in unsigned, with no balance column.
            }
            if (debit == null && balance == null) {
                skipped.add(raw.trim());
                continue;
            }

            String description = String.join(" ", tokens).trim();
            if (nameAboveTypeBelow) {
                description = (above + " " + description).trim();
            }
            if (nameAboveTypeBelow || wrapsBelow || detailBelow) {
                awaitingType = rows.size();
            }
            rows.add(new Row(dated.date(), description.isEmpty() ? "Statement entry" : description,
                    first.signed().abs(), debit, balance, raw.trim()));
        }

        // Directions the text doesn't print come from the running balance, read oldest first.
        boolean newestFirst = balanceSteps(rows, true) > balanceSteps(rows, false);
        List<Row> oldestFirst = new ArrayList<>(rows);
        if (newestFirst) {
            Collections.reverse(oldestFirst);
        }

        List<Line> lines = new ArrayList<>();
        BigDecimal before = openingBalance;
        boolean first = true;
        for (Row row : oldestFirst) {
            Boolean debit = row.debit();
            if (debit == null) {
                if (before != null) {
                    debit = row.balance().compareTo(before) < 0;
                } else if (row.balance().compareTo(row.amount()) < 0) {
                    // No opening balance printed (UnionBank's history): money in would mean the account
                    // started below zero, which a bank or wallet account can't, so it was money out.
                    debit = true;
                } else if (row.balance().compareTo(row.amount()) == 0) {
                    // shortcut: balance = amount is read as the account's first money in (from zero); money out
                    // from twice the amount is possible but rarer. Anything else goes to skipped.
                    debit = false;
                }
            }
            if (debit == null) {
                skipped.add(row.raw());
            } else {
                lines.add(new Line(row.date(), row.description(), row.amount(), debit));
                if (first && openingBalance == null && row.balance() != null) {
                    openingBalance = debit ? row.balance().add(row.amount()) : row.balance().subtract(row.amount());
                }
            }
            first = false;
            if (row.balance() != null) {
                before = row.balance();
            }
        }
        if (closingBalance == null && !oldestFirst.isEmpty() && oldestFirst.get(oldestFirst.size() - 1).balance() != null) {
            closingBalance = oldestFirst.get(oldestFirst.size() - 1).balance();
        }
        if (newestFirst) {
            Collections.reverse(lines);
        }
        if (bank.card) {
            // A card's balance is what's owed, so the ledger holds it negative: previous balance → total due,
            // over the month that ends on the statement date.
            CardSummary card = cardSummary(text);
            if (card != null) {
                return new Result(lines, skipped,
                        card.previousBalance() == null ? null : card.previousBalance().negate(),
                        card.totalDue().negate(),
                        card.statementDate().minusMonths(1).plusDays(1),
                        card.statementDate());
            }
        }
        return new Result(lines, skipped, openingBalance, closingBalance, years.periodStart(), years.periodEnd());
    }

    /**
     * How many rows' balance differs from the row before them in time by exactly their amount, reading
     * the list as newest first or as oldest first. The order with more such steps is the real one.
     */
    private static int balanceSteps(List<Row> rows, boolean newestFirst) {
        int steps = 0;
        for (int i = 0; i < rows.size(); i++) {
            int previous = newestFirst ? i + 1 : i - 1;
            if (previous < 0 || previous >= rows.size()) {
                continue;
            }
            BigDecimal now = rows.get(i).balance();
            BigDecimal then = rows.get(previous).balance();
            if (now != null && then != null && now.subtract(then).abs().compareTo(rows.get(i).amount()) == 0) {
                steps++;
            }
        }
        return steps;
    }

    /** Dates a month-only row on the month's last day; null if {@code month} isn't a month name. */
    private static Dated monthEnd(String month, String rest, YearRule years) {
        try {
            LocalDate first = LocalDate.parse("1 " + month + " " + years.year(), fmt("d MMM yyyy"));
            LocalDate date = first.withDayOfMonth(first.lengthOfMonth());
            if (years.periodStart() != null && date.isBefore(years.periodStart())) {
                date = date.withYear(years.endYear());
            }
            return new Dated(date, rest.trim());
        } catch (DateTimeParseException e) {
            return null;
        }
    }

    private static YearRule yearRule(String text) {
        Matcher period = PERIOD.matcher(text);
        if (period.find()) {
            try {
                LocalDate start = LocalDate.parse(period.group(1).replaceAll("\\s+", " "), fmt("d MMM yyyy"));
                LocalDate end = LocalDate.parse(period.group(2).replaceAll("\\s+", " "), fmt("d MMM yyyy"));
                return new YearRule(start.getYear(), start, end, end.getYear());
            } catch (DateTimeParseException ignored) {
                // fall through to the first year mentioned
            }
        }
        // GCash: "2026-09-10 to 2026-11-08".
        Matcher iso = PERIOD_ISO.matcher(text);
        if (iso.find()) {
            LocalDate start = LocalDate.parse(iso.group(1));
            LocalDate end = LocalDate.parse(iso.group(2));
            return new YearRule(start.getYear(), start, end, end.getYear());
        }
        // Maya Savings: "1-30 Sep 2026" (its statement date, printed first, can be next year's).
        Matcher days = PERIOD_DAYS.matcher(text);
        if (days.find()) {
            try {
                LocalDate start = LocalDate.parse(days.group(1) + " " + days.group(3), fmt("d MMM yyyy"));
                LocalDate end = LocalDate.parse(days.group(2) + " " + days.group(3), fmt("d MMM yyyy"));
                return new YearRule(start.getYear(), start, end, end.getYear());
            } catch (DateTimeParseException ignored) {
                // fall through to the first year mentioned
            }
        }
        // shortcut: without a "d MMM yyyy to d MMM yyyy" period, year-less dates take the first 20xx year in the text; Dec–Jan statements get Dec rows wrong. Add the bank's period format to PERIOD when one turns up.
        Matcher yearMatch = YEAR.matcher(text);
        int year = yearMatch.find() ? Integer.parseInt(yearMatch.group(1)) : LocalDate.now().getYear();
        return new YearRule(year, null, null, year);
    }

    /** Returns the date and the rest of the line, or null when the line doesn't start with a date. */
    private static Dated leadingDate(String line, YearRule years) {
        for (DateFormat format : DATE_FORMATS) {
            Matcher m = format.pattern().matcher(line);
            if (!m.find()) {
                continue;
            }
            String value = m.group(1).replace(",", "");
            if (!format.hasYear()) {
                value = value + " " + years.year();
            }
            try {
                LocalDate date = LocalDate.parse(value, format.formatter());
                if (!format.hasYear() && years.periodStart() != null && date.isBefore(years.periodStart())) {
                    date = date.withYear(years.endYear());
                }
                // Maya Savings: "7 Sep, 04:35 PM Wallet Transfer …".
                String rest = TIME.matcher(line.substring(m.end()).replaceFirst("^,", "").trim()).replaceFirst("").trim();
                return new Dated(date, rest);
            } catch (DateTimeParseException ignored) {
                // e.g. "Ref 12" looks like "MMM d" but isn't a month
            }
        }
        return null;
    }

    /** Pops amount tokens (and standalone DR/CR markers) off the end of {@code tokens}, in line order. */
    private static List<Amount> trailingAmounts(List<String> tokens) {
        List<Amount> amounts = new ArrayList<>();
        Boolean pendingMarker = null;
        while (!tokens.isEmpty()) {
            String token = tokens.get(tokens.size() - 1);
            String upper = token.toUpperCase(Locale.ROOT);
            if (upper.equals("DR") || upper.equals("CR")) {
                pendingMarker = upper.equals("DR");
                tokens.remove(tokens.size() - 1);
                continue;
            }
            Matcher m = AMOUNT.matcher(token);
            if (!m.matches()) {
                break;
            }
            tokens.remove(tokens.size() - 1);
            boolean negative = m.group(1) != null || m.group(2) != null || m.group(4) != null;
            BigDecimal value = new BigDecimal(m.group(3).replace(",", ""));
            Boolean debit = negative ? Boolean.TRUE
                    : m.group(5) != null ? Boolean.valueOf(m.group(5).equalsIgnoreCase("DR"))
                    : pendingMarker;
            amounts.add(0, new Amount(negative ? value.negate() : value, debit));
            pendingMarker = null;
        }
        return amounts;
    }

    private static DateTimeFormatter fmt(String pattern) {
        return new DateTimeFormatterBuilder().parseCaseInsensitive().appendPattern(pattern)
                .toFormatter(Locale.ENGLISH);
    }
}
