package com.finankal.api.controller;

import com.finankal.api.dto.CreateTransactionRequestDto;
import com.finankal.api.dto.LedgerEntryDto;
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
     */
    @PostMapping
    public ResponseEntity<TransactionDto> createTransaction(
            @RequestBody CreateTransactionRequestDto request
    ) {

        String accountIds = request.getEntries()
                .stream()
                .map(LedgerEntryDto::getAccountId)
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
     * <p>
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
}