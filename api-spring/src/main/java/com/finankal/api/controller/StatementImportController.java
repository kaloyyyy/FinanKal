package com.finankal.api.controller;

import com.finankal.api.dto.BankStatementDto;
import com.finankal.api.dto.StatementImportResultDto;
import com.finankal.api.service.BankStatementParser.Bank;
import com.finankal.api.service.StatementImportService;
import com.finankal.api.service.StatementImportService.Upload;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/api/bank-statements")
public class StatementImportController {

    @Autowired
    private StatementImportService statementImportService;

    /** GET /api/bank-statements?userId=...: the user's imported statements, oldest period first. */
    @GetMapping
    public ResponseEntity<List<BankStatementDto>> listStatements(@RequestParam("userId") UUID userId) {
        return ResponseEntity.ok(statementImportService.listStatements(userId));
    }

    /**
     * Import one or more PDF bank / e-wallet statements. Returns one result per file.
     *
     * POST /api/bank-statements/import  (multipart/form-data)
     * files (repeated), userId,
     * bank (MARIBANK|BDO|MAYA|GCASH|GOTYME|BPI, optional: detected from each statement),
     * sourceAccountId (optional: found by the statement's account number, or created),
     * password (optional, used for every file),
     * replace (optional, default false: true re-imports a statement already on file, replacing its rows)
     */
    @PostMapping(value = "/import", consumes = MediaType.MULTIPART_FORM_DATA_VALUE)
    public ResponseEntity<List<StatementImportResultDto>> importStatements(
            @RequestParam("files") List<MultipartFile> files,
            @RequestParam("userId") String userId,
            @RequestParam(value = "bank", required = false) Bank bank,
            @RequestParam(value = "sourceAccountId", required = false) String sourceAccountId,
            @RequestParam(value = "password", required = false) String password,
            @RequestParam(value = "replace", required = false, defaultValue = "false") boolean replace
    ) throws IOException {
        if (files.isEmpty()) {
            throw new IllegalArgumentException("Choose at least one statement file");
        }
        List<Upload> uploads = new ArrayList<>();
        for (MultipartFile file : files) {
            uploads.add(new Upload(file.getOriginalFilename(), file.getBytes()));
        }
        String accountId = sourceAccountId == null || sourceAccountId.isBlank() ? null : sourceAccountId;
        return ResponseEntity.ok(statementImportService.importStatements(uploads, password, bank, userId, accountId, replace));
    }
}
