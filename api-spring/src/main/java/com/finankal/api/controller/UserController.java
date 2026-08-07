package com.finankal.api.controller;

import com.finankal.api.dto.AccountDto;
import com.finankal.api.service.UserService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.math.BigDecimal;
import java.util.List;
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
}
