package com.finankal.api.dto;

import lombok.Data;

import java.util.List;

@Data
public class TransactionPageDto {

    private List<TransactionDto> transactions;

    private int total;

    private int limit;

    private int offset;
}