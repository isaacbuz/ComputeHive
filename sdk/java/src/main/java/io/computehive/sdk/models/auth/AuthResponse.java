package io.computehive.sdk.models.auth;


import java.time.LocalDateTime;

/**
 * Authentication response from the server.
 */
public class AuthResponse {
    
    /**
     * Access token for API authentication.
     */
    private String accessToken;
    
    /**
     * Refresh token for obtaining new access tokens.
     */
    private String refreshToken;
    
    /**
     * Token type (usually "Bearer").
     */
    private String tokenType;
    
    /**
     * Token expiration time in seconds.
     */
    private long expiresIn;
    
    /**
     * Token expiration timestamp.
     */
    private LocalDateTime expiresAt;
    
    /**
     * User profile information.
     */
    private UserProfile user;
    
    /**
     * Whether the user requires two-factor authentication.
     */
    private boolean requiresTwoFactor;
    
    /**
     * Session ID for tracking user sessions.
     */
    private String sessionId;
    
    /**
     * Whether the user is a new user (first time login).
     */
    private boolean isNewUser;
    
    /**
     * Account status information.
     */
    private AccountStatus accountStatus;
    
    /**
     * Account type information.
     */
    private AccountType accountType;
    
    /**
     * Subscription information.
     */
    private SubscriptionInfo subscription;
    
    /**
     * Account status enumeration.
     */
    public enum AccountStatus {
        ACTIVE,
        INACTIVE,
        SUSPENDED,
        PENDING_VERIFICATION,
        DELETED
    }
    
    /**
     * Account type enumeration.
     */
    public enum AccountType {
        INDIVIDUAL,
        BUSINESS,
        ENTERPRISE,
        RESELLER
    }
    
    /**
     * Subscription information.
     */
    public static class SubscriptionInfo {
        private String planId;
        
        private String planName;
        
        private String status;
        
        private LocalDateTime startDate;
        
        private LocalDateTime endDate;
        
        private boolean autoRenew;
        
        private double monthlyCost;
        
        private String currency;
        
        private int maxJobs;
        
        private int maxConcurrentJobs;
        
        private long maxStorageGB;
        
        private boolean prioritySupport;
    }
} 