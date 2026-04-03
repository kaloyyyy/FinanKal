package com.finankal.api.mapper;

import com.finankal.api.dto.AccountDto;
import com.finankal.api.finance.FinanceProtos;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.time.OffsetDateTime;

@Component
public class AccountMapper {

    public AccountDto toDto(FinanceProtos.GetAccountSummaryResponse response) {
        AccountDto dto = new AccountDto();
        dto.setId(response.getAccountId());
        dto.setName(response.getName());
        dto.setType(response.getType());
        dto.setBalance(new BigDecimal(response.getBalance()));
        
        // Parse RFC3339 timestamp with timezone
        OffsetDateTime odt = OffsetDateTime.parse(response.getCreatedAt());
        LocalDateTime localDateTime = odt.toLocalDateTime();
        dto.setCreatedAt(localDateTime);
        
        return dto;
    }
}
