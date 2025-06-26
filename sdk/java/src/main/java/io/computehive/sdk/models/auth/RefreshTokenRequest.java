package io.computehive.sdk.models.auth;


/**
 * Request for refreshing an access token.
 */
public class RefreshTokenRequest {
    
    /**
     * The refresh token to use for getting a new access token.
     */
    private String refreshToken;
    
    /**
     * Optional client identifier.
     */
    private String clientId;
    
    /**
     * Optional client secret.
     */
    private String clientSecret;
} 