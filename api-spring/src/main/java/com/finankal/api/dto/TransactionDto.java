package com.finankal.api.dto;

import lombok.Data;
import java.util.List;

@Data
public class TransactionDto {
    private String id;
    private String description;
    private List<LedgerEntryDto> entries;
}
