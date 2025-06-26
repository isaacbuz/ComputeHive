package io.computehive.sdk.auth;

import com.google.gson.Gson;
import io.computehive.sdk.ComputeHiveClient;
import io.computehive.sdk.exceptions.ComputeHiveException;
import io.computehive.sdk.models.auth.AuthResponse;
import io.computehive.sdk.models.auth.Credentials;
import io.computehive.sdk.models.auth.RefreshTokenRequest;
import io.computehive.sdk.models.auth.UserProfile;
import okhttp3.*;
import okhttp3.MediaType;

import java.io.IOException;
import java.util.concurrent.CompletableFuture;

/**
 * Service for handling authentication operations.
 */
public class AuthService {
    
    private static final MediaType JSON = MediaType.get("application/json; charset=utf-8");
    
    private final ComputeHiveClient client;
    private final Gson gson;
    
    public AuthService(ComputeHiveClient client) {
        this.client = client;
        this.gson = client.getGson();
    }
    
    /**
     * Authenticate with email and password.
     * 
     * @param credentials User credentials
     * @return CompletableFuture containing the authentication response
     */
    public CompletableFuture<AuthResponse> authenticate(Credentials credentials) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                String json = gson.toJson(credentials);
                RequestBody body = RequestBody.create(json, JSON);
                
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/auth/login")
                        .post(body)
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Authentication failed: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, AuthResponse.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Authentication request failed", e);
            }
        });
    }
    
    /**
     * Register a new user account.
     * 
     * @param credentials User registration credentials
     * @return CompletableFuture containing the authentication response
     */
    public CompletableFuture<AuthResponse> register(Credentials credentials) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                String json = gson.toJson(credentials);
                RequestBody body = RequestBody.create(json, JSON);
                
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/auth/register")
                        .post(body)
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Registration failed: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, AuthResponse.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Registration request failed", e);
            }
        });
    }
    
    /**
     * Refresh the access token using a refresh token.
     * 
     * @param refreshToken The refresh token
     * @return CompletableFuture containing the new authentication response
     */
    public CompletableFuture<AuthResponse> refreshToken(String refreshToken) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                RefreshTokenRequest requestBody = RefreshTokenRequest.builder()
                        .refreshToken(refreshToken)
                        .build();
                
                String json = gson.toJson(requestBody);
                RequestBody body = RequestBody.create(json, JSON);
                
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/auth/refresh")
                        .post(body)
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Token refresh failed: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, AuthResponse.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Token refresh request failed", e);
            }
        });
    }
    
    /**
     * Get the current user's profile.
     * 
     * @return CompletableFuture containing the user profile
     */
    public CompletableFuture<UserProfile> getProfile() {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/auth/profile")
                        .get()
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get profile: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, UserProfile.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Profile request failed", e);
            }
        });
    }
    
    /**
     * Update the current user's profile.
     * 
     * @param profile The updated profile
     * @return CompletableFuture containing the updated user profile
     */
    public CompletableFuture<UserProfile> updateProfile(UserProfile profile) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                String json = gson.toJson(profile);
                RequestBody body = RequestBody.create(json, JSON);
                
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/auth/profile")
                        .put(body)
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to update profile: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, UserProfile.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Profile update request failed", e);
            }
        });
    }
    
    /**
     * Logout the current user.
     * 
     * @return CompletableFuture that completes when logout is successful
     */
    public CompletableFuture<Void> logout() {
        return CompletableFuture.runAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/auth/logout")
                        .post(RequestBody.create("", null))
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        log.warn("Logout request failed: {}", response.code());
                    }
                }
            } catch (IOException e) {
                log.warn("Logout request failed", e);
            }
        });
    }
    
    /**
     * Request a password reset.
     * 
     * @param email The user's email address
     * @return CompletableFuture that completes when the request is sent
     */
    public CompletableFuture<Void> requestPasswordReset(String email) {
        return CompletableFuture.runAsync(() -> {
            try {
                String json = gson.toJson(new PasswordResetRequest(email));
                RequestBody body = RequestBody.create(json, JSON);
                
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/auth/password-reset")
                        .post(body)
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Password reset request failed: " + response.code());
                    }
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Password reset request failed", e);
            }
        });
    }
    
    /**
     * Reset password with token.
     * 
     * @param token The reset token
     * @param newPassword The new password
     * @return CompletableFuture that completes when password is reset
     */
    public CompletableFuture<Void> resetPassword(String token, String newPassword) {
        return CompletableFuture.runAsync(() -> {
            try {
                PasswordResetConfirmRequest requestBody = PasswordResetConfirmRequest.builder()
                        .token(token)
                        .newPassword(newPassword)
                        .build();
                
                String json = gson.toJson(requestBody);
                RequestBody body = RequestBody.create(json, JSON);
                
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/auth/password-reset/confirm")
                        .post(body)
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Password reset failed: " + response.code());
                    }
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Password reset request failed", e);
            }
        });
    }
    
    // Helper classes for requests
    private static class PasswordResetRequest {
        private final String email;
        
        public PasswordResetRequest(String email) {
            this.email = email;
        }
        
        public String getEmail() {
            return email;
        }
    }
    
    private static class PasswordResetConfirmRequest {
        private final String token;
        private final String newPassword;
        
        @lombok.Builder
        public PasswordResetConfirmRequest(String token, String newPassword) {
            this.token = token;
            this.newPassword = newPassword;
        }
        
        public String getToken() {
            return token;
        }
        
        public String getNewPassword() {
            return newPassword;
        }
    }
} 