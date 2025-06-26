package io.computehive.sdk.exceptions;

/**
 * Exception thrown by the ComputeHive SDK.
 */
public class ComputeHiveException extends RuntimeException {
    
    private final int statusCode;
    private final String errorCode;
    
    /**
     * Constructs a new ComputeHiveException with the specified message.
     * 
     * @param message The detail message
     */
    public ComputeHiveException(String message) {
        super(message);
        this.statusCode = -1;
        this.errorCode = null;
    }
    
    /**
     * Constructs a new ComputeHiveException with the specified message and cause.
     * 
     * @param message The detail message
     * @param cause The cause
     */
    public ComputeHiveException(String message, Throwable cause) {
        super(message, cause);
        this.statusCode = -1;
        this.errorCode = null;
    }
    
    /**
     * Constructs a new ComputeHiveException with the specified message, status code, and error code.
     * 
     * @param message The detail message
     * @param statusCode The HTTP status code
     * @param errorCode The error code
     */
    public ComputeHiveException(String message, int statusCode, String errorCode) {
        super(message);
        this.statusCode = statusCode;
        this.errorCode = errorCode;
    }
    
    /**
     * Constructs a new ComputeHiveException with the specified message, cause, status code, and error code.
     * 
     * @param message The detail message
     * @param cause The cause
     * @param statusCode The HTTP status code
     * @param errorCode The error code
     */
    public ComputeHiveException(String message, Throwable cause, int statusCode, String errorCode) {
        super(message, cause);
        this.statusCode = statusCode;
        this.errorCode = errorCode;
    }
    
    /**
     * Get the HTTP status code associated with this exception.
     * 
     * @return The status code, or -1 if not available
     */
    public int getStatusCode() {
        return statusCode;
    }
    
    /**
     * Get the error code associated with this exception.
     * 
     * @return The error code, or null if not available
     */
    public String getErrorCode() {
        return errorCode;
    }
    
    /**
     * Check if this exception has a status code.
     * 
     * @return true if a status code is available, false otherwise
     */
    public boolean hasStatusCode() {
        return statusCode != -1;
    }
    
    /**
     * Check if this exception has an error code.
     * 
     * @return true if an error code is available, false otherwise
     */
    public boolean hasErrorCode() {
        return errorCode != null;
    }
    
    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder();
        sb.append("ComputeHiveException");
        
        if (hasStatusCode() || hasErrorCode()) {
            sb.append(" [");
            if (hasStatusCode()) {
                sb.append("statusCode=").append(statusCode);
            }
            if (hasStatusCode() && hasErrorCode()) {
                sb.append(", ");
            }
            if (hasErrorCode()) {
                sb.append("errorCode=").append(errorCode);
            }
            sb.append("]");
        }
        
        sb.append(": ").append(getMessage());
        
        if (getCause() != null) {
            sb.append("; caused by: ").append(getCause().getMessage());
        }
        
        return sb.toString();
    }
} 