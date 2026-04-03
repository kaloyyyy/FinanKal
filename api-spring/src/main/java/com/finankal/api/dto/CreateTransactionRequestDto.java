package com.finankal.api.dto;

import lombok.Data;
import java.util.List;

@Data
public class CreateTransactionRequestDto {
    private String description;
    private List<LedgerEntryDto> entries;
}
