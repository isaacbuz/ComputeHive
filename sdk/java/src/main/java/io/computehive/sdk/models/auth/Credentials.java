package io.computehive.sdk.models.auth;


/**
 * User credentials for authentication.
 */
public class Credentials {
    
    /**
     * User's email address.
     */
    private String email;
    
    /**
     * User's password.
     */
    private String password;
    
    /**
     * Optional two-factor authentication code.
     */
    private String twoFactorCode;
    
    /**
     * Optional remember me flag.
     */
    private boolean rememberMe;
} 