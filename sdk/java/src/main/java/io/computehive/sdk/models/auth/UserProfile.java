package io.computehive.sdk.models.auth;


import java.time.LocalDateTime;
import java.util.List;

/**
 * User profile information.
 */
public class UserProfile {
    
    /**
     * Unique user identifier.
     */
    private String id;
    
    /**
     * User's email address.
     */
    private String email;
    
    /**
     * User's display name.
     */
    private String displayName;
    
    /**
     * User's first name.
     */
    private String firstName;
    
    /**
     * User's last name.
     */
    private String lastName;
    
    /**
     * User's avatar URL.
     */
    private String avatarUrl;
    
    /**
     * User's company/organization.
     */
    private String company;
    
    /**
     * User's role in the organization.
     */
    private String role;
    
    /**
     * User's phone number.
     */
    private String phone;
    
    /**
     * User's timezone.
     */
    private String timezone;
    
    /**
     * User's preferred language.
     */
    private String language;
    
    /**
     * User's account status.
     */
    private AccountStatus status;
    
    /**
     * User's account type.
     */
    private AccountType accountType;
    
    /**
     * User's subscription plan.
     */
    private String subscriptionPlan;
    
    /**
     * User's billing information.
     */
    private BillingInfo billingInfo;
    
    /**
     * User's preferences.
     */
    private UserPreferences preferences;
    
    /**
     * User's API keys.
     */
    private List<ApiKey> apiKeys;
    
    /**
     * User's security settings.
     */
    private SecuritySettings securitySettings;
    
    /**
     * Account creation timestamp.
     */
    private LocalDateTime createdAt;
    
    /**
     * Last profile update timestamp.
     */
    private LocalDateTime updatedAt;
    
    /**
     * Last login timestamp.
     */
    private LocalDateTime lastLoginAt;
    
    /**
     * Whether the user has verified their email.
     */
    private boolean emailVerified;
    
    /**
     * Whether two-factor authentication is enabled.
     */
    private boolean twoFactorEnabled;
    
    /**
     * User's account status.
     */
    public enum AccountStatus {
        ACTIVE,
        INACTIVE,
        SUSPENDED,
        PENDING_VERIFICATION,
        DELETED
    }
    
    /**
     * User's account type.
     */
    public enum AccountType {
        INDIVIDUAL,
        BUSINESS,
        ENTERPRISE,
        RESELLER
    }
    
    /**
     * Billing information for the user.
     */
    public static class BillingInfo {
        private String billingEmail;
        private String billingAddress;
        private String city;
        private String state;
        private String country;
        private String postalCode;
        private String taxId;
        private String paymentMethod;
        private String currency;
        private boolean autoRecharge;
        private double rechargeThreshold;
        private double rechargeAmount;
    }
    
    /**
     * User preferences and settings.
     */
    public static class UserPreferences {
        private String theme;
        private boolean emailNotifications;
        private boolean smsNotifications;
        private boolean pushNotifications;
        private String defaultRegion;
        private String defaultInstanceType;
        private boolean autoScaling;
        private int maxConcurrentJobs;
        private String jobTimeout;
        private boolean costOptimization;
        private String dataRetention;
    }
    
    /**
     * API key information.
     */
    public static class ApiKey {
        private String id;
        private String name;
        private String keyPrefix;
        private LocalDateTime createdAt;
        private LocalDateTime lastUsedAt;
        private boolean active;
        private List<String> permissions;
    }
    
    /**
     * Security settings for the user.
     */
    public static class SecuritySettings {
        private boolean requireMfa;
        private int sessionTimeout;
        private boolean allowApiAccess;
        private List<String> allowedIpRanges;
        private boolean auditLogging;
        private String passwordPolicy;
        private int failedLoginAttempts;
        private LocalDateTime lockoutUntil;
    }
} 