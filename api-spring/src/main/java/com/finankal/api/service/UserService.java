package com.finankal.api.service;

import com.finankal.api.dto.AccountDto;
import com.finankal.api.dto.CreateInstallmentsRequestDto;
import com.finankal.api.dto.LoanDetailsDto;
import com.finankal.api.dto.MoneyFlowDto;
import com.finankal.api.dto.PayableDto;
import com.finankal.api.dto.RecentTransactionDto;
import com.finankal.api.dto.TransactionDto;
import com.finankal.api.dto.TransactionPageDto;
import com.finankal.api.finance.FinanceEngineGrpc;
import com.finankal.api.finance.FinanceProtos;
import com.finankal.api.mapper.AccountMapper;
import com.finankal.api.mapper.TransactionMapper;
import io.grpc.StatusRuntimeException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.UUID;

@Service
public class UserService {

    @Autowired
    private FinanceEngineGrpc.FinanceEngineBlockingStub financeEngineStub;

    @Autowired
    private AccountMapper accountMapper;
    @Autowired
    private TransactionMapper transactionMapper;

    /** Deletes all of the user's transactions, keeping accounts. Returns how many were deleted. */
    public long clearUserTransactions(UUID userId) {
        FinanceProtos.ClearUserTransactionsRequest request =
                FinanceProtos.ClearUserTransactionsRequest.newBuilder()
                        .setUserId(userId.toString())
                        .build();

        return financeEngineStub.clearUserTransactions(request).getDeletedTransactions();
    }

    /** Money moved between the user's accounts, largest first. from / to are inclusive days, null = no limit. */
    public List<MoneyFlowDto> getMoneyFlows(UUID userId, LocalDate from, LocalDate to) {
        FinanceProtos.GetMoneyFlowsRequest request = FinanceProtos.GetMoneyFlowsRequest.newBuilder()
                .setUserId(userId.toString())
                .setFrom(from == null ? "" : from.toString())
                .setTo(to == null ? "" : to.toString())
                .build();
        return financeEngineStub.getMoneyFlows(request).getFlowsList().stream()
                .map(f -> new MoneyFlowDto(f.getFromAccountId(), f.getToAccountId(), new BigDecimal(f.getAmount()), f.getTransactions()))
                .toList();
    }

    /** Unpaid loan and card installments, soonest first. */
    public List<PayableDto> getPayables(UUID userId) {
        FinanceProtos.GetPayablesRequest request = FinanceProtos.GetPayablesRequest.newBuilder()
                .setUserId(userId.toString())
                .build();
        return financeEngineStub.getPayables(request).getPayablesList().stream()
                .map(p -> new PayableDto(p.getName(), p.getAccountId(), p.getAccountName(), p.getDueDate(), p.getScheduledDate(),
                        new BigDecimal(p.getAmount()),
                        p.getInstallmentNumber() == 0 ? null : p.getInstallmentNumber(),
                        p.getInstallments() == 0 ? null : p.getInstallments(),
                        p.getSourceAccountId().isBlank() ? null : p.getSourceAccountId(),
                        p.getScheduleId().isBlank() ? null : p.getScheduleId(),
                        p.getOccurrenceId().isBlank() ? null : p.getOccurrenceId(),
                        p.getPaikotCardId().isBlank() ? null : new PayableDto.Paikot(p.getPaikotCardId(), p.getPaikotCardName(),
                                p.getPaikotPayFrom(), p.getPaikotPayBy(), p.getPaikotDueDate(), new BigDecimal(p.getPaikotFee())),
                        p.getPrincipal().isBlank() ? null : new BigDecimal(p.getPrincipal())))
                .toList();
    }

    public void setPaikotLink(UUID userId, UUID accountId, PayableDto.Link link) {
        var request = FinanceProtos.SetPaikotLinkRequest.newBuilder()
                .setUserId(userId.toString()).setAccountId(accountId.toString())
                .setCardId(link.cardId() == null ? "" : link.cardId())
                .setFeePercent(link.feePercent() == null ? "" : link.feePercent().toPlainString())
                .build();
        try { financeEngineStub.setPaikotLink(request); }
        catch (StatusRuntimeException e) { throw apiError(e); }
    }

    /** Records a loan's repayment schedule and returns its id. Rejected schedules throw IllegalArgumentException (400). */
    public String createInstallments(UUID userId, CreateInstallmentsRequestDto dto) {
        FinanceProtos.CreateInstallmentsRequest.Builder request = FinanceProtos.CreateInstallmentsRequest.newBuilder()
                .setUserId(userId.toString())
                .setName(dto.name() == null ? "" : dto.name())
                .setSourceAccountId(dto.sourceAccountId() == null ? "" : dto.sourceAccountId())
                .setAccountId(dto.accountId() == null ? "" : dto.accountId());
        if (dto.installments() != null) {
            for (CreateInstallmentsRequestDto.Installment i : dto.installments()) {
                request.addInstallments(FinanceProtos.Installment.newBuilder()
                        .setDueDate(i.dueDate() == null ? "" : i.dueDate())
                        .setAmount(i.amount() == null ? "" : i.amount().toPlainString()));
            }
        }
        try {
            return financeEngineStub.createInstallments(request.build()).getScheduleId();
        } catch (StatusRuntimeException e) {
            if (e.getStatus().getCode() == io.grpc.Status.Code.INVALID_ARGUMENT) {
                throw new IllegalArgumentException(e.getStatus().getDescription());
            }
            throw e;
        }
    }

    public void updateInstallments(UUID userId, UUID scheduleId, CreateInstallmentsRequestDto dto) {
        FinanceProtos.UpdateInstallmentsRequest.Builder b = FinanceProtos.UpdateInstallmentsRequest.newBuilder()
                .setUserId(userId.toString()).setScheduleId(scheduleId.toString())
                .setName(dto.name() == null ? "" : dto.name())
                .setSourceAccountId(dto.sourceAccountId() == null ? "" : dto.sourceAccountId())
                .setFirstInstallmentNumber(dto.firstInstallmentNumber());
        if (dto.installments() != null) for (var i : dto.installments()) b.addInstallments(FinanceProtos.Installment.newBuilder()
                .setDueDate(i.dueDate() == null ? "" : i.dueDate()).setAmount(i.amount() == null ? "" : i.amount().toPlainString()));
        try { financeEngineStub.updateInstallments(b.build()); }
        catch (StatusRuntimeException e) { throw apiError(e); }
    }

    public void deleteInstallments(UUID userId, UUID scheduleId) {
        try { financeEngineStub.deleteInstallments(FinanceProtos.DeleteInstallmentsRequest.newBuilder().setUserId(userId.toString()).setScheduleId(scheduleId.toString()).build()); }
        catch (StatusRuntimeException e) { throw apiError(e); }
    }

    public void payInstallment(UUID userId, UUID occurrenceId) {
        try { financeEngineStub.payInstallment(FinanceProtos.PayInstallmentRequest.newBuilder().setUserId(userId.toString()).setOccurrenceId(occurrenceId.toString()).build()); }
        catch (StatusRuntimeException e) { throw apiError(e); }
    }

    public LoanDetailsDto getLoan(UUID userId, UUID accountId) {
        FinanceProtos.LoanDetails d;
        try { d = financeEngineStub.getLoan(FinanceProtos.GetLoanRequest.newBuilder().setUserId(userId.toString()).setAccountId(accountId.toString()).build()); }
        catch (StatusRuntimeException e) { throw apiError(e); }
        return new LoanDetailsDto(d.getAccountId(),d.getName(),d.getBookedOn(),new BigDecimal(d.getDailyRate()),d.getPrincipalOn(),new BigDecimal(d.getPrincipal()));
    }

    public void updateLoan(UUID userId, LoanDetailsDto dto) {
        var d=FinanceProtos.LoanDetails.newBuilder().setAccountId(dto.accountId()).setName(dto.name())
                .setBookedOn(dto.bookedOn()).setDailyRate(dto.dailyRate().toPlainString())
                .setPrincipalOn(dto.principalOn()).setPrincipal(dto.principal().toPlainString()).build();
        try { financeEngineStub.updateLoan(FinanceProtos.UpdateLoanRequest.newBuilder().setUserId(userId.toString()).setLoan(d).build()); }
        catch (StatusRuntimeException e) { throw apiError(e); }
    }

    public void deleteLoan(UUID userId, UUID accountId) {
        try { financeEngineStub.deleteLoan(FinanceProtos.DeleteLoanRequest.newBuilder().setUserId(userId.toString()).setAccountId(accountId.toString()).build()); }
        catch (StatusRuntimeException e) { throw apiError(e); }
    }

    private static RuntimeException apiError(StatusRuntimeException e) {
        return switch (e.getStatus().getCode()) {
            case INVALID_ARGUMENT, NOT_FOUND, FAILED_PRECONDITION -> new IllegalArgumentException(e.getStatus().getDescription());
            default -> e;
        };
    }

    public BigDecimal getUserTotalCredit(String userId) {
        FinanceProtos.GetUserTotalRequest request = FinanceProtos.GetUserTotalRequest.newBuilder()
                .setUserId(userId)
                .build();
        FinanceProtos.GetUserTotalResponse response = financeEngineStub.getUserTotalCredit(request);
        return new BigDecimal(response.getTotalCredit());
    }

    public BigDecimal getUserTotalDebit(String userId) {
        FinanceProtos.GetUserTotalRequest request = FinanceProtos.GetUserTotalRequest.newBuilder()
                .setUserId(userId)
                .build();
        FinanceProtos.GetUserTotalResponse response = financeEngineStub.getUserTotalDebit(request);
        return new BigDecimal(response.getTotalDebit());
    }

    public BigDecimal getUserNetWorth(String userId) {
        FinanceProtos.GetUserTotalRequest request = FinanceProtos.GetUserTotalRequest.newBuilder()
                .setUserId(userId)
                .build();
        FinanceProtos.GetUserTotalResponse response = financeEngineStub.getUserNetWorth(request);
        return new BigDecimal(response.getTotalBalance());
    }

    public List<AccountDto> getUserAccounts(UUID userId) {

        var request = FinanceProtos.GetUserAccountsRequest.newBuilder()
                .setUserId(userId.toString())
                .build();

        var response = financeEngineStub.getUserAccounts(request);

        return response.getAccountsList()
                .stream()
                .map(accountMapper::toDto)
                .toList();
    }

    public List<RecentTransactionDto> getRecentTransactions(
            UUID userId,
            int limit
    ) {

        var request =
                FinanceProtos.GetRecentTransactionsRequest.newBuilder()
                        .setUserId(userId.toString())
                        .setLimit(limit)
                        .build();

        var response =
                financeEngineStub.getRecentTransactions(request);

        return response.getTransactionsList()
                .stream()
                .map(transaction ->
                        new RecentTransactionDto(
                                UUID.fromString(transaction.getId()),
                                transaction.getDescription(),
                                new BigDecimal(transaction.getAmount()),
                                transaction.getDirection(),
                                transaction.getAccountName(),
                                OffsetDateTime.parse(transaction.getCreatedAt())
                        )
                )
                .toList();
    }

    /**
     * Get all transactions for a user.
     */
    public TransactionPageDto getUserTransactions(
            UUID userId,
            int limit,
            int offset
    ) {

        FinanceProtos.GetUserTransactionsRequest request =
                FinanceProtos.GetUserTransactionsRequest.newBuilder()
                        .setUserId(userId.toString())
                        .setLimit(limit)
                        .setOffset(offset)
                        .build();

        FinanceProtos.GetUserTransactionsResponse response =
                financeEngineStub.getUserTransactions(request);

        TransactionPageDto dto = new TransactionPageDto();

        dto.setTransactions(
                response.getTransactionsList()
                        .stream()
                        .map(transactionMapper::toDto)
                        .toList()
        );

        dto.setTotal(response.getTotal());
        dto.setLimit(limit);
        dto.setOffset(offset);

        return dto;
    }
}
