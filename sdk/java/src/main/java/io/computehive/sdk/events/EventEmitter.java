package io.computehive.sdk.events;


import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * Event emitter for handling events in the SDK.
 */
public class EventEmitter {
    
    private final ConcurrentHashMap<String, List<EventListener>> listeners = new ConcurrentHashMap<>();
    
    /**
     * Add an event listener.
     * 
     * @param eventType The event type to listen for
     * @param listener The event listener
     */
    public void on(String eventType, EventListener listener) {
        if (eventType == null || listener == null) {
            throw new IllegalArgumentException("Event type and listener cannot be null");
        }
        
        listeners.computeIfAbsent(eventType, k -> new CopyOnWriteArrayList<>()).add(listener);
        log.debug("Added listener for event type: {}", eventType);
    }
    
    /**
     * Remove an event listener.
     * 
     * @param eventType The event type
     * @param listener The event listener to remove
     */
    public void off(String eventType, EventListener listener) {
        if (eventType == null || listener == null) {
            return;
        }
        
        List<EventListener> eventListeners = listeners.get(eventType);
        if (eventListeners != null) {
            eventListeners.remove(listener);
            log.debug("Removed listener for event type: {}", eventType);
        }
    }
    
    /**
     * Remove all listeners for a specific event type.
     * 
     * @param eventType The event type
     */
    public void removeAllListeners(String eventType) {
        if (eventType == null) {
            return;
        }
        
        listeners.remove(eventType);
        log.debug("Removed all listeners for event type: {}", eventType);
    }
    
    /**
     * Remove all listeners for all event types.
     */
    public void removeAllListeners() {
        listeners.clear();
        log.debug("Removed all listeners");
    }
    
    /**
     * Get the number of listeners for a specific event type.
     * 
     * @param eventType The event type
     * @return The number of listeners
     */
    public int listenerCount(String eventType) {
        if (eventType == null) {
            return 0;
        }
        
        List<EventListener> eventListeners = listeners.get(eventType);
        return eventListeners != null ? eventListeners.size() : 0;
    }
    
    /**
     * Get all event types that have listeners.
     * 
     * @return List of event types
     */
    public List<String> eventNames() {
        return new ArrayList<>(listeners.keySet());
    }
    
    /**
     * Emit an event to all registered listeners.
     * 
     * @param eventType The event type
     * @param data The event data
     */
    public void emit(String eventType, Object data) {
        if (eventType == null) {
            return;
        }
        
        List<EventListener> eventListeners = listeners.get(eventType);
        if (eventListeners != null && !eventListeners.isEmpty()) {
            log.debug("Emitting event: {} with {} listeners", eventType, eventListeners.size());
            
            for (EventListener listener : eventListeners) {
                try {
                    listener.onEvent(eventType, data);
                } catch (Exception e) {
                    log.error("Error in event listener for event type: {}", eventType, e);
                }
            }
        } else {
            log.debug("No listeners for event type: {}", eventType);
        }
    }
    
    /**
     * Emit an event to all registered listeners synchronously.
     * 
     * @param eventType The event type
     * @param data The event data
     */
    public void emitSync(String eventType, Object data) {
        emit(eventType, data);
    }
    
    /**
     * Add a one-time event listener that will be removed after the first emission.
     * 
     * @param eventType The event type to listen for
     * @param listener The event listener
     */
    public void once(String eventType, EventListener listener) {
        if (eventType == null || listener == null) {
            throw new IllegalArgumentException("Event type and listener cannot be null");
        }
        
        EventListener onceListener = new EventListener() {
            @Override
            public void onEvent(String type, Object data) {
                // Remove this listener after first call
                off(eventType, this);
                // Call the original listener
                listener.onEvent(type, data);
            }
        };
        
        on(eventType, onceListener);
    }
    
    /**
     * Add a listener that will be called before other listeners.
     * 
     * @param eventType The event type to listen for
     * @param listener The event listener
     */
    public void prependListener(String eventType, EventListener listener) {
        if (eventType == null || listener == null) {
            throw new IllegalArgumentException("Event type and listener cannot be null");
        }
        
        List<EventListener> eventListeners = listeners.computeIfAbsent(eventType, k -> new CopyOnWriteArrayList<>());
        eventListeners.add(0, listener);
        log.debug("Added prepend listener for event type: {}", eventType);
    }
    
    /**
     * Add a one-time listener that will be called before other listeners.
     * 
     * @param eventType The event type to listen for
     * @param listener The event listener
     */
    public void prependOnceListener(String eventType, EventListener listener) {
        if (eventType == null || listener == null) {
            throw new IllegalArgumentException("Event type and listener cannot be null");
        }
        
        EventListener onceListener = new EventListener() {
            @Override
            public void onEvent(String type, Object data) {
                // Remove this listener after first call
                off(eventType, this);
                // Call the original listener
                listener.onEvent(type, data);
            }
        };
        
        prependListener(eventType, onceListener);
    }
    
    /**
     * Get the maximum number of listeners that can be registered for an event type.
     * 
     * @return The maximum number of listeners (default: unlimited)
     */
    public int getMaxListeners() {
        return Integer.MAX_VALUE;
    }
    
    /**
     * Set the maximum number of listeners that can be registered for an event type.
     * 
     * @param maxListeners The maximum number of listeners
     */
    public void setMaxListeners(int maxListeners) {
        // This is a simplified implementation - in a real implementation,
        // you would want to enforce this limit when adding listeners
        log.warn("setMaxListeners is not implemented in this version");
    }
    
    /**
     * Get the raw listeners for a specific event type.
     * 
     * @param eventType The event type
     * @return List of event listeners
     */
    public List<EventListener> rawListeners(String eventType) {
        if (eventType == null) {
            return new ArrayList<>();
        }
        
        List<EventListener> eventListeners = listeners.get(eventType);
        return eventListeners != null ? new ArrayList<>(eventListeners) : new ArrayList<>();
    }
} 