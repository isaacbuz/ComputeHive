import ReactNativeBiometrics, { BiometryTypes } from 'react-native-biometrics';
import * as Keychain from 'react-native-keychain';

interface BiometricResult {
  success: boolean;
  error?: string;
}

class BiometricServiceClass {
  private rnBiometrics: ReactNativeBiometrics;

  constructor() {
    this.rnBiometrics = new ReactNativeBiometrics({ allowDeviceCredentials: true });
  }

  async isAvailable(): Promise<boolean> {
    try {
      const { available, biometryType } = await this.rnBiometrics.isSensorAvailable();
      return available;
    } catch (error) {
      console.error('Biometric availability check failed:', error);
      return false;
    }
  }

  async getBiometryType(): Promise<string | null> {
    try {
      const { available, biometryType } = await this.rnBiometrics.isSensorAvailable();
      if (available && biometryType) {
        switch (biometryType) {
          case BiometryTypes.TouchID:
            return 'Touch ID';
          case BiometryTypes.FaceID:
            return 'Face ID';
          case BiometryTypes.Biometrics:
            return 'Biometrics';
          default:
            return biometryType;
        }
      }
      return null;
    } catch (error) {
      console.error('Get biometry type failed:', error);
      return null;
    }
  }

  async authenticate(promptMessage: string = 'Authenticate'): Promise<BiometricResult> {
    try {
      const { success } = await this.rnBiometrics.simplePrompt({
        promptMessage,
        cancelButtonText: 'Cancel',
        fallbackPromptMessage: 'Use passcode',
      });

      return { success };
    } catch (error: any) {
      return {
        success: false,
        error: error.message || 'Authentication failed',
      };
    }
  }

  async createKeys(): Promise<boolean> {
    try {
      const { publicKey } = await this.rnBiometrics.createKeys();
      console.log('Biometric keys created, public key:', publicKey);
      return true;
    } catch (error) {
      console.error('Create keys failed:', error);
      return false;
    }
  }

  async deleteKeys(): Promise<boolean> {
    try {
      const { keysDeleted } = await this.rnBiometrics.deleteKeys();
      return keysDeleted;
    } catch (error) {
      console.error('Delete keys failed:', error);
      return false;
    }
  }

  async createSignature(
    payload: string,
    promptMessage: string = 'Sign transaction'
  ): Promise<{ success: boolean; signature?: string; error?: string }> {
    try {
      const { success, signature } = await this.rnBiometrics.createSignature({
        promptMessage,
        payload,
        cancelButtonText: 'Cancel',
      });

      if (success && signature) {
        return { success: true, signature };
      }
      
      return { success: false, error: 'Signature creation failed' };
    } catch (error: any) {
      return {
        success: false,
        error: error.message || 'Signature creation failed',
      };
    }
  }

  async checkAvailability(): Promise<{
    available: boolean;
    biometryType?: string;
    error?: string;
  }> {
    try {
      const { available, biometryType, error } = await this.rnBiometrics.isSensorAvailable();
      
      if (available && biometryType) {
        let type = biometryType;
        switch (biometryType) {
          case BiometryTypes.TouchID:
            type = 'Touch ID';
            break;
          case BiometryTypes.FaceID:
            type = 'Face ID';
            break;
          case BiometryTypes.Biometrics:
            type = 'Biometrics';
            break;
        }
        
        return { available: true, biometryType: type };
      }
      
      return { available: false, error: error || 'Biometrics not available' };
    } catch (error: any) {
      return {
        available: false,
        error: error.message || 'Failed to check biometric availability',
      };
    }
  }

  // Secure credential storage using Keychain
  async saveCredentials(username: string, password: string): Promise<boolean> {
    try {
      await Keychain.setInternetCredentials(
        'com.computehive.app',
        username,
        password
      );
      return true;
    } catch (error) {
      console.error('Save credentials failed:', error);
      return false;
    }
  }

  async getCredentials(): Promise<{ username: string; password: string } | null> {
    try {
      const credentials = await Keychain.getInternetCredentials('com.computehive.app');
      if (credentials) {
        return {
          username: credentials.username,
          password: credentials.password,
        };
      }
      return null;
    } catch (error) {
      console.error('Get credentials failed:', error);
      return null;
    }
  }

  async deleteCredentials(): Promise<boolean> {
    try {
      await Keychain.resetInternetCredentials('com.computehive.app');
      return true;
    } catch (error) {
      console.error('Delete credentials failed:', error);
      return false;
    }
  }

  // Check if biometric authentication is enrolled
  async isEnrolled(): Promise<boolean> {
    try {
      const { available, biometryType } = await this.rnBiometrics.isSensorAvailable();
      return available && biometryType !== null && biometryType !== undefined;
    } catch (error) {
      console.error('Enrollment check failed:', error);
      return false;
    }
  }
}

export const BiometricService = new BiometricServiceClass(); 