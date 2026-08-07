package com.finankal.api.controller;

import com.finankal.api.dto.AccountDto;
import com.finankal.api.service.AccountService;
import java.math.BigDecimal;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/accounts")
public class AccountController {

    private static final Logger logger = LoggerFactory.getLogger(AccountController.class);

    @Autowired
    private AccountService accountService;

    @GetMapping("/{id}")
    public ResponseEntity<AccountDto> getAccount(@PathVariable String id) {
        logger.info("Fetching account summary for account: {}", id);
        AccountDto account = accountService.getAccountSummary(id);
        logger.info("Successfully fetched account summary for account: {}", id);
        return ResponseEntity.ok(account);
    }

    @GetMapping("/{id}/balance")
    public ResponseEntity<BigDecimal> getBalance(@PathVariable String id) {
        logger.info("Fetching balance for account: {}", id);
        BigDecimal balance = accountService.getBalance(id);
        logger.info("Successfully fetched balance for account: {} - Balance: {}", id, balance);
        return ResponseEntity.ok(balance);
    }
}
