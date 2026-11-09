package com.finankal.api.service;

import com.finankal.api.dto.AccountDto;
import com.finankal.api.dto.CreateInstallmentsRequestDto;
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
                .map(p -> new PayableDto(p.getName(), p.getAccountId(), p.getAccountName(), p.getDueDate(),
                        new BigDecimal(p.getAmount()),
                        p.getInstallmentNumber() == 0 ? null : p.getInstallmentNumber(),
                        p.getInstallments() == 0 ? null : p.getInstallments()))
                .toList();
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
