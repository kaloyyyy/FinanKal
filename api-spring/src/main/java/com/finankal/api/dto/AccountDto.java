package com.finankal.api.dto;

import lombok.Data;
import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
public class AccountDto {
    private String id;
    private String name;
    private String type;
    private BigDecimal balance;
    private BigDecimal creditLimit;
    private Integer billingDay;
    private String statementShift;
    private LocalDateTime createdAt;
}
