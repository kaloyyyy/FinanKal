package com.finankal.api.service;

import com.finankal.api.dto.*;
import com.finankal.api.finance.FinanceEngineGrpc;
import com.finankal.api.finance.FinanceProtos;
import com.finankal.api.mapper.CreditCardMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class CreditCardService {

    private static final Logger logger = LoggerFactory.getLogger(CreditCardService.class);

    @Autowired
    private FinanceEngineGrpc.FinanceEngineBlockingStub financeEngineStub;

    @Autowired
    private CreditCardMapper creditCardMapper;

    public CreateCreditCardResponseDto createCreditCard(CreateCreditCardRequestDto requestDto) {

        if (!requestDto.isValid()) {
            throw new IllegalArgumentException("Either accountId or accountName must be provided.");
        }

        logger.info("Creating credit card. accountId={}, accountName={}, creditLimit={}", requestDto.getAccountId(), requestDto.getAccountName(), requestDto.getCreditLimit());

        FinanceProtos.CreateCreditCardRequest protoRequest = creditCardMapper.toCreateCreditCardProto(requestDto);

        FinanceProtos.CreateCreditCardResponse protoResponse = financeEngineStub.createCreditCard(protoRequest);

        CreateCreditCardResponseDto responseDto = creditCardMapper.toCreateCreditCardResponseDto(protoResponse);

        logger.info("Successfully created credit card with ID: {}", responseDto.getCreditCardId());

        return responseDto;
    }

    public RecordCreditCardTransactionResponseDto recordTransaction(RecordCreditCardTransactionRequestDto requestDto) {
        logger.info("Recording credit card transaction for card: {} amount: {} description: '{}'", requestDto.getCardId(), requestDto.getAmount(), requestDto.getDescription());

        FinanceProtos.RecordCreditCardTransactionRequest protoRequest = creditCardMapper.toRecordTransactionProto(requestDto);
        FinanceProtos.RecordCreditCardTransactionResponse protoResponse = financeEngineStub.recordCreditCardTransaction(protoRequest);

        RecordCreditCardTransactionResponseDto responseDto = creditCardMapper.toRecordTransactionResponseDto(protoResponse);
        logger.info("Successfully recorded transaction with ID: {}", responseDto.getTransactionId());

        return responseDto;
    }

    public PayCreditCardStatementResponseDto payStatement(PayCreditCardStatementRequestDto requestDto) {
        logger.info("Processing payment for statement: {} card: {} amount: {}", requestDto.getStatementId(), requestDto.getCardId(), requestDto.getAmount());

        FinanceProtos.PayCreditCardStatementRequest protoRequest = creditCardMapper.toPayStatementProto(requestDto);
        FinanceProtos.PayCreditCardStatementResponse protoResponse = financeEngineStub.payCreditCardStatement(protoRequest);

        PayCreditCardStatementResponseDto responseDto = creditCardMapper.toPayStatementResponseDto(protoResponse);
        logger.info("Successfully processed payment with transaction ID: {}", responseDto.getTransactionId());

        return responseDto;
    }

    public CreditCardStatementDto getStatement(String statementId) {
        logger.info("Retrieving credit card statement: {}", statementId);

        FinanceProtos.GetCreditCardStatementRequest protoRequest = FinanceProtos.GetCreditCardStatementRequest.newBuilder().setStatementId(statementId).build();

        FinanceProtos.GetCreditCardStatementResponse protoResponse = financeEngineStub.getCreditCardStatement(protoRequest);

        CreditCardStatementDto statementDto = creditCardMapper.toStatementDto(protoResponse);
        logger.info("Successfully retrieved statement for card: {}", statementDto.getCreditCardId());

        return statementDto;
    }
}

