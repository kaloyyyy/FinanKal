package com.finankal.api.controller;

import com.finankal.api.dto.CreateTransactionEntryDto;
import com.finankal.api.dto.CreateTransactionRequestDto;
import com.finankal.api.dto.TransactionDto;
import com.finankal.api.service.TransactionService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.UUID;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/api/transactions")
public class TransactionController {

    private static final Logger logger =
            LoggerFactory.getLogger(TransactionController.class);

    @Autowired
    private TransactionService transactionService;


    /**
     * Create a new double-entry transaction.
     *
     * POST /api/transactions
     */
    @PostMapping
    public ResponseEntity<TransactionDto> createTransaction(
            @RequestBody CreateTransactionRequestDto request
    ) {

        String accountIds = request.getEntries()
                .stream()
                .map(CreateTransactionEntryDto::getAccountId)
                .collect(Collectors.joining(", "));

        logger.info(
                "Creating new transaction with description: '{}' involving accounts: {}",
                request.getDescription(),
                accountIds
        );

        TransactionDto transaction =
                transactionService.createTransaction(request);

        logger.info(
                "Successfully created transaction with ID: {}",
                transaction.getId()
        );

        return ResponseEntity.ok(transaction);
    }


    /**
     * Get a specific transaction.
     *
     * GET /api/transactions/{id}
     */
    @GetMapping("/{id}")
    public ResponseEntity<TransactionDto> getTransaction(
            @PathVariable UUID id
    ) {

        logger.info(
                "Getting transaction: {}",
                id
        );

        TransactionDto transaction =
                transactionService.getTransaction(id);

        return ResponseEntity.ok(transaction);
    }


    /**
     * Edit a transaction's description, date and its entries' accounts and amounts.
     *
     * PUT /api/transactions/{id}?userId=...
     */
    @PutMapping("/{id}")
    public ResponseEntity<Void> updateTransaction(
            @PathVariable UUID id,
            @RequestParam UUID userId,
            @RequestBody CreateTransactionRequestDto request
    ) {
        logger.info("Updating transaction: {}", id);
        transactionService.updateTransaction(userId.toString(), id, request);
        return ResponseEntity.noContent().build();
    }


    /**
     * Delete a transaction.
     *
     * DELETE /api/transactions/{id}?userId=...
     */
    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteTransaction(
            @PathVariable UUID id,
            @RequestParam UUID userId
    ) {
        logger.info("Deleting transaction: {}", id);
        transactionService.deleteTransaction(userId.toString(), id);
        return ResponseEntity.noContent().build();
    }
}