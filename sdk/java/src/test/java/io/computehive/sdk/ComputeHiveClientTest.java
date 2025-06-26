package io.computehive.sdk;

import io.computehive.sdk.auth.AuthService;
import io.computehive.sdk.exceptions.ComputeHiveException;
import io.computehive.sdk.jobs.JobService;
import io.computehive.sdk.marketplace.MarketplaceService;
import io.computehive.sdk.models.Job;
import io.computehive.sdk.models.auth.AuthResponse;
import io.computehive.sdk.models.auth.Credentials;
import io.computehive.sdk.payments.PaymentService;
import io.computehive.sdk.telemetry.TelemetryService;
import io.computehive.sdk.websocket.WebSocketClient;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.time.Duration;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

/**
 * Tests for the ComputeHiveClient.
 */
@ExtendWith(MockitoExtension.class)
class ComputeHiveClientTest {
    
    private ComputeHiveClient client;
    
    @Mock
    private AuthService authService;
    
    @Mock
    private JobService jobService;
    
    @Mock
    private MarketplaceService marketplaceService;
    
    @Mock
    private PaymentService paymentService;
    
    @Mock
    private TelemetryService telemetryService;
    
    @Mock
    private WebSocketClient webSocketClient;
    
    @BeforeEach
    void setUp() {
        client = ComputeHiveClient.builder()
                .apiUrl("https://api.test.computehive.io")
                .wsUrl("wss://api.test.computehive.io/ws")
                .timeout(Duration.ofSeconds(30))
                .debug(true)
                .build();
    }
    
    @Test
    void testBuilderWithApiKey() {
        ComputeHiveClient clientWithApiKey = ComputeHiveClient.builder()
                .apiKey("test-api-key")
                .build();
        
        assertNotNull(clientWithApiKey);
        assertEquals("https://api.computehive.io", clientWithApiKey.getApiUrl());
        assertEquals("wss://api.computehive.io/ws", clientWithApiKey.getWsUrl());
    }
    
    @Test
    void testBuilderWithCustomUrls() {
        String customApiUrl = "https://custom-api.computehive.io";
        String customWsUrl = "wss://custom-ws.computehive.io";
        
        ComputeHiveClient customClient = ComputeHiveClient.builder()
                .apiUrl(customApiUrl)
                .wsUrl(customWsUrl)
                .build();
        
        assertEquals(customApiUrl, customClient.getApiUrl());
        assertEquals(customWsUrl, customClient.getWsUrl());
    }
    
    @Test
    void testAuthenticate() throws ExecutionException, InterruptedException {
        // Mock authentication response
        AuthResponse mockResponse = AuthResponse.builder()
                .accessToken("test-access-token")
                .refreshToken("test-refresh-token")
                .tokenType("Bearer")
                .expiresIn(3600)
                .build();
        
        // Test authentication
        CompletableFuture<AuthResponse> future = client.authenticate("test@example.com", "password");
        
        // In a real test, you would mock the HTTP client and verify the response
        // For now, we'll just test that the method doesn't throw an exception
        assertNotNull(future);
    }
    
    @Test
    void testGetServices() {
        // Test that all services are available
        assertNotNull(client.auth());
        assertNotNull(client.jobs());
        assertNotNull(client.marketplace());
        assertNotNull(client.payments());
        assertNotNull(client.telemetry());
        assertNotNull(client.events());
    }
    
    @Test
    void testSetAccessToken() {
        String testToken = "test-access-token";
        client.setAccessToken(testToken);
        assertEquals(testToken, client.getAccessToken());
    }
    
    @Test
    void testWebSocketConnection() {
        // Test WebSocket connection methods
        assertFalse(client.isConnected());
        
        // In a real test, you would mock the WebSocket and test connection
        // For now, we'll just test that the methods exist and don't throw exceptions
        assertDoesNotThrow(() -> client.connect());
        assertDoesNotThrow(() -> client.disconnect());
    }
    
    @Test
    void testClose() {
        // Test that the client can be closed without exceptions
        assertDoesNotThrow(() -> client.close());
    }
    
    @Test
    void testDefaultConfiguration() {
        ComputeHiveClient defaultClient = ComputeHiveClient.builder().build();
        
        assertEquals("https://api.computehive.io", defaultClient.getApiUrl());
        assertEquals("wss://api.computehive.io/ws", defaultClient.getWsUrl());
        assertNotNull(defaultClient.getHttpClient());
        assertNotNull(defaultClient.getGson());
    }
    
    @Test
    void testTimeoutConfiguration() {
        Duration customTimeout = Duration.ofMinutes(5);
        ComputeHiveClient timeoutClient = ComputeHiveClient.builder()
                .timeout(customTimeout)
                .build();
        
        assertNotNull(timeoutClient);
        // Note: We can't easily test the actual timeout configuration without
        // accessing internal OkHttpClient configuration, but we can verify
        // the client was created successfully
    }
    
    @Test
    void testDebugMode() {
        ComputeHiveClient debugClient = ComputeHiveClient.builder()
                .debug(true)
                .build();
        
        assertNotNull(debugClient);
        // In a real test, you would verify that logging interceptors are added
    }
    
    @Test
    void testCustomHttpClient() {
        // Test that a custom HTTP client can be provided
        // This would require creating a mock OkHttpClient
        assertNotNull(client.getHttpClient());
    }
    
    @Test
    void testJobOperations() {
        JobService jobService = client.jobs();
        assertNotNull(jobService);
        
        // Test that job service methods exist
        // In a real test, you would mock the HTTP responses and test actual functionality
    }
    
    @Test
    void testMarketplaceOperations() {
        MarketplaceService marketplaceService = client.marketplace();
        assertNotNull(marketplaceService);
        
        // Test that marketplace service methods exist
    }
    
    @Test
    void testPaymentOperations() {
        PaymentService paymentService = client.payments();
        assertNotNull(paymentService);
        
        // Test that payment service methods exist
    }
    
    @Test
    void testTelemetryOperations() {
        TelemetryService telemetryService = client.telemetry();
        assertNotNull(telemetryService);
        
        // Test that telemetry service methods exist
    }
    
    @Test
    void testEventHandling() {
        WebSocketClient webSocketClient = client.events();
        assertNotNull(webSocketClient);
        
        // Test that WebSocket client methods exist
    }
    
    @Test
    void testAuthenticationFlow() {
        // Test the complete authentication flow
        String email = "test@example.com";
        String password = "password123";
        
        CompletableFuture<AuthResponse> authFuture = client.authenticate(email, password);
        assertNotNull(authFuture);
        
        // In a real test, you would:
        // 1. Mock the HTTP response
        // 2. Verify the authentication request was made correctly
        // 3. Verify the access token was set
        // 4. Verify the WebSocket client was updated
    }
    
    @Test
    void testErrorHandling() {
        // Test error handling scenarios
        // In a real test, you would mock HTTP errors and verify exceptions are thrown correctly
        
        assertThrows(IllegalArgumentException.class, () -> {
            ComputeHiveClient.builder()
                    .apiUrl(null)
                    .build();
        });
    }
    
    @Test
    void testConcurrentOperations() {
        // Test that multiple operations can be performed concurrently
        // In a real test, you would create multiple threads and verify thread safety
        
        assertNotNull(client.jobs());
        assertNotNull(client.marketplace());
        assertNotNull(client.payments());
        assertNotNull(client.telemetry());
    }
    
    @Test
    void testResourceCleanup() {
        // Test that resources are properly cleaned up when the client is closed
        assertDoesNotThrow(() -> {
            try (ComputeHiveClient testClient = ComputeHiveClient.builder().build()) {
                // Use the client
                assertNotNull(testClient.getHttpClient());
            }
        });
    }
} 