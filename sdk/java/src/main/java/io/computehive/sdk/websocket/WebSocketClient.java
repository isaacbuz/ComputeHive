package io.computehive.sdk.websocket;

import com.google.gson.Gson;
import io.computehive.sdk.events.EventEmitter;
import io.computehive.sdk.events.EventListener;
import okhttp3.*;
import okio.ByteString;

import java.util.concurrent.TimeUnit;

/**
 * WebSocket client for real-time event handling.
 */
public class WebSocketClient {
    
    private final String wsUrl;
    private final OkHttpClient httpClient;
    private final Gson gson;
    private final EventEmitter eventEmitter;
    
    private WebSocket webSocket;
    private String accessToken;
    private boolean isConnected = false;
    private int reconnectAttempts = 0;
    private static final int MAX_RECONNECT_ATTEMPTS = 5;
    private static final long RECONNECT_DELAY_MS = 1000;
    
    public WebSocketClient(String wsUrl, String accessToken) {
        this.wsUrl = wsUrl;
        this.accessToken = accessToken;
        this.gson = new Gson();
        this.eventEmitter = new EventEmitter();
        
        this.httpClient = new OkHttpClient.Builder()
                .connectTimeout(30, TimeUnit.SECONDS)
                .readTimeout(30, TimeUnit.SECONDS)
                .writeTimeout(30, TimeUnit.SECONDS)
                .build();
    }
    
    /**
     * Connect to the WebSocket server.
     */
    public void connect() {
        if (isConnected) {
            log.warn("WebSocket is already connected");
            return;
        }
        
        try {
            Request.Builder requestBuilder = new Request.Builder()
                    .url(wsUrl);
            
            if (accessToken != null) {
                requestBuilder.addHeader("Authorization", "Bearer " + accessToken);
            }
            
            Request request = requestBuilder.build();
            
            webSocket = httpClient.newWebSocket(request, new WebSocketListener() {
                @Override
                public void onOpen(WebSocket webSocket, Response response) {
                    log.info("WebSocket connected successfully");
                    isConnected = true;
                    reconnectAttempts = 0;
                    eventEmitter.emit("connected", null);
                }
                
                @Override
                public void onMessage(WebSocket webSocket, String text) {
                    handleMessage(text);
                }
                
                @Override
                public void onMessage(WebSocket webSocket, ByteString bytes) {
                    handleMessage(bytes.utf8());
                }
                
                @Override
                public void onClosing(WebSocket webSocket, int code, String reason) {
                    log.info("WebSocket closing: {} - {}", code, reason);
                    isConnected = false;
                    eventEmitter.emit("closing", new WebSocketCloseEvent(code, reason));
                }
                
                @Override
                public void onClosed(WebSocket webSocket, int code, String reason) {
                    log.info("WebSocket closed: {} - {}", code, reason);
                    isConnected = false;
                    eventEmitter.emit("closed", new WebSocketCloseEvent(code, reason));
                }
                
                @Override
                public void onFailure(WebSocket webSocket, Throwable t, Response response) {
                    log.error("WebSocket connection failed", t);
                    isConnected = false;
                    eventEmitter.emit("error", t);
                    
                    // Attempt to reconnect
                    if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
                        reconnectAttempts++;
                        log.info("Attempting to reconnect (attempt {}/{})", reconnectAttempts, MAX_RECONNECT_ATTEMPTS);
                        
                        httpClient.dispatcher().executorService().schedule(() -> {
                            connect();
                        }, RECONNECT_DELAY_MS * reconnectAttempts, TimeUnit.MILLISECONDS);
                    } else {
                        log.error("Max reconnection attempts reached");
                        eventEmitter.emit("maxReconnectAttemptsReached", null);
                    }
                }
            });
            
        } catch (Exception e) {
            log.error("Failed to create WebSocket connection", e);
            eventEmitter.emit("error", e);
        }
    }
    
    /**
     * Disconnect from the WebSocket server.
     */
    public void disconnect() {
        if (webSocket != null) {
            webSocket.close(1000, "Client disconnecting");
            webSocket = null;
        }
        isConnected = false;
    }
    
    /**
     * Send a message to the WebSocket server.
     * 
     * @param message The message to send
     * @return true if the message was sent successfully
     */
    public boolean sendMessage(String message) {
        if (!isConnected || webSocket == null) {
            log.warn("WebSocket is not connected");
            return false;
        }
        
        return webSocket.send(message);
    }
    
    /**
     * Send a JSON message to the WebSocket server.
     * 
     * @param data The data to send as JSON
     * @return true if the message was sent successfully
     */
    public boolean sendJson(Object data) {
        try {
            String json = gson.toJson(data);
            return sendMessage(json);
        } catch (Exception e) {
            log.error("Failed to serialize message to JSON", e);
            return false;
        }
    }
    
    /**
     * Subscribe to a specific event type.
     * 
     * @param eventType The event type to subscribe to
     * @return true if the subscription was sent successfully
     */
    public boolean subscribe(String eventType) {
        SubscribeMessage subscribeMessage = SubscribeMessage.builder()
                .type("subscribe")
                .eventType(eventType)
                .build();
        
        return sendJson(subscribeMessage);
    }
    
    /**
     * Unsubscribe from a specific event type.
     * 
     * @param eventType The event type to unsubscribe from
     * @return true if the unsubscription was sent successfully
     */
    public boolean unsubscribe(String eventType) {
        UnsubscribeMessage unsubscribeMessage = UnsubscribeMessage.builder()
                .type("unsubscribe")
                .eventType(eventType)
                .build();
        
        return sendJson(unsubscribeMessage);
    }
    
    /**
     * Add an event listener.
     * 
     * @param eventType The event type to listen for
     * @param listener The event listener
     */
    public void on(String eventType, EventListener listener) {
        eventEmitter.on(eventType, listener);
    }
    
    /**
     * Remove an event listener.
     * 
     * @param eventType The event type
     * @param listener The event listener to remove
     */
    public void off(String eventType, EventListener listener) {
        eventEmitter.off(eventType, listener);
    }
    
    /**
     * Check if the WebSocket is connected.
     * 
     * @return true if connected, false otherwise
     */
    public boolean isConnected() {
        return isConnected;
    }
    
    /**
     * Update the access token for authentication.
     * 
     * @param accessToken The new access token
     */
    public void updateAccessToken(String accessToken) {
        this.accessToken = accessToken;
        
        // Reconnect if currently connected to update authentication
        if (isConnected) {
            log.info("Reconnecting with new access token");
            disconnect();
            connect();
        }
    }
    
    /**
     * Handle incoming WebSocket messages.
     * 
     * @param message The received message
     */
    private void handleMessage(String message) {
        try {
            log.debug("Received WebSocket message: {}", message);
            
            // Try to parse as JSON first
            try {
                WebSocketMessage wsMessage = gson.fromJson(message, WebSocketMessage.class);
                
                if (wsMessage != null && wsMessage.getType() != null) {
                    switch (wsMessage.getType()) {
                        case "event":
                            handleEventMessage(wsMessage);
                            break;
                        case "ping":
                            handlePingMessage();
                            break;
                        case "pong":
                            handlePongMessage();
                            break;
                        case "error":
                            handleErrorMessage(wsMessage);
                            break;
                        default:
                            log.warn("Unknown message type: {}", wsMessage.getType());
                    }
                }
            } catch (Exception e) {
                log.warn("Failed to parse message as JSON, treating as raw message");
                eventEmitter.emit("message", message);
            }
            
        } catch (Exception e) {
            log.error("Error handling WebSocket message", e);
        }
    }
    
    /**
     * Handle event messages.
     * 
     * @param message The event message
     */
    private void handleEventMessage(WebSocketMessage message) {
        String eventType = message.getEventType();
        Object data = message.getData();
        
        if (eventType != null) {
            eventEmitter.emit(eventType, data);
        }
    }
    
    /**
     * Handle ping messages.
     */
    private void handlePingMessage() {
        // Respond with pong
        PongMessage pongMessage = PongMessage.builder()
                .type("pong")
                .timestamp(System.currentTimeMillis())
                .build();
        
        sendJson(pongMessage);
    }
    
    /**
     * Handle pong messages.
     */
    private void handlePongMessage() {
        eventEmitter.emit("pong", null);
    }
    
    /**
     * Handle error messages.
     * 
     * @param message The error message
     */
    private void handleErrorMessage(WebSocketMessage message) {
        log.error("WebSocket error: {}", message.getData());
        eventEmitter.emit("error", message.getData());
    }
    
    // Helper classes
    @lombok.Data
    @lombok.Builder
    @lombok.NoArgsConstructor
    @lombok.AllArgsConstructor
    private static class WebSocketMessage {
        private String type;
        private String eventType;
        private Object data;
        private Long timestamp;
    }
    
    @lombok.Data
    @lombok.Builder
    @lombok.NoArgsConstructor
    @lombok.AllArgsConstructor
    private static class SubscribeMessage {
        private String type;
        private String eventType;
    }
    
    @lombok.Data
    @lombok.Builder
    @lombok.NoArgsConstructor
    @lombok.AllArgsConstructor
    private static class UnsubscribeMessage {
        private String type;
        private String eventType;
    }
    
    @lombok.Data
    @lombok.Builder
    @lombok.NoArgsConstructor
    @lombok.AllArgsConstructor
    private static class PongMessage {
        private String type;
        private Long timestamp;
    }
    
    /**
     * WebSocket close event information.
     */
    @lombok.Data
    @lombok.AllArgsConstructor
    public static class WebSocketCloseEvent {
        private int code;
        private String reason;
    }
} 