import * as LocalAuthentication from 'expo-local-authentication';
import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';

export interface BiometricType {
  type: 'fingerprint' | 'facial' | 'iris' | 'none';
  name: string;
}

export interface BiometricResult {
  success: boolean;
  error?: string;
  biometricType?: BiometricType;
}

class BiometricService {
  private static instance: BiometricService;
  private isAvailable: boolean = false;
  private biometricType: BiometricType | null = null;

  private constructor() {}

  static getInstance(): BiometricService {
    if (!BiometricService.instance) {
      BiometricService.instance = new BiometricService();
    }
    return BiometricService.instance;
  }

  /**
   * Check if biometric authentication is available on the device
   */
  async checkAvailability(): Promise<boolean> {
    try {
      const hasHardware = await LocalAuthentication.hasHardwareAsync();
      const isEnrolled = await LocalAuthentication.isEnrolledAsync();
      
      this.isAvailable = hasHardware && isEnrolled;
      
      if (this.isAvailable) {
        const supportedTypes = await LocalAuthentication.supportedAuthenticationTypesAsync();
        this.biometricType = this.mapBiometricType(supportedTypes);
      }
      
      return this.isAvailable;
    } catch (error) {
      console.error('Error checking biometric availability:', error);
      this.isAvailable = false;
      return false;
    }
  }

  /**
   * Get the type of biometric authentication available
   */
  getBiometricType(): BiometricType | null {
    return this.biometricType;
  }

  /**
   * Check if biometric authentication is available
   */
  isBiometricAvailable(): boolean {
    return this.isAvailable;
  }

  /**
   * Authenticate using biometrics
   */
  async authenticate(promptMessage?: string): Promise<BiometricResult> {
    try {
      if (!this.isAvailable) {
        return {
          success: false,
          error: 'Biometric authentication is not available',
        };
      }

      const result = await LocalAuthentication.authenticateAsync({
        promptMessage: promptMessage || 'Authenticate to continue',
        fallbackLabel: 'Use passcode',
        cancelLabel: 'Cancel',
        disableDeviceFallback: false,
      });

      if (result.success) {
        return {
          success: true,
          biometricType: this.biometricType || undefined,
        };
      } else {
        return {
          success: false,
          error: this.getErrorMessage(result.error),
        };
      }
    } catch (error) {
      console.error('Biometric authentication error:', error);
      return {
        success: false,
        error: 'Authentication failed',
      };
    }
  }

  /**
   * Store credentials securely for biometric authentication
   */
  async storeCredentials(email: string, password: string): Promise<boolean> {
    try {
      const credentials = JSON.stringify({ email, password });
      await SecureStore.setItemAsync('biometric_credentials', credentials);
      return true;
    } catch (error) {
      console.error('Error storing credentials:', error);
      return false;
    }
  }

  /**
   * Retrieve stored credentials
   */
  async getStoredCredentials(): Promise<{ email: string; password: string } | null> {
    try {
      const credentials = await SecureStore.getItemAsync('biometric_credentials');
      if (credentials) {
        return JSON.parse(credentials);
      }
      return null;
    } catch (error) {
      console.error('Error retrieving credentials:', error);
      return null;
    }
  }

  /**
   * Remove stored credentials
   */
  async removeStoredCredentials(): Promise<boolean> {
    try {
      await SecureStore.deleteItemAsync('biometric_credentials');
      return true;
    } catch (error) {
      console.error('Error removing credentials:', error);
      return false;
    }
  }

  /**
   * Check if credentials are stored
   */
  async hasStoredCredentials(): Promise<boolean> {
    try {
      const credentials = await SecureStore.getItemAsync('biometric_credentials');
      return !!credentials;
    } catch (error) {
      console.error('Error checking stored credentials:', error);
      return false;
    }
  }

  /**
   * Get biometric authentication status
   */
  async getAuthenticationStatus(): Promise<{
    isAvailable: boolean;
    biometricType: BiometricType | null;
    hasStoredCredentials: boolean;
  }> {
    const isAvailable = await this.checkAvailability();
    const hasStoredCredentials = await this.hasStoredCredentials();

    return {
      isAvailable,
      biometricType: this.biometricType,
      hasStoredCredentials,
    };
  }

  /**
   * Map authentication types to readable names
   */
  private mapBiometricType(types: number[]): BiometricType {
    if (types.includes(LocalAuthentication.AuthenticationType.FINGERPRINT)) {
      return {
        type: 'fingerprint',
        name: Platform.OS === 'ios' ? 'Touch ID' : 'Fingerprint',
      };
    }
    
    if (types.includes(LocalAuthentication.AuthenticationType.FACIAL_RECOGNITION)) {
      return {
        type: 'facial',
        name: Platform.OS === 'ios' ? 'Face ID' : 'Face Recognition',
      };
    }
    
    if (types.includes(LocalAuthentication.AuthenticationType.IRIS)) {
      return {
        type: 'iris',
        name: 'Iris Recognition',
      };
    }
    
    return {
      type: 'none',
      name: 'None',
    };
  }

  /**
   * Get user-friendly error messages
   */
  private getErrorMessage(error: string): string {
    switch (error) {
      case 'UserCancel':
        return 'Authentication was cancelled';
      case 'UserFallback':
        return 'User chose to use fallback authentication';
      case 'SystemCancel':
        return 'Authentication was cancelled by the system';
      case 'AuthenticationFailed':
        return 'Authentication failed';
      case 'PasscodeNotSet':
        return 'No passcode is set on the device';
      case 'NotAvailable':
        return 'Biometric authentication is not available';
      case 'NotEnrolled':
        return 'No biometric authentication is enrolled';
      case 'Lockout':
        return 'Too many failed attempts. Try again later';
      case 'AppCancel':
        return 'Authentication was cancelled by the app';
      case 'InvalidContext':
        return 'Invalid authentication context';
      case 'NotInteractive':
        return 'Authentication requires user interaction';
      default:
        return 'Authentication failed';
    }
  }

  /**
   * Get platform-specific biometric name
   */
  getPlatformBiometricName(): string {
    if (Platform.OS === 'ios') {
      if (this.biometricType?.type === 'facial') {
        return 'Face ID';
      } else if (this.biometricType?.type === 'fingerprint') {
        return 'Touch ID';
      }
    } else if (Platform.OS === 'android') {
      if (this.biometricType?.type === 'facial') {
        return 'Face Recognition';
      } else if (this.biometricType?.type === 'fingerprint') {
        return 'Fingerprint';
      }
    }
    
    return 'Biometric Authentication';
  }

  /**
   * Check if the device supports biometric authentication
   */
  async isDeviceSupported(): Promise<boolean> {
    try {
      const hasHardware = await LocalAuthentication.hasHardwareAsync();
      return hasHardware;
    } catch (error) {
      console.error('Error checking device support:', error);
      return false;
    }
  }

  /**
   * Get supported authentication types
   */
  async getSupportedTypes(): Promise<number[]> {
    try {
      return await LocalAuthentication.supportedAuthenticationTypesAsync();
    } catch (error) {
      console.error('Error getting supported types:', error);
      return [];
    }
  }
}

export default BiometricService.getInstance();
