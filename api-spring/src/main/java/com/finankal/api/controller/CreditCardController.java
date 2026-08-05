package com.finankal.api.controller;

import com.finankal.api.dto.*;
import com.finankal.api.service.CreditCardService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/credit-cards")
public class CreditCardController {

    private static final Logger logger = LoggerFactory.getLogger(CreditCardController.class);

    @Autowired
    private CreditCardService creditCardService;

    @PostMapping
    public ResponseEntity<CreateCreditCardResponseDto> createCreditCard(@RequestBody CreateCreditCardRequestDto request) {

        if (!request.isValid()) {
            logger.warn("POST /api/credit-cards - Either accountId or accountName must be provided.");
            return ResponseEntity.badRequest().build();
        }

        logger.info("POST /api/credit-cards - Creating credit card. accountId={}, accountName={}, creditLimit={}", request.getAccountId(), request.getAccountName(), request.getCreditLimit());

        CreateCreditCardResponseDto response = creditCardService.createCreditCard(request);

        return ResponseEntity.ok(response);
    }

    @PostMapping("/transactions")
    public ResponseEntity<RecordCreditCardTransactionResponseDto> recordTransaction(@RequestBody RecordCreditCardTransactionRequestDto request) {
        logger.info("POST /api/credit-cards/transactions - Recording purchase on card: {}", request.getCardId());
        RecordCreditCardTransactionResponseDto response = creditCardService.recordTransaction(request);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/payment")
    public ResponseEntity<PayCreditCardStatementResponseDto> payStatement(@RequestBody PayCreditCardStatementRequestDto request) {
        logger.info("POST /api/credit-cards/payment - Processing payment for statement: {}", request.getStatementId());
        PayCreditCardStatementResponseDto response = creditCardService.payStatement(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/statements/{statementId}")
    public ResponseEntity<CreditCardStatementDto> getStatement(@PathVariable String statementId) {
        logger.info("GET /api/credit-cards/statements/{} - Retrieving statement", statementId);
        CreditCardStatementDto response = creditCardService.getStatement(statementId);
        return ResponseEntity.ok(response);
    }
}

