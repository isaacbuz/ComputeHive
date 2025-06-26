package io.computehive.sdk;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import io.computehive.sdk.auth.AuthService;
import io.computehive.sdk.exceptions.ComputeHiveException;
import io.computehive.sdk.jobs.JobService;
import io.computehive.sdk.marketplace.MarketplaceService;
import io.computehive.sdk.models.auth.AuthResponse;
import io.computehive.sdk.models.auth.Credentials;
import io.computehive.sdk.payments.PaymentService;
import io.computehive.sdk.telemetry.TelemetryService;
import io.computehive.sdk.websocket.WebSocketClient;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.logging.HttpLoggingInterceptor;

import java.time.Duration;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;

/**
 * Main client for interacting with the ComputeHive API.
 * 
 * <pre>{@code
 * ComputeHiveClient client = ComputeHiveClient.builder()
 *     .apiKey("your-api-key")
 *     .build();
 * 
 * // Or with email/password authentication
 * ComputeHiveClient client = ComputeHiveClient.builder()
 *     .apiUrl("https://api.computehive.io")
 *     .build();
 * 
 * client.authenticate("user@example.com", "password").join();
 * }</pre>
 */
public class ComputeHiveClient implements AutoCloseable {
    
    private static final String DEFAULT_API_URL = "https://api.computehive.io";
    private static final String DEFAULT_WS_URL = "wss://api.computehive.io/ws";
    
    private final String apiUrl;
    private final String wsUrl;
    private final OkHttpClient httpClient;
    private final Gson gson;
    private final WebSocketClient webSocketClient;
    
    // Services
    private final AuthService authService;
    private final JobService jobService;
    private final MarketplaceService marketplaceService;
    private final PaymentService paymentService;
    private final TelemetryService telemetryService;
    
    private String accessToken;
    
    private ComputeHiveClient(
            String apiUrl,
            String wsUrl,
            String apiKey,
            Duration timeout,
            boolean debug,
            OkHttpClient customHttpClient) {
        
        this.apiUrl = apiUrl != null ? apiUrl : DEFAULT_API_URL;
        this.wsUrl = wsUrl != null ? wsUrl : DEFAULT_WS_URL;
        
        // Configure Gson
        this.gson = new GsonBuilder()
                .setDateFormat("yyyy-MM-dd'T'HH:mm:ss.SSS'Z'")
                .create();
        
        // Configure HTTP client
        if (customHttpClient != null) {
            this.httpClient = customHttpClient;
        } else {
            OkHttpClient.Builder builder = new OkHttpClient.Builder()
                    .connectTimeout(timeout != null ? timeout : Duration.ofSeconds(30))
                    .readTimeout(timeout != null ? timeout : Duration.ofSeconds(30))
                    .writeTimeout(timeout != null ? timeout : Duration.ofSeconds(30));
            
            // Add auth interceptor
            builder.addInterceptor(chain -> {
                Request original = chain.request();
                Request.Builder requestBuilder = original.newBuilder();
                
                if (accessToken != null) {
                    requestBuilder.header("Authorization", "Bearer " + accessToken);
                } else if (apiKey != null) {
                    requestBuilder.header("X-API-Key", apiKey);
                }
                
                return chain.proceed(requestBuilder.build());
            });
            
            // Add logging interceptor if debug is enabled
            if (debug) {
                HttpLoggingInterceptor loggingInterceptor = new HttpLoggingInterceptor(
                    message -> log.debug("[HTTP] {}", message)
                );
                loggingInterceptor.setLevel(HttpLoggingInterceptor.Level.BODY);
                builder.addInterceptor(loggingInterceptor);
            }
            
            this.httpClient = builder.build();
        }
        
        // Initialize WebSocket client
        this.webSocketClient = new WebSocketClient(this.wsUrl, this.accessToken);
        
        // Initialize services
        this.authService = new AuthService(this);
        this.jobService = new JobService(this);
        this.marketplaceService = new MarketplaceService(this);
        this.paymentService = new PaymentService(this);
        this.telemetryService = new TelemetryService(this);
        
        // Set initial access token if API key provided
        if (apiKey != null) {
            this.accessToken = apiKey;
        }
    }
    
    /**
     * Authenticate with email and password.
     * 
     * @param email User email
     * @param password User password
     * @return CompletableFuture containing the authentication response
     */
    public CompletableFuture<AuthResponse> authenticate(String email, String password) {
        Credentials credentials = Credentials.builder()
                .email(email)
                .password(password)
                .build();
        
        return authService.authenticate(credentials)
                .thenApply(response -> {
                    this.accessToken = response.getAccessToken();
                    this.webSocketClient.updateAccessToken(this.accessToken);
                    return response;
                });
    }
    
    /**
     * Connect to WebSocket for real-time updates.
     */
    public void connect() {
        webSocketClient.connect();
    }
    
    /**
     * Disconnect from WebSocket.
     */
    public void disconnect() {
        webSocketClient.disconnect();
    }
    
    /**
     * Check if WebSocket is connected.
     * 
     * @return true if connected, false otherwise
     */
    public boolean isConnected() {
        return webSocketClient.isConnected();
    }
    
    /**
     * Get the current access token.
     * 
     * @return The access token or null if not authenticated
     */
    public String getAccessToken() {
        return accessToken;
    }
    
    /**
     * Set a new access token.
     * 
     * @param accessToken The new access token
     */
    public void setAccessToken(String accessToken) {
        this.accessToken = accessToken;
        this.webSocketClient.updateAccessToken(accessToken);
    }
    
    /**
     * Get the job service for job-related operations.
     * 
     * @return The job service
     */
    public JobService jobs() {
        return jobService;
    }
    
    /**
     * Get the marketplace service for marketplace operations.
     * 
     * @return The marketplace service
     */
    public MarketplaceService marketplace() {
        return marketplaceService;
    }
    
    /**
     * Get the payment service for payment operations.
     * 
     * @return The payment service
     */
    public PaymentService payments() {
        return paymentService;
    }
    
    /**
     * Get the telemetry service for metrics and monitoring.
     * 
     * @return The telemetry service
     */
    public TelemetryService telemetry() {
        return telemetryService;
    }
    
    /**
     * Get the authentication service.
     * 
     * @return The auth service
     */
    public AuthService auth() {
        return authService;
    }
    
    /**
     * Get the WebSocket event emitter for subscribing to real-time events.
     * 
     * @return The WebSocket client
     */
    public WebSocketClient events() {
        return webSocketClient;
    }
    
    @Override
    public void close() {
        try {
            disconnect();
            httpClient.dispatcher().executorService().shutdown();
            httpClient.connectionPool().evictAll();
            if (httpClient.cache() != null) {
                httpClient.cache().close();
            }
        } catch (Exception e) {
            log.error("Error closing ComputeHive client", e);
        }
    }
    
    /**
     * Create a new ComputeHive client builder.
     * 
     * @return A new builder instance
     */
    public static ComputeHiveClientBuilder builder() {
        return new ComputeHiveClientBuilder();
    }
    
    /**
     * Custom builder class for better API.
     */
    public static class ComputeHiveClientBuilder {
    }
} 