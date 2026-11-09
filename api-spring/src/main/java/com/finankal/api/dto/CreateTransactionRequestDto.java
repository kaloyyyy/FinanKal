package com.finankal.api.dto;

import lombok.Data;

import java.time.LocalDate;
import java.util.List;

@Data
public class CreateTransactionRequestDto {

    private String description;

    private List<CreateTransactionEntryDto> entries;

    /** Day the transaction happened (yyyy-MM-dd). Optional; defaults to now. */
    private LocalDate date;
}
