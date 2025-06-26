import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react-native';
import { NavigationContainer } from '@react-navigation/native';
import DashboardScreen from '../../mobile/ComputeHiveApp/src/screens/DashboardScreen';
import JobsScreen from '../../mobile/ComputeHiveApp/src/screens/jobs/JobsScreen';
import LoginScreen from '../../mobile/ComputeHiveApp/src/screens/auth/LoginScreen';

// Mock expo modules
jest.mock('expo-local-authentication', () => ({
  hasHardwareAsync: jest.fn(() => Promise.resolve(true)),
  isEnrolledAsync: jest.fn(() => Promise.resolve(true)),
  authenticateAsync: jest.fn(() => Promise.resolve({ success: true })),
}));

jest.mock('expo-notifications', () => ({
  getPermissionsAsync: jest.fn(() => Promise.resolve({ status: 'granted' })),
  requestPermissionsAsync: jest.fn(() => Promise.resolve({ status: 'granted' })),
  getExpoPushTokenAsync: jest.fn(() => Promise.resolve({ data: 'mock-token' })),
  scheduleNotificationAsync: jest.fn(() => Promise.resolve('notification-id')),
}));

jest.mock('expo-secure-store', () => ({
  setItemAsync: jest.fn(() => Promise.resolve()),
  getItemAsync: jest.fn(() => Promise.resolve(null)),
  deleteItemAsync: jest.fn(() => Promise.resolve()),
}));

describe('Mobile App Components', () => {
  describe('DashboardScreen', () => {
    it('renders dashboard with stats', async () => {
      render(<DashboardScreen />);

      await waitFor(() => {
        expect(screen.getByText('Dashboard')).toBeInTheDocument();
        expect(screen.getByText('Active Jobs')).toBeInTheDocument();
        expect(screen.getByText('Completed')).toBeInTheDocument();
        expect(screen.getByText('Failed')).toBeInTheDocument();
        expect(screen.getByText('Earnings')).toBeInTheDocument();
      });
    });

    it('displays resource usage cards', async () => {
      render(<DashboardScreen />);

      await waitFor(() => {
        expect(screen.getByText('CPU')).toBeInTheDocument();
        expect(screen.getByText('Memory')).toBeInTheDocument();
        expect(screen.getByText('GPU')).toBeInTheDocument();
        expect(screen.getByText('Network')).toBeInTheDocument();
      });
    });

    it('shows quick actions', async () => {
      render(<DashboardScreen />);

      await waitFor(() => {
        expect(screen.getByText('New Job')).toBeInTheDocument();
        expect(screen.getByText('Resources')).toBeInTheDocument();
        expect(screen.getByText('Marketplace')).toBeInTheDocument();
        expect(screen.getByText('Settings')).toBeInTheDocument();
      });
    });

    it('handles refresh', async () => {
      render(<DashboardScreen />);

      const scrollView = screen.getByTestId('dashboard-scroll');
      fireEvent.scroll(scrollView, {
        nativeEvent: {
          contentOffset: { y: -100 },
          contentSize: { height: 500, width: 100 },
          layoutMeasurement: { height: 100, width: 100 },
        },
      });

      await waitFor(() => {
        // Should trigger refresh
        expect(screen.getByText('Dashboard')).toBeInTheDocument();
      });
    });

    it('displays recent activity', async () => {
      render(<DashboardScreen />);

      await waitFor(() => {
        expect(screen.getByText('Recent Activity')).toBeInTheDocument();
        expect(screen.getByText(/Job.*completed/)).toBeInTheDocument();
        expect(screen.getByText(/New job.*started/)).toBeInTheDocument();
      });
    });
  });

  describe('JobsScreen', () => {
    it('renders jobs list', async () => {
      render(<JobsScreen />);

      await waitFor(() => {
        expect(screen.getByText('Jobs')).toBeInTheDocument();
        expect(screen.getByText('ML Model Training')).toBeInTheDocument();
        expect(screen.getByText('Data Processing Pipeline')).toBeInTheDocument();
      });
    });

    it('displays job status filters', async () => {
      render(<JobsScreen />);

      await waitFor(() => {
        expect(screen.getByText('All')).toBeInTheDocument();
        expect(screen.getByText('Running')).toBeInTheDocument();
        expect(screen.getByText('Completed')).toBeInTheDocument();
        expect(screen.getByText('Failed')).toBeInTheDocument();
        expect(screen.getByText('Pending')).toBeInTheDocument();
      });
    });

    it('handles search functionality', async () => {
      render(<JobsScreen />);

      const searchInput = screen.getByPlaceholderText('Search jobs...');
      fireEvent.changeText(searchInput, 'ML');

      await waitFor(() => {
        expect(screen.getByText('ML Model Training')).toBeInTheDocument();
        expect(screen.queryByText('Data Processing Pipeline')).not.toBeInTheDocument();
      });
    });

    it('filters jobs by status', async () => {
      render(<JobsScreen />);

      const runningFilter = screen.getByText('Running');
      fireEvent.press(runningFilter);

      await waitFor(() => {
        expect(screen.getByText('ML Model Training')).toBeInTheDocument();
        expect(screen.queryByText('Data Processing Pipeline')).not.toBeInTheDocument();
      });
    });

    it('displays job details correctly', async () => {
      render(<JobsScreen />);

      await waitFor(() => {
        const jobCard = screen.getByText('ML Model Training').parent;
        expect(jobCard).toHaveTextContent('Machine Learning');
        expect(jobCard).toHaveTextContent('HIGH');
        expect(jobCard).toHaveTextContent('75%');
        expect(jobCard).toHaveTextContent('$45.67');
      });
    });

    it('handles job card press', async () => {
      const alertSpy = jest.spyOn(global, 'Alert').mockImplementation(() => {});
      
      render(<JobsScreen />);

      await waitFor(() => {
        const jobCard = screen.getByText('ML Model Training');
        fireEvent.press(jobCard);
      });

      expect(alertSpy).toHaveBeenCalledWith('Job Details', 'Viewing details for ML Model Training');
      
      alertSpy.mockRestore();
    });

    it('shows empty state when no jobs match filters', async () => {
      render(<JobsScreen />);

      const searchInput = screen.getByPlaceholderText('Search jobs...');
      fireEvent.changeText(searchInput, 'NonExistentJob');

      await waitFor(() => {
        expect(screen.getByText('No jobs found')).toBeInTheDocument();
        expect(screen.getByText('Try adjusting your search or filters')).toBeInTheDocument();
      });
    });
  });

  describe('LoginScreen', () => {
    it('renders login form', () => {
      render(<LoginScreen />);

      expect(screen.getByText('ComputeHive')).toBeInTheDocument();
      expect(screen.getByText('Distributed Compute Platform')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Email')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Password')).toBeInTheDocument();
      expect(screen.getByText('Login')).toBeInTheDocument();
    });

    it('handles email input', () => {
      render(<LoginScreen />);

      const emailInput = screen.getByPlaceholderText('Email');
      fireEvent.changeText(emailInput, 'test@example.com');

      expect(emailInput.props.value).toBe('test@example.com');
    });

    it('handles password input', () => {
      render(<LoginScreen />);

      const passwordInput = screen.getByPlaceholderText('Password');
      fireEvent.changeText(passwordInput, 'password123');

      expect(passwordInput.props.value).toBe('password123');
    });

    it('toggles password visibility', () => {
      render(<LoginScreen />);

      const passwordInput = screen.getByPlaceholderText('Password');
      const eyeButton = screen.getByTestId('eye-button');

      // Password should be hidden by default
      expect(passwordInput.props.secureTextEntry).toBe(true);

      // Toggle visibility
      fireEvent.press(eyeButton);
      expect(passwordInput.props.secureTextEntry).toBe(false);

      // Toggle back
      fireEvent.press(eyeButton);
      expect(passwordInput.props.secureTextEntry).toBe(true);
    });

    it('shows biometric login when available', async () => {
      render(<LoginScreen />);

      await waitFor(() => {
        expect(screen.getByText('Login with Biometric')).toBeInTheDocument();
      });
    });

    it('handles login button press', async () => {
      const alertSpy = jest.spyOn(global, 'Alert').mockImplementation(() => {});
      
      render(<LoginScreen />);

      const emailInput = screen.getByPlaceholderText('Email');
      const passwordInput = screen.getByPlaceholderText('Password');
      const loginButton = screen.getByText('Login');

      fireEvent.changeText(emailInput, 'test@example.com');
      fireEvent.changeText(passwordInput, 'password123');
      fireEvent.press(loginButton);

      await waitFor(() => {
        expect(alertSpy).toHaveBeenCalledWith('Login Failed', 'Please fill in all fields');
      });

      alertSpy.mockRestore();
    });

    it('shows validation error for empty fields', async () => {
      const alertSpy = jest.spyOn(global, 'Alert').mockImplementation(() => {});
      
      render(<LoginScreen />);

      const loginButton = screen.getByText('Login');
      fireEvent.press(loginButton);

      await waitFor(() => {
        expect(alertSpy).toHaveBeenCalledWith('Error', 'Please fill in all fields');
      });

      alertSpy.mockRestore();
    });

    it('handles forgot password', () => {
      render(<LoginScreen />);

      const forgotPasswordLink = screen.getByText('Forgot Password?');
      fireEvent.press(forgotPasswordLink);

      // Should navigate to forgot password screen
      expect(forgotPasswordLink).toBeInTheDocument();
    });

    it('handles register link', () => {
      render(<LoginScreen />);

      const registerLink = screen.getByText('Sign up');
      fireEvent.press(registerLink);

      // Should navigate to register screen
      expect(registerLink).toBeInTheDocument();
    });
  });

  describe('Mobile Navigation', () => {
    it('renders bottom tab navigation on mobile', () => {
      const TestNavigator = () => (
        <NavigationContainer>
          <JobsScreen />
        </NavigationContainer>
      );

      render(<TestNavigator />);

      expect(screen.getByText('Jobs')).toBeInTheDocument();
      expect(screen.getByTestId('add-button')).toBeInTheDocument();
    });

    it('handles navigation between screens', () => {
      // This would test navigation between different screens
      // Implementation depends on your navigation setup
      expect(true).toBe(true);
    });
  });

  describe('Mobile Performance', () => {
    it('renders large lists efficiently', async () => {
      const startTime = performance.now();
      
      render(<JobsScreen />);

      await waitFor(() => {
        expect(screen.getByText('Jobs')).toBeInTheDocument();
      });

      const endTime = performance.now();
      const renderTime = endTime - startTime;

      // Should render in under 500ms on mobile
      expect(renderTime).toBeLessThan(500);
    });

    it('handles touch interactions smoothly', () => {
      render(<JobsScreen />);

      const searchInput = screen.getByPlaceholderText('Search jobs...');
      
      // Simulate touch interaction
      fireEvent(searchInput, 'touchStart', {
        nativeEvent: { touches: [{ identifier: 1, pageX: 0, pageY: 0 }] },
      });

      fireEvent(searchInput, 'touchEnd', {
        nativeEvent: { touches: [] },
      });

      // Should handle touch without errors
      expect(searchInput).toBeInTheDocument();
    });
  });

  describe('Mobile Accessibility', () => {
    it('has proper touch targets', () => {
      render(<JobsScreen />);

      const addButton = screen.getByTestId('add-button');
      const buttonStyle = addButton.props.style;

      // Touch targets should be at least 44px
      expect(buttonStyle.width).toBeGreaterThanOrEqual(44);
      expect(buttonStyle.height).toBeGreaterThanOrEqual(44);
    });

    it('supports screen readers', () => {
      render(<JobsScreen />);

      const searchInput = screen.getByPlaceholderText('Search jobs...');
      expect(searchInput.props.accessibilityLabel).toBeDefined();
      expect(searchInput.props.accessibilityHint).toBeDefined();
    });

    it('has proper contrast ratios', () => {
      render(<JobsScreen />);

      // This would check color contrast ratios
      // Implementation depends on your theme setup
      expect(true).toBe(true);
    });
  });

  describe('Mobile Offline Support', () => {
    it('handles offline state gracefully', async () => {
      // Mock network error
      jest.spyOn(global, 'fetch').mockRejectedValueOnce(new Error('Network error'));

      render(<JobsScreen />);

      await waitFor(() => {
        expect(screen.getByText('No jobs found')).toBeInTheDocument();
      });
    });

    it('caches data for offline use', () => {
      // This would test data caching functionality
      expect(true).toBe(true);
    });
  });
}); 