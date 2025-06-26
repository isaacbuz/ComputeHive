package io.computehive.sdk.payments;

import com.google.gson.Gson;
import com.google.gson.reflect.TypeToken;
import io.computehive.sdk.ComputeHiveClient;
import io.computehive.sdk.exceptions.ComputeHiveException;
import okhttp3.*;

import java.io.IOException;
import java.lang.reflect.Type;
import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;

/**
 * Service for payment and billing operations.
 */
public class PaymentService {
    
    private static final MediaType JSON = MediaType.get("application/json; charset=utf-8");
    
    private final ComputeHiveClient client;
    private final Gson gson;
    private final OkHttpClient httpClient;
    
    public PaymentService(ComputeHiveClient client) {
        this.client = client;
        this.gson = new Gson();
        this.httpClient = client.getHttpClient();
    }
    
    /**
     * Get account balance.
     */
    public CompletableFuture<AccountBalance> getBalance() {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/payments/balance")
                        .get()
                        .build();
                
                try (Response response = httpClient.newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get balance: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, AccountBalance.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to get balance", e);
            }
        });
    }
    
    /**
     * Get transaction history.
     */
    public CompletableFuture<List<Transaction>> getTransactions(TransactionFilter filter) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                HttpUrl.Builder urlBuilder = HttpUrl.parse(client.getApiUrl() + "/api/v1/payments/transactions").newBuilder();
                
                if (filter != null) {
                    if (filter.getType() != null) {
                        urlBuilder.addQueryParameter("type", filter.getType());
                    }
                    if (filter.getStatus() != null) {
                        urlBuilder.addQueryParameter("status", filter.getStatus());
                    }
                    if (filter.getStartDate() != null) {
                        urlBuilder.addQueryParameter("start_date", filter.getStartDate().toString());
                    }
                    if (filter.getEndDate() != null) {
                        urlBuilder.addQueryParameter("end_date", filter.getEndDate().toString());
                    }
                    if (filter.getLimit() != null) {
                        urlBuilder.addQueryParameter("limit", filter.getLimit().toString());
                    }
                    if (filter.getOffset() != null) {
                        urlBuilder.addQueryParameter("offset", filter.getOffset().toString());
                    }
                }
                
                Request request = new Request.Builder()
                        .url(urlBuilder.build())
                        .get()
                        .build();
                
                try (Response response = httpClient.newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get transactions: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    Type listType = new TypeToken<List<Transaction>>(){}.getType();
                    return gson.fromJson(responseBody, listType);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Error getting transactions", e);
            }
        });
    }
    
    /**
     * Add funds to account.
     */
    public CompletableFuture<Transaction> addFunds(AddFundsRequest request) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                String json = gson.toJson(request);
                RequestBody body = RequestBody.create(json, JSON);
                
                Request httpRequest = new Request.Builder()
                        .url(client.getApiUrl() + "/payments/add-funds")
                        .post(body)
                        .build();
                
                try (Response response = httpClient.newCall(httpRequest).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to add funds: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, Transaction.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Error adding funds", e);
            }
        });
    }
    
    /**
     * Withdraw funds from account.
     */
    public CompletableFuture<Transaction> withdrawFunds(WithdrawFundsRequest request) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                String json = gson.toJson(request);
                RequestBody body = RequestBody.create(json, JSON);
                
                Request httpRequest = new Request.Builder()
                        .url(client.getApiUrl() + "/api/v1/payments/withdraw")
                        .post(body)
                        .build();
                
                try (Response response = httpClient.newCall(httpRequest).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to withdraw funds: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, Transaction.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Error withdrawing funds", e);
            }
        });
    }
    
    /**
     * Get payment methods.
     */
    public CompletableFuture<List<PaymentMethod>> getPaymentMethods() {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/payments/methods")
                        .get()
                        .build();
                
                try (Response response = httpClient.newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get payment methods: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    Type listType = new TypeToken<List<PaymentMethod>>(){}.getType();
                    return gson.fromJson(responseBody, listType);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Error getting payment methods", e);
            }
        });
    }
    
    /**
     * Add a payment method.
     */
    public CompletableFuture<PaymentMethod> addPaymentMethod(PaymentMethodRequest request) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                String json = gson.toJson(request);
                RequestBody body = RequestBody.create(json, JSON);
                
                Request httpRequest = new Request.Builder()
                        .url(client.getApiUrl() + "/payments/methods")
                        .post(body)
                        .build();
                
                try (Response response = httpClient.newCall(httpRequest).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to add payment method: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, PaymentMethod.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Error adding payment method", e);
            }
        });
    }
    
    /**
     * Remove a payment method.
     */
    public CompletableFuture<Void> removePaymentMethod(String methodId) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/payments/methods/" + methodId)
                        .delete()
                        .build();
                
                try (Response response = httpClient.newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to remove payment method: " + response.code());
                    }
                    return null;
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Error removing payment method", e);
            }
        });
    }
    
    /**
     * Get payment history.
     */
    public CompletableFuture<List<PaymentTransaction>> getPaymentHistory(Integer limit, Integer offset) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                HttpUrl.Builder urlBuilder = HttpUrl.parse(client.getApiUrl() + "/payments/history").newBuilder();
                
                if (limit != null) {
                    urlBuilder.addQueryParameter("limit", limit.toString());
                }
                if (offset != null) {
                    urlBuilder.addQueryParameter("offset", offset.toString());
                }
                
                Request request = new Request.Builder()
                        .url(urlBuilder.build())
                        .get()
                        .build();
                
                try (Response response = httpClient.newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get payment history: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    PaymentHistoryResponse historyResponse = gson.fromJson(responseBody, PaymentHistoryResponse.class);
                    return historyResponse.getTransactions();
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to get payment history", e);
            }
        });
    }
    
    /**
     * Get billing information.
     */
    public CompletableFuture<BillingInfo> getBillingInfo() {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/payments/billing")
                        .get()
                        .build();
                
                try (Response response = httpClient.newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get billing info: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, BillingInfo.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to get billing info", e);
            }
        });
    }
    
    /**
     * Update billing information.
     */
    public CompletableFuture<BillingInfo> updateBillingInfo(BillingInfo billingInfo) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                String json = gson.toJson(billingInfo);
                RequestBody body = RequestBody.create(json, JSON);
                
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/payments/billing")
                        .put(body)
                        .build();
                
                try (Response response = httpClient.newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to update billing info: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, BillingInfo.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to update billing info", e);
            }
        });
    }
    
    /**
     * Get invoice by ID.
     */
    public CompletableFuture<Invoice> getInvoice(String invoiceId) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/payments/invoices/" + invoiceId)
                        .get()
                        .build();
                
                try (Response response = httpClient.newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get invoice: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, Invoice.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to get invoice", e);
            }
        });
    }
    
    /**
     * List invoices.
     */
    public CompletableFuture<List<Invoice>> listInvoices(InvoiceStatus status, Integer limit, Integer offset) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                HttpUrl.Builder urlBuilder = HttpUrl.parse(client.getApiUrl() + "/payments/invoices").newBuilder();
                
                if (status != null) {
                    urlBuilder.addQueryParameter("status", status.name());
                }
                if (limit != null) {
                    urlBuilder.addQueryParameter("limit", limit.toString());
                }
                if (offset != null) {
                    urlBuilder.addQueryParameter("offset", offset.toString());
                }
                
                Request request = new Request.Builder()
                        .url(urlBuilder.build())
                        .get()
                        .build();
                
                try (Response response = httpClient.newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to list invoices: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    InvoiceListResponse invoiceListResponse = gson.fromJson(responseBody, InvoiceListResponse.class);
                    return invoiceListResponse.getInvoices();
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to list invoices", e);
            }
        });
    }
    
    /**
     * Transaction model.
     */
    public static class Transaction {
        private String id;
        private String type;
        private String status;
        private BigDecimal amount;
        private String currency;
        private String description;
        private LocalDateTime createdAt;
        private LocalDateTime completedAt;
        private Map<String, Object> metadata;
        
        // Getters and setters
        public String getId() { return id; }
        public void setId(String id) { this.id = id; }
        
        public String getType() { return type; }
        public void setType(String type) { this.type = type; }
        
        public String getStatus() { return status; }
        public void setStatus(String status) { this.status = status; }
        
        public BigDecimal getAmount() { return amount; }
        public void setAmount(BigDecimal amount) { this.amount = amount; }
        
        public String getCurrency() { return currency; }
        public void setCurrency(String currency) { this.currency = currency; }
        
        public String getDescription() { return description; }
        public void setDescription(String description) { this.description = description; }
        
        public LocalDateTime getCreatedAt() { return createdAt; }
        public void setCreatedAt(LocalDateTime createdAt) { this.createdAt = createdAt; }
        
        public LocalDateTime getCompletedAt() { return completedAt; }
        public void setCompletedAt(LocalDateTime completedAt) { this.completedAt = completedAt; }
        
        public Map<String, Object> getMetadata() { return metadata; }
        public void setMetadata(Map<String, Object> metadata) { this.metadata = metadata; }
    }
    
    /**
     * Transaction filter.
     */
    public static class TransactionFilter {
        private String type;
        private String status;
        private LocalDateTime startDate;
        private LocalDateTime endDate;
        private Integer limit;
        private Integer offset;
        
        // Getters and setters
        public String getType() { return type; }
        public void setType(String type) { this.type = type; }
        
        public String getStatus() { return status; }
        public void setStatus(String status) { this.status = status; }
        
        public LocalDateTime getStartDate() { return startDate; }
        public void setStartDate(LocalDateTime startDate) { this.startDate = startDate; }
        
        public LocalDateTime getEndDate() { return endDate; }
        public void setEndDate(LocalDateTime endDate) { this.endDate = endDate; }
        
        public Integer getLimit() { return limit; }
        public void setLimit(Integer limit) { this.limit = limit; }
        
        public Integer getOffset() { return offset; }
        public void setOffset(Integer offset) { this.offset = offset; }
    }
    
    /**
     * Add funds request.
     */
    public static class AddFundsRequest {
        private BigDecimal amount;
        private String currency;
        private String paymentMethodId;
        private String description;
        
        // Getters and setters
        public BigDecimal getAmount() { return amount; }
        public void setAmount(BigDecimal amount) { this.amount = amount; }
        
        public String getCurrency() { return currency; }
        public void setCurrency(String currency) { this.currency = currency; }
        
        public String getPaymentMethodId() { return paymentMethodId; }
        public void setPaymentMethodId(String paymentMethodId) { this.paymentMethodId = paymentMethodId; }
        
        public String getDescription() { return description; }
        public void setDescription(String description) { this.description = description; }
    }
    
    /**
     * Withdraw funds request.
     */
    public static class WithdrawFundsRequest {
        private BigDecimal amount;
        private String currency;
        private String destination;
        private String description;
        
        // Getters and setters
        public BigDecimal getAmount() { return amount; }
        public void setAmount(BigDecimal amount) { this.amount = amount; }
        
        public String getCurrency() { return currency; }
        public void setCurrency(String currency) { this.currency = currency; }
        
        public String getDestination() { return destination; }
        public void setDestination(String destination) { this.destination = destination; }
        
        public String getDescription() { return description; }
        public void setDescription(String description) { this.description = description; }
    }
    
    /**
     * Payment method model.
     */
    public static class PaymentMethod {
        private String id;
        private String type;
        private String name;
        private String last4;
        private String brand;
        private String expiryMonth;
        private String expiryYear;
        private boolean isDefault;
        private LocalDateTime createdAt;
        
        // Getters and setters
        public String getId() { return id; }
        public void setId(String id) { this.id = id; }
        
        public String getType() { return type; }
        public void setType(String type) { this.type = type; }
        
        public String getName() { return name; }
        public void setName(String name) { this.name = name; }
        
        public String getLast4() { return last4; }
        public void setLast4(String last4) { this.last4 = last4; }
        
        public String getBrand() { return brand; }
        public void setBrand(String brand) { this.brand = brand; }
        
        public String getExpiryMonth() { return expiryMonth; }
        public void setExpiryMonth(String expiryMonth) { this.expiryMonth = expiryMonth; }
        
        public String getExpiryYear() { return expiryYear; }
        public void setExpiryYear(String expiryYear) { this.expiryYear = expiryYear; }
        
        public boolean isDefault() { return isDefault; }
        public void setDefault(boolean isDefault) { this.isDefault = isDefault; }
        
        public LocalDateTime getCreatedAt() { return createdAt; }
        public void setCreatedAt(LocalDateTime createdAt) { this.createdAt = createdAt; }
    }
    
    /**
     * Payment method request.
     */
    public static class PaymentMethodRequest {
        private String type;
        private String cardNumber;
        private String expiryMonth;
        private String expiryYear;
        private String cvc;
        private String name;
        private String billingAddress;
        
        // Getters and setters
        public String getType() { return type; }
        public void setType(String type) { this.type = type; }
        
        public String getCardNumber() { return cardNumber; }
        public void setCardNumber(String cardNumber) { this.cardNumber = cardNumber; }
        
        public String getExpiryMonth() { return expiryMonth; }
        public void setExpiryMonth(String expiryMonth) { this.expiryMonth = expiryMonth; }
        
        public String getExpiryYear() { return expiryYear; }
        public void setExpiryYear(String expiryYear) { this.expiryYear = expiryYear; }
        
        public String getCvc() { return cvc; }
        public void setCvc(String cvc) { this.cvc = cvc; }
        
        public String getName() { return name; }
        public void setName(String name) { this.name = name; }
        
        public String getBillingAddress() { return billingAddress; }
        public void setBillingAddress(String billingAddress) { this.billingAddress = billingAddress; }
    }
    
    private static class PaymentHistoryResponse {
        private List<PaymentTransaction> transactions;
        private int total;
        
        public List<PaymentTransaction> getTransactions() { return transactions; }
        public void setTransactions(List<PaymentTransaction> transactions) { this.transactions = transactions; }
        public int getTotal() { return total; }
        public void setTotal(int total) { this.total = total; }
    }
    
    private static class InvoiceListResponse {
        private List<Invoice> invoices;
        private int total;
        
        public List<Invoice> getInvoices() { return invoices; }
        public void setInvoices(List<Invoice> invoices) { this.invoices = invoices; }
        public int getTotal() { return total; }
        public void setTotal(int total) { this.total = total; }
    }
    
    /**
     * Account balance model.
     */
    public static class AccountBalance {
        private double balance;
        private String currency;
        private double pendingCharges;
        private double availableCredit;
        private long lastUpdated;
        
        // Getters and setters
        public double getBalance() { return balance; }
        public void setBalance(double balance) { this.balance = balance; }
        public String getCurrency() { return currency; }
        public void setCurrency(String currency) { this.currency = currency; }
        public double getPendingCharges() { return pendingCharges; }
        public void setPendingCharges(double pendingCharges) { this.pendingCharges = pendingCharges; }
        public double getAvailableCredit() { return availableCredit; }
        public void setAvailableCredit(double availableCredit) { this.availableCredit = availableCredit; }
        public long getLastUpdated() { return lastUpdated; }
        public void setLastUpdated(long lastUpdated) { this.lastUpdated = lastUpdated; }
    }
    
    /**
     * Payment transaction model.
     */
    public static class PaymentTransaction {
        private String id;
        private String type;
        private double amount;
        private String currency;
        private TransactionStatus status;
        private String description;
        private long timestamp;
        private String paymentMethod;
        private String reference;
        
        // Getters and setters
        public String getId() { return id; }
        public void setId(String id) { this.id = id; }
        public String getType() { return type; }
        public void setType(String type) { this.type = type; }
        public double getAmount() { return amount; }
        public void setAmount(double amount) { this.amount = amount; }
        public String getCurrency() { return currency; }
        public void setCurrency(String currency) { this.currency = currency; }
        public TransactionStatus getStatus() { return status; }
        public void setStatus(TransactionStatus status) { this.status = status; }
        public String getDescription() { return description; }
        public void setDescription(String description) { this.description = description; }
        public long getTimestamp() { return timestamp; }
        public void setTimestamp(long timestamp) { this.timestamp = timestamp; }
        public String getPaymentMethod() { return paymentMethod; }
        public void setPaymentMethod(String paymentMethod) { this.paymentMethod = paymentMethod; }
        public String getReference() { return reference; }
        public void setReference(String reference) { this.reference = reference; }
    }
    
    /**
     * Transaction status enum.
     */
    public enum TransactionStatus {
        PENDING,
        COMPLETED,
        FAILED,
        CANCELLED,
        REFUNDED
    }
    
    /**
     * Billing info model.
     */
    public static class BillingInfo {
        private String billingEmail;
        private String billingAddress;
        private String city;
        private String state;
        private String country;
        private String postalCode;
        private String taxId;
        private String defaultCurrency;
        private boolean autoRecharge;
        private double rechargeThreshold;
        private double rechargeAmount;
        
        // Getters and setters
        public String getBillingEmail() { return billingEmail; }
        public void setBillingEmail(String billingEmail) { this.billingEmail = billingEmail; }
        public String getBillingAddress() { return billingAddress; }
        public void setBillingAddress(String billingAddress) { this.billingAddress = billingAddress; }
        public String getCity() { return city; }
        public void setCity(String city) { this.city = city; }
        public String getState() { return state; }
        public void setState(String state) { this.state = state; }
        public String getCountry() { return country; }
        public void setCountry(String country) { this.country = country; }
        public String getPostalCode() { return postalCode; }
        public void setPostalCode(String postalCode) { this.postalCode = postalCode; }
        public String getTaxId() { return taxId; }
        public void setTaxId(String taxId) { this.taxId = taxId; }
        public String getDefaultCurrency() { return defaultCurrency; }
        public void setDefaultCurrency(String defaultCurrency) { this.defaultCurrency = defaultCurrency; }
        public boolean isAutoRecharge() { return autoRecharge; }
        public void setAutoRecharge(boolean autoRecharge) { this.autoRecharge = autoRecharge; }
        public double getRechargeThreshold() { return rechargeThreshold; }
        public void setRechargeThreshold(double rechargeThreshold) { this.rechargeThreshold = rechargeThreshold; }
        public double getRechargeAmount() { return rechargeAmount; }
        public void setRechargeAmount(double rechargeAmount) { this.rechargeAmount = rechargeAmount; }
    }
    
    /**
     * Invoice model.
     */
    public static class Invoice {
        private String id;
        private String number;
        private double amount;
        private String currency;
        private InvoiceStatus status;
        private long dueDate;
        private long issuedDate;
        private long paidDate;
        private List<InvoiceItem> items;
        private String pdfUrl;
        
        // Getters and setters
        public String getId() { return id; }
        public void setId(String id) { this.id = id; }
        public String getNumber() { return number; }
        public void setNumber(String number) { this.number = number; }
        public double getAmount() { return amount; }
        public void setAmount(double amount) { this.amount = amount; }
        public String getCurrency() { return currency; }
        public void setCurrency(String currency) { this.currency = currency; }
        public InvoiceStatus getStatus() { return status; }
        public void setStatus(InvoiceStatus status) { this.status = status; }
        public long getDueDate() { return dueDate; }
        public void setDueDate(long dueDate) { this.dueDate = dueDate; }
        public long getIssuedDate() { return issuedDate; }
        public void setIssuedDate(long issuedDate) { this.issuedDate = issuedDate; }
        public long getPaidDate() { return paidDate; }
        public void setPaidDate(long paidDate) { this.paidDate = paidDate; }
        public List<InvoiceItem> getItems() { return items; }
        public void setItems(List<InvoiceItem> items) { this.items = items; }
        public String getPdfUrl() { return pdfUrl; }
        public void setPdfUrl(String pdfUrl) { this.pdfUrl = pdfUrl; }
    }
    
    /**
     * Invoice status enum.
     */
    public enum InvoiceStatus {
        DRAFT,
        PENDING,
        PAID,
        OVERDUE,
        CANCELLED
    }
    
    /**
     * Invoice item model.
     */
    public static class InvoiceItem {
        private String description;
        private int quantity;
        private double unitPrice;
        private double totalPrice;
        private String currency;
        
        // Getters and setters
        public String getDescription() { return description; }
        public void setDescription(String description) { this.description = description; }
        public int getQuantity() { return quantity; }
        public void setQuantity(int quantity) { this.quantity = quantity; }
        public double getUnitPrice() { return unitPrice; }
        public void setUnitPrice(double unitPrice) { this.unitPrice = unitPrice; }
        public double getTotalPrice() { return totalPrice; }
        public void setTotalPrice(double totalPrice) { this.totalPrice = totalPrice; }
        public String getCurrency() { return currency; }
        public void setCurrency(String currency) { this.currency = currency; }
    }
} 