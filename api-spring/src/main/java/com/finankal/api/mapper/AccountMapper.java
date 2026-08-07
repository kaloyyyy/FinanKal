package com.finankal.api.mapper;

import com.finankal.api.dto.AccountDto;
import com.finankal.api.finance.FinanceProtos;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;

@Component
public class AccountMapper {

    public AccountDto toDto(FinanceProtos.GetAccountSummaryResponse response) {
        return toDto(response.getAccount());
    }

    public AccountDto toDto(FinanceProtos.Account account) {
        AccountDto dto = new AccountDto();
        dto.setId(account.getId());
        dto.setName(account.getName());
        dto.setType(account.getType());
        dto.setBalance(new BigDecimal(account.getBalance()));

        Instant instant = Instant.ofEpochSecond(
                account.getCreatedAt().getSeconds(),
                account.getCreatedAt().getNanos()
        );

        dto.setCreatedAt(
                LocalDateTime.ofInstant(instant, ZoneOffset.UTC)
        );

        return dto;
    }
}
