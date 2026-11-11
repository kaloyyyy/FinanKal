package com.finankal.api.controller;

import com.finankal.api.dto.AccountDto;
import com.finankal.api.dto.CreateInstallmentsRequestDto;
import com.finankal.api.dto.LoanDetailsDto;
import com.finankal.api.dto.MoneyFlowDto;
import com.finankal.api.dto.PayableDto;
import com.finankal.api.dto.RecentTransactionDto;
import com.finankal.api.dto.TransactionDto;
import com.finankal.api.dto.TransactionPageDto;
import com.finankal.api.service.UserService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
@RequestMapping("/api/users")
public class UserController {

    private static final Logger logger = LoggerFactory.getLogger(UserController.class);

    @Autowired
    private UserService userService;

    @GetMapping("/{userId}/accounts")
    public ResponseEntity<List<AccountDto>> getUserAccounts(
            @PathVariable UUID userId
    ) {
        return ResponseEntity.ok(
                userService.getUserAccounts(userId)
        );
    }

    @GetMapping("/{id}/total-credit")
    public ResponseEntity<BigDecimal> getUserTotalCredit(@PathVariable UUID id) {
        logger.info("Fetching total credit for user: {}", id);
        BigDecimal total = userService.getUserTotalCredit(id.toString());
        logger.info("Successfully fetched total credit for user: {} - Total: {}", id, total);
        return ResponseEntity.ok(total);
    }

    @GetMapping("/{id}/total-debit")
    public ResponseEntity<BigDecimal> getUserTotalDebit(@PathVariable UUID id) {
        logger.info("Fetching total debit for user: {}", id);
        BigDecimal total = userService.getUserTotalDebit(id.toString());
        logger.info("Successfully fetched total debit for user: {} - Total: {}", id, total);
        return ResponseEntity.ok(total);
    }

    @GetMapping("/{id}/net-worth")
    public ResponseEntity<BigDecimal> getUserNetWorth(@PathVariable UUID id) {
        logger.info("Fetching net worth for user: {}", id);
        BigDecimal total = userService.getUserNetWorth(id.toString());
        logger.info("Successfully fetched net worth for user: {} - Total: {}", id, total);
        return ResponseEntity.ok(total);
    }

    @GetMapping("/{id}/transactions/recent")
    public List<RecentTransactionDto> getRecentTransactions(
            @PathVariable UUID id,
            @RequestParam(defaultValue = "5") int limit
    ) {

        int safeLimit = Math.clamp(limit, 1, 20);

        return userService.getRecentTransactions(
                id,
                safeLimit
        );
    }

    @GetMapping("/{userId}/transactions")
    public ResponseEntity<TransactionPageDto> getUserTransactions(
            @PathVariable UUID userId,
            @RequestParam(defaultValue = "20") int limit,
            @RequestParam(defaultValue = "0") int offset
    ) {

        limit = Math.clamp(limit, 1, 100);
        offset = Math.max(offset, 0);

        logger.info(
                "Getting transactions for user: {} limit={} offset={}",
                userId,
                limit,
                offset
        );

        return ResponseEntity.ok(
                userService.getUserTransactions(
                        userId,
                        limit,
                        offset
                )
        );
    }

    /**
     * Delete all of the user's transactions. Accounts are kept and their balances return to zero.
     *
     * DELETE /api/users/{userId}/transactions
     */
    @DeleteMapping("/{userId}/transactions")
    public ResponseEntity<Map<String, Long>> clearUserTransactions(@PathVariable UUID userId) {
        long deleted = userService.clearUserTransactions(userId);

        logger.warn("Cleared {} transactions for user: {}", deleted, userId);

        return ResponseEntity.ok(Map.of("deletedTransactions", deleted));
    }

    /** GET /api/users/{userId}/payables: unpaid loan and card installments, soonest first. */
    @GetMapping("/{userId}/payables")
    public ResponseEntity<List<PayableDto>> getPayables(@PathVariable UUID userId) {
        return ResponseEntity.ok(userService.getPayables(userId));
    }

    /** POST /api/users/{userId}/payables: adds a loan's installments; they then show up in GET payables. */
    @PostMapping("/{userId}/payables")
    public ResponseEntity<Map<String, String>> createInstallments(
            @PathVariable UUID userId,
            @RequestBody CreateInstallmentsRequestDto request
    ) {
        String scheduleId = userService.createInstallments(userId, request);
        logger.info("Added {} installments for user {} (schedule {})",
                request.installments() == null ? 0 : request.installments().size(), userId, scheduleId);
        return ResponseEntity.status(201).body(Map.of("scheduleId", scheduleId));
    }

    /** PUT /api/users/{userId}/paikot/{accountId}: the card that can pay this account's dues (forecast only). */
    @PutMapping("/{userId}/paikot/{accountId}")
    public ResponseEntity<Void> setPaikotLink(@PathVariable UUID userId, @PathVariable UUID accountId, @RequestBody PayableDto.Link link) {
        userService.setPaikotLink(userId, accountId, link);
        return ResponseEntity.noContent().build();
    }

    @PutMapping("/{userId}/payables/{scheduleId}")
    public ResponseEntity<Void> updateInstallments(@PathVariable UUID userId, @PathVariable UUID scheduleId, @RequestBody CreateInstallmentsRequestDto request) {
        userService.updateInstallments(userId,scheduleId,request); return ResponseEntity.noContent().build();
    }

    @DeleteMapping("/{userId}/payables/{scheduleId}")
    public ResponseEntity<Void> deleteInstallments(@PathVariable UUID userId,@PathVariable UUID scheduleId) {
        userService.deleteInstallments(userId,scheduleId); return ResponseEntity.noContent().build();
    }

    @PostMapping("/{userId}/payables/occurrences/{occurrenceId}/pay")
    public ResponseEntity<Void> payInstallment(@PathVariable UUID userId,@PathVariable UUID occurrenceId) {
        userService.payInstallment(userId,occurrenceId); return ResponseEntity.noContent().build();
    }

    @GetMapping("/{userId}/loans/{accountId}")
    public ResponseEntity<LoanDetailsDto> getLoan(@PathVariable UUID userId,@PathVariable UUID accountId) {
        return ResponseEntity.ok(userService.getLoan(userId,accountId));
    }

    @PutMapping("/{userId}/loans/{accountId}")
    public ResponseEntity<Void> updateLoan(@PathVariable UUID userId,@PathVariable UUID accountId,@RequestBody LoanDetailsDto request) {
        if (!accountId.toString().equals(request.accountId())) throw new IllegalArgumentException("Loan account does not match URL");
        userService.updateLoan(userId,request); return ResponseEntity.noContent().build();
    }

    @DeleteMapping("/{userId}/loans/{accountId}")
    public ResponseEntity<Void> deleteLoan(@PathVariable UUID userId,@PathVariable UUID accountId) {
        userService.deleteLoan(userId,accountId); return ResponseEntity.noContent().build();
    }

    /** GET /api/users/{userId}/money-flows?from=yyyy-MM-dd&to=yyyy-MM-dd (both optional, inclusive). */
    @GetMapping("/{userId}/money-flows")
    public ResponseEntity<List<MoneyFlowDto>> getMoneyFlows(
            @PathVariable UUID userId,
            @RequestParam(required = false) LocalDate from,
            @RequestParam(required = false) LocalDate to
    ) {
        return ResponseEntity.ok(userService.getMoneyFlows(userId, from, to));
    }
}
