package com.finankal.api.service;

import com.finankal.api.finance.FinanceEngineGrpc;
import com.finankal.api.finance.FinanceProtos;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;

@Service
public class UserService {

    @Autowired
    private FinanceEngineGrpc.FinanceEngineBlockingStub financeEngineStub;

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

    public BigDecimal getUserTotalBalance(String userId) {
        FinanceProtos.GetUserTotalRequest request = FinanceProtos.GetUserTotalRequest.newBuilder()
                .setUserId(userId)
                .build();
        FinanceProtos.GetUserTotalResponse response = financeEngineStub.getUserTotalBalance(request);
        return new BigDecimal(response.getTotalBalance());
    }
}
