package com.finankal.api.controller;

import com.finankal.api.dto.LedgerEntryDto;
import com.finankal.api.service.LedgerEntryService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api/ledger-entries")
public class LedgerEntryController {

    private static final Logger logger = LoggerFactory.getLogger(LedgerEntryController.class);

    @Autowired
    private LedgerEntryService ledgerEntryService;

    @GetMapping("/{accountId}")
    public ResponseEntity<List<LedgerEntryDto>> getLedgerEntries(@PathVariable String accountId) {
        logger.info("Fetching ledger entries for account: {}", accountId);
        List<LedgerEntryDto> entries = ledgerEntryService.getLedgerEntries(accountId);
        logger.info("Successfully fetched {} ledger entries for account: {}", entries.size(), accountId);
        return ResponseEntity.ok(entries);
    }
}
