package com.finankal.api.controller;

import com.finankal.api.service.UserService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.math.BigDecimal;
import java.util.UUID;

@RestController
@RequestMapping("/api/users")
public class UserController {

    private static final Logger logger = LoggerFactory.getLogger(UserController.class);

    @Autowired
    private UserService userService;

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

    @GetMapping("/{id}/total-balance")
    public ResponseEntity<BigDecimal> getUserTotalBalance(@PathVariable UUID id) {
        logger.info("Fetching total balance for user: {}", id);
        BigDecimal total = userService.getUserTotalBalance(id.toString());
        logger.info("Successfully fetched total balance for user: {} - Total: {}", id, total);
        return ResponseEntity.ok(total);
    }
}
