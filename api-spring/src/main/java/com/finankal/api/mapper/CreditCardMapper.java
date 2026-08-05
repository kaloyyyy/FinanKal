package com.finankal.api.mapper;

import com.finankal.api.dto.*;
import com.finankal.api.finance.FinanceProtos;
import com.google.protobuf.Timestamp;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.format.DateTimeFormatter;

@Component
public class CreditCardMapper {

    public FinanceProtos.CreateCreditCardRequest toCreateCreditCardProto(CreateCreditCardRequestDto dto) {

        FinanceProtos.CreateCreditCardRequest.Builder builder =
                FinanceProtos.CreateCreditCardRequest.newBuilder()
                        .setCreditLimit(dto.getCreditLimit().toPlainString())
                        .setBillingDay(dto.getBillingDay())
                        .setPaymentDueDays(dto.getPaymentDueDays());

        if (dto.hasAccountId()) {
            builder.setAccountId(dto.getAccountId());
        }

        if (dto.hasAccountName()) {
            builder.setAccountName(dto.getAccountName());
        }

        return builder.build();
    }

    public CreateCreditCardResponseDto toCreateCreditCardResponseDto(FinanceProtos.CreateCreditCardResponse response) {
        CreateCreditCardResponseDto dto = new CreateCreditCardResponseDto();
        dto.setCreditCardId(response.getCreditCardId());
        return dto;
    }

    public FinanceProtos.RecordCreditCardTransactionRequest toRecordTransactionProto(RecordCreditCardTransactionRequestDto dto) {
        Timestamp timestamp = Timestamp.newBuilder()
                .setSeconds(dto.getPurchaseDate().getEpochSecond())
                .setNanos(dto.getPurchaseDate().getNano())
                .build();

        FinanceProtos.RecordCreditCardTransactionRequest.Builder builder = FinanceProtos.RecordCreditCardTransactionRequest.newBuilder()
                .setCardId(dto.getCardId())
                .setAmount(dto.getAmount().toString())
                .setPurchaseDate(timestamp);

        if (dto.getDescription() != null) {
            builder.setDescription(dto.getDescription());
        }

        return builder.build();
    }

    public RecordCreditCardTransactionResponseDto toRecordTransactionResponseDto(FinanceProtos.RecordCreditCardTransactionResponse response) {
        RecordCreditCardTransactionResponseDto dto = new RecordCreditCardTransactionResponseDto();
        dto.setTransactionId(response.getTransactionId());
        return dto;
    }

    public FinanceProtos.PayCreditCardStatementRequest toPayStatementProto(PayCreditCardStatementRequestDto dto) {
        FinanceProtos.PayCreditCardStatementRequest.Builder builder = FinanceProtos.PayCreditCardStatementRequest.newBuilder()
                .setStatementId(dto.getStatementId())
                .setCardId(dto.getCardId())
                .setPaymentAccountId(dto.getPaymentAccountId())
                .setAmount(dto.getAmount().toString());

        if (dto.getDescription() != null) {
            builder.setDescription(dto.getDescription());
        }

        return builder.build();
    }

    public PayCreditCardStatementResponseDto toPayStatementResponseDto(FinanceProtos.PayCreditCardStatementResponse response) {
        PayCreditCardStatementResponseDto dto = new PayCreditCardStatementResponseDto();
        dto.setTransactionId(response.getTransactionId());
        return dto;
    }

    public CreditCardStatementDto toStatementDto(FinanceProtos.GetCreditCardStatementResponse response) {
        CreditCardStatementDto dto = new CreditCardStatementDto();
        dto.setStatementId(response.getStatementId());
        dto.setCreditCardId(response.getCreditCardId());
        dto.setStartDate(LocalDate.parse(response.getStartDate(), DateTimeFormatter.ISO_DATE));
        dto.setEndDate(LocalDate.parse(response.getEndDate(), DateTimeFormatter.ISO_DATE));
        dto.setStatementDate(LocalDate.parse(response.getStatementDate(), DateTimeFormatter.ISO_DATE));
        dto.setDueDate(LocalDate.parse(response.getDueDate(), DateTimeFormatter.ISO_DATE));
        dto.setTotalAmount(new BigDecimal(response.getTotalAmount()));
        dto.setStatus(response.getStatus());
        return dto;
    }
}

