package io.computehive.sdk.events;

/**
 * Interface for event listeners.
 */
@FunctionalInterface
public interface EventListener {
    
    /**
     * Called when an event is emitted.
     * 
     * @param eventType The type of event that was emitted
     * @param data The data associated with the event (can be null)
     */
    void onEvent(String eventType, Object data);
} 