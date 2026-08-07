package com.finankal.api.service;

import com.finankal.api.dto.AccountDto;
import com.finankal.api.finance.FinanceEngineGrpc;
import com.finankal.api.finance.FinanceProtos;
import com.finankal.api.mapper.AccountMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.util.List;
import java.util.UUID;

@Service
public class UserService {

    @Autowired
    private FinanceEngineGrpc.FinanceEngineBlockingStub financeEngineStub;

    @Autowired
    private AccountMapper accountMapper;

    public BigDecimal getUserTotalCredit(String userId) {
        FinanceProtos.GetUserTotalRequest request = FinanceProtos.GetUserTotalRequest.newBuilder()
                .setUserId(userId)
                .build();
        FinanceProtos.GetUserTotalResponse response = financeEngineStub.getUserTotalCredit(request);
        return new BigDecimal(response.getTotalCredit());
    }

    public BigDecimal getUserTotalDebit(String userId) {
        FinanceProtos.GetUserTotalRequest request = FinanceProtos.GetUserTotalRequest.newBuilder()
                .setUserId(userId)
                .build();
        FinanceProtos.GetUserTotalResponse response = financeEngineStub.getUserTotalDebit(request);
        return new BigDecimal(response.getTotalDebit());
    }

    public BigDecimal getUserNetWorth(String userId) {
        FinanceProtos.GetUserTotalRequest request = FinanceProtos.GetUserTotalRequest.newBuilder()
                .setUserId(userId)
                .build();
        FinanceProtos.GetUserTotalResponse response = financeEngineStub.getUserNetWorth(request);
        return new BigDecimal(response.getTotalBalance());
    }

    public List<AccountDto> getUserAccounts(UUID userId) {

        var request = FinanceProtos.GetUserAccountsRequest.newBuilder()
                .setUserId(userId.toString())
                .build();

        var response = financeEngineStub.getUserAccounts(request);

        return response.getAccountsList()
                .stream()
                .map(accountMapper::toDto)
                .toList();
    }
}
