package com.finankal.api.service;

import com.finankal.api.dto.AccountDto;
import com.finankal.api.finance.FinanceEngineGrpc;
import com.finankal.api.finance.FinanceProtos;
import com.finankal.api.mapper.AccountMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.util.List;

@Service
public class AccountService {

    @Autowired
    private FinanceEngineGrpc.FinanceEngineBlockingStub financeEngineStub;

    @Autowired
    private AccountMapper accountMapper;

    public AccountDto getAccountSummary(String accountId) {
        FinanceProtos.GetAccountSummaryRequest request = FinanceProtos.GetAccountSummaryRequest.newBuilder()
                .setAccountId(accountId)
                .build();
        FinanceProtos.GetAccountSummaryResponse response = financeEngineStub.getAccountSummary(request);
        return accountMapper.toDto(response);
    }

    public BigDecimal getBalance(String accountId) {
        FinanceProtos.GetBalanceRequest request = FinanceProtos.GetBalanceRequest.newBuilder()
                .setAccountId(accountId)
                .build();
        FinanceProtos.GetBalanceResponse response = financeEngineStub.getBalance(request);
        return new BigDecimal(response.getBalance());
    }
}
