package com.finankal.api.service;

import com.finankal.api.dto.AccountDto;
import com.finankal.api.dto.BankStatementDto;
import com.finankal.api.dto.StatementImportResultDto;
import com.finankal.api.dto.StatementImportResultDto.Status;
import com.finankal.api.finance.FinanceEngineGrpc;
import com.finankal.api.finance.FinanceProtos;
import com.finankal.api.service.BankStatementParser.Bank;
import com.finankal.api.service.BankStatementParser.Line;
import io.grpc.StatusRuntimeException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.time.Instant;
import java.time.LocalDate;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.HexFormat;
import java.util.List;
import java.util.Locale;
import java.util.UUID;

/**
 * Parses PDF statements and hands each to the engine, which imports it in one transaction
 * (see docs/adr/0001-bank-statement-import.md). The bank and account are detected from the
 * statement unless the caller chooses them.
 */
@Service
public class StatementImportService {

    private static final Logger logger = LoggerFactory.getLogger(StatementImportService.class);

    @Autowired
    private FinanceEngineGrpc.FinanceEngineBlockingStub financeEngineStub;

    @Autowired
    private UserService userService;

    public record Upload(String fileName, byte[] pdf) {}

    /** bill is a credit card statement's summary (null for bank statements and card histories). */
    private record Parsed(String fileName, Bank bank, BankStatementParser.Result result, String accountNumber,
                          LocalDate start, LocalDate end, String sha256, BankStatementParser.CardSummary bill) {}

    /**
     * bank and accountId are optional: null means detect from each statement. One file failing,
     * or already being imported, doesn't stop the others.
     */
    public List<StatementImportResultDto> importStatements(
            List<Upload> uploads,
            String password,
            Bank bank,
            String userId,
            String accountId
    ) {
        if (userId == null || userId.isBlank()) {
            throw new IllegalArgumentException("userId is required");
        }
        // The account id comes from the browser; check ownership before the engine uses it.
        if (accountId != null && userService.getUserAccounts(UUID.fromString(userId)).stream().noneMatch(a -> a.getId().equals(accountId))) {
            throw new IllegalArgumentException("The selected account doesn't belong to this user");
        }

        List<StatementImportResultDto> results = new ArrayList<>();
        List<Parsed> ready = new ArrayList<>();
        for (Upload upload : uploads) {
            try {
                if (upload.pdf().length == 0) {
                    throw new IllegalArgumentException("The file is empty.");
                }
                String text = StatementPdfText.extract(upload.pdf(), password);
                if (LoanDisclosureParser.isDisclosure(text)) {
                    // Loans first, so statements in the same upload can pay into them.
                    results.add(sendLoan(upload, LoanDisclosureParser.parse(text), userId));
                } else {
                    ready.add(parse(upload, text, bank));
                }
            } catch (IllegalArgumentException e) {
                results.add(StatementImportResultDto.notImported(upload.fileName(), Status.FAILED, null, e.getMessage()));
            }
        }
        // Oldest first, so each statement's balance check already includes the months before it.
        ready.sort(Comparator.comparing(Parsed::start));
        for (Parsed parsed : ready) {
            results.add(send(parsed, userId, accountId));
        }
        return results;
    }

    /** A loan disclosure statement: the engine creates the loan, books what was lent and deducted, and its deadlines. */
    private StatementImportResultDto sendLoan(Upload upload, LoanDisclosureParser.Loan loan, String userId) {
        // Payments come from the lender's own wallet / bank account, when it's tracked.
        String source = userService.getUserAccounts(UUID.fromString(userId)).stream()
                .filter(a -> "ASSET".equals(a.getType()) && a.getName().startsWith(loan.lender()))
                .map(AccountDto::getId)
                .findFirst().orElse("");
        FinanceProtos.ImportLoanRequest.Builder request = FinanceProtos.ImportLoanRequest.newBuilder()
                .setUserId(userId)
                .setAccountNumber(loan.accountNumber())
                .setAccountName(loan.accountName())
                .setBookedOn(loan.bookedOn().toString())
                .setAmount(loan.amount().toPlainString())
                .setDeducted(loan.deducted().toPlainString())
                .setDailyRate(loan.dailyRate().toPlainString())
                .setSourceAccountId(source)
                .setScheduleName(String.format("%s ₱%,.0f (%d months, from %s)", loan.accountName(), loan.amount(),
                        loan.installments().size(), loan.bookedOn().format(DateTimeFormatter.ofPattern("dd MMM yyyy", Locale.ENGLISH))))
                .setFileSha256(sha256(upload.pdf()));
        for (LoanDisclosureParser.Installment in : loan.installments()) {
            request.addInstallments(FinanceProtos.Installment.newBuilder()
                    .setDueDate(in.dueDate().toString())
                    .setAmount(in.payment().toPlainString()));
        }

        FinanceProtos.ImportLoanResponse response;
        try {
            response = financeEngineStub.importLoan(request.build());
        } catch (StatusRuntimeException e) {
            String reason = e.getStatus().getDescription();
            switch (e.getStatus().getCode()) {
                case ALREADY_EXISTS:
                    return StatementImportResultDto.notImported(upload.fileName(), Status.ALREADY_IMPORTED, loan.accountName(), reason);
                case INVALID_ARGUMENT, NOT_FOUND:
                    return StatementImportResultDto.notImported(upload.fileName(), Status.FAILED, loan.accountName(), reason);
                default:
                    throw e;
            }
        }
        logger.info("Imported {} loan {} with {} installments", loan.lender(), response.getAccountId(), loan.installments().size());
        // Two transactions: the money lent (less deductions) and the deductions themselves.
        return new StatementImportResultDto(upload.fileName(), Status.IMPORTED, loan.accountName(), response.getAccountName(),
                loan.deducted().signum() > 0 ? 2 : 1, List.of(), List.of(), List.of(), List.of(), null);
    }

    public List<BankStatementDto> listStatements(UUID userId) {
        FinanceProtos.ListBankStatementsResponse response = financeEngineStub.listBankStatements(
                FinanceProtos.ListBankStatementsRequest.newBuilder().setUserId(userId.toString()).build());
        return response.getStatementsList().stream()
                .map(s -> new BankStatementDto(s.getId(), s.getAccountId(), s.getBank(), s.getPeriodStart(), s.getPeriodEnd(),
                        Instant.ofEpochSecond(s.getImportedAt().getSeconds(), s.getImportedAt().getNanos()).toString()))
                .toList();
    }

    private static Parsed parse(Upload upload, String text, Bank chosen) {
        Bank bank = chosen != null ? chosen : BankStatementParser.detectBank(text);
        if (bank == null) {
            throw new IllegalArgumentException("Couldn't tell which bank this statement is from. Choose the bank and import it again.");
        }

        String unsupported = BankStatementParser.unsupported(text, bank);
        if (unsupported != null) {
            throw new IllegalArgumentException(unsupported);
        }

        BankStatementParser.Result parsed = BankStatementParser.parse(text, bank);
        LocalDate start = parsed.periodStart() != null ? parsed.periodStart()
                : parsed.lines().stream().map(Line::date).min(LocalDate::compareTo).orElse(null);
        LocalDate end = parsed.periodEnd() != null ? parsed.periodEnd()
                : parsed.lines().stream().map(Line::date).max(LocalDate::compareTo).orElse(null);
        if (start == null || end == null) {
            throw new IllegalArgumentException("No transactions or statement period were found in this PDF. Check that it's a "
                    + bank.label + " statement.");
        }
        BankStatementParser.CardSummary bill = bank.card ? BankStatementParser.cardSummary(text) : null;
        return new Parsed(upload.fileName(), bank, parsed, BankStatementParser.accountNumber(text, bank), start, end,
                sha256(upload.pdf()), bill);
    }

    private StatementImportResultDto send(Parsed p, String userId, String accountId) {
        String number = p.accountNumber() == null ? "" : p.accountNumber();
        List<AccountDto> accounts = userService.getUserAccounts(UUID.fromString(userId));
        if (accountId == null && number.isEmpty() && !p.bank().card) {
            // No number printed (GCash's transaction history, some Maya Savings statements): use the user's only
            // account of that kind, e.g. their one "GCash…" or "Maya Savings…" account.
            String label = BankStatementParser.accountLabel(p.bank(), null);
            List<String> same = accounts.stream()
                    .filter(a -> "ASSET".equals(a.getType()) && a.getName().startsWith(label))
                    .map(AccountDto::getId).toList();
            if (same.size() == 1) {
                accountId = same.get(0);
            }
        }
        FinanceProtos.ImportBankStatementRequest.Builder request = FinanceProtos.ImportBankStatementRequest.newBuilder()
                .setUserId(userId)
                .setAccountId(accountId == null ? "" : accountId)
                .setAccountNumber(number)
                .setAccountName(BankStatementParser.accountLabel(p.bank(), p.accountNumber())
                        + (number.isEmpty() ? "" : " ••" + number.substring(Math.max(0, number.length() - 4))))
                .setAccountType(p.bank().card ? "CREDIT_CARD" : "ASSET")
                // The card bill's deadline, as printed; the engine tracks it until payments cover it.
                .setStatementDate(p.bill() == null ? "" : p.bill().statementDate().toString())
                .setDueDate(p.bill() == null || p.bill().dueDate() == null ? "" : p.bill().dueDate().toString())
                .setTotalDue(p.bill() == null ? "" : p.bill().totalDue().toPlainString())
                .setBank(p.bank().name())
                .setPeriodStart(p.start().toString())
                .setPeriodEnd(p.end().toString())
                .setOpeningBalance(amount(p.result().openingBalance()))
                .setClosingBalance(amount(p.result().closingBalance()))
                .setFileSha256(p.sha256());
        // A payment naming one of the user's loans goes into it (the engine moves money to an own account named
        // by the category), so its interest and deadlines follow. Without that loan it stays an Outgoing Transfer.
        List<String> loans = accounts.stream()
                .filter(a -> "LIABILITY".equals(a.getType())).map(AccountDto::getName).toList();
        for (Line line : p.result().lines()) {
            String loan = BankStatementParser.loanPaidBy(p.bank(), line);
            request.addRows(FinanceProtos.StatementRow.newBuilder()
                    .setDate(line.date().toString())
                    .setDescription(line.description())
                    .setAmount(line.amount().toPlainString())
                    .setMoneyOut(line.debit())
                    .setCategory(loan != null && loans.contains(loan) ? loan : BankStatementParser.category(p.bank(), line)));
        }

        FinanceProtos.ImportBankStatementResponse response;
        try {
            response = financeEngineStub.importBankStatement(request.build());
        } catch (StatusRuntimeException e) {
            String reason = e.getStatus().getDescription();
            switch (e.getStatus().getCode()) {
                case ALREADY_EXISTS:
                    return StatementImportResultDto.notImported(p.fileName(), Status.ALREADY_IMPORTED, p.bank().label, reason);
                case INVALID_ARGUMENT, NOT_FOUND:
                    return StatementImportResultDto.notImported(p.fileName(), Status.FAILED, p.bank().label, reason);
                default:
                    throw e;
            }
        }

        logger.info("Imported {} rows from {} statement {} ({} skipped, {} duplicates, {} warnings)",
                response.getImported(), p.bank(), response.getStatementId(), p.result().skipped().size(),
                response.getDuplicatesCount(), response.getWarningsCount());

        return new StatementImportResultDto(p.fileName(), Status.IMPORTED, p.bank().label, response.getAccountName(),
                response.getImported(), p.result().skipped(), response.getDuplicatesList(), response.getTransfersList(),
                response.getWarningsList(), null);
    }

    /** "" tells the engine the statement doesn't print this balance. */
    private static String amount(BigDecimal value) {
        return value == null ? "" : value.toPlainString();
    }

    private static String sha256(byte[] bytes) {
        try {
            return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(bytes));
        } catch (NoSuchAlgorithmException e) {
            throw new IllegalStateException("SHA-256 is always available on the JVM", e);
        }
    }
}
