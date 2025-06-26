import React from 'react';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createDrawerNavigator } from '@react-navigation/drawer';
import Icon from 'react-native-vector-icons/MaterialIcons';
import { useTheme } from '../contexts/ThemeContext';
import { useAuth } from '../contexts/AuthContext';

// Import screens
import { SplashScreen } from '../screens/SplashScreen';
import { LoginScreen } from '../screens/auth/LoginScreen';
import { RegisterScreen } from '../screens/auth/RegisterScreen';
import { BiometricSetupScreen } from '../screens/auth/BiometricSetupScreen';
import { DashboardScreen } from '../screens/DashboardScreen';
import { JobsScreen } from '../screens/jobs/JobsScreen';
import { JobDetailsScreen } from '../screens/jobs/JobDetailsScreen';
import { CreateJobScreen } from '../screens/jobs/CreateJobScreen';
import { MarketplaceScreen } from '../screens/marketplace/MarketplaceScreen';
import { ResourceDetailsScreen } from '../screens/marketplace/ResourceDetailsScreen';
import { ResourcesScreen } from '../screens/resources/ResourcesScreen';
import { ResourceMonitorScreen } from '../screens/resources/ResourceMonitorScreen';
import { SettingsScreen } from '../screens/settings/SettingsScreen';
import { ProfileScreen } from '../screens/settings/ProfileScreen';
import { NotificationsScreen } from '../screens/settings/NotificationsScreen';

// Navigation types
export type RootStackParamList = {
  Splash: undefined;
  Auth: undefined;
  Main: undefined;
};

export type AuthStackParamList = {
  Login: undefined;
  Register: undefined;
  BiometricSetup: undefined;
};

export type MainTabParamList = {
  Dashboard: undefined;
  Jobs: undefined;
  Marketplace: undefined;
  Resources: undefined;
  Settings: undefined;
};

export type JobsStackParamList = {
  JobsList: undefined;
  JobDetails: { jobId: string };
  CreateJob: undefined;
};

export type MarketplaceStackParamList = {
  MarketplaceList: undefined;
  ResourceDetails: { resourceId: string };
};

export type ResourcesStackParamList = {
  ResourcesList: undefined;
  ResourceMonitor: { resourceId: string };
};

export type SettingsStackParamList = {
  SettingsList: undefined;
  Profile: undefined;
  Notifications: undefined;
};

const RootStack = createNativeStackNavigator<RootStackParamList>();
const AuthStack = createNativeStackNavigator<AuthStackParamList>();
const Tab = createBottomTabNavigator<MainTabParamList>();
const JobsStack = createNativeStackNavigator<JobsStackParamList>();
const MarketplaceStack = createNativeStackNavigator<MarketplaceStackParamList>();
const ResourcesStack = createNativeStackNavigator<ResourcesStackParamList>();
const SettingsStack = createNativeStackNavigator<SettingsStackParamList>();

// Auth Navigator
function AuthNavigator() {
  return (
    <AuthStack.Navigator
      screenOptions={{
        headerShown: false,
        animation: 'slide_from_right',
      }}
    >
      <AuthStack.Screen name="Login" component={LoginScreen} />
      <AuthStack.Screen name="Register" component={RegisterScreen} />
      <AuthStack.Screen name="BiometricSetup" component={BiometricSetupScreen} />
    </AuthStack.Navigator>
  );
}

// Jobs Navigator
function JobsNavigator() {
  const { theme } = useTheme();
  
  return (
    <JobsStack.Navigator
      screenOptions={{
        headerStyle: {
          backgroundColor: theme.colors.background,
        },
        headerTintColor: theme.colors.text,
      }}
    >
      <JobsStack.Screen 
        name="JobsList" 
        component={JobsScreen}
        options={{ title: 'Jobs' }}
      />
      <JobsStack.Screen 
        name="JobDetails" 
        component={JobDetailsScreen}
        options={{ title: 'Job Details' }}
      />
      <JobsStack.Screen 
        name="CreateJob" 
        component={CreateJobScreen}
        options={{ title: 'Create Job' }}
      />
    </JobsStack.Navigator>
  );
}

// Marketplace Navigator
function MarketplaceNavigator() {
  const { theme } = useTheme();
  
  return (
    <MarketplaceStack.Navigator
      screenOptions={{
        headerStyle: {
          backgroundColor: theme.colors.background,
        },
        headerTintColor: theme.colors.text,
      }}
    >
      <MarketplaceStack.Screen 
        name="MarketplaceList" 
        component={MarketplaceScreen}
        options={{ title: 'Marketplace' }}
      />
      <MarketplaceStack.Screen 
        name="ResourceDetails" 
        component={ResourceDetailsScreen}
        options={{ title: 'Resource Details' }}
      />
    </MarketplaceStack.Navigator>
  );
}

// Resources Navigator
function ResourcesNavigator() {
  const { theme } = useTheme();
  
  return (
    <ResourcesStack.Navigator
      screenOptions={{
        headerStyle: {
          backgroundColor: theme.colors.background,
        },
        headerTintColor: theme.colors.text,
      }}
    >
      <ResourcesStack.Screen 
        name="ResourcesList" 
        component={ResourcesScreen}
        options={{ title: 'Resources' }}
      />
      <ResourcesStack.Screen 
        name="ResourceMonitor" 
        component={ResourceMonitorScreen}
        options={{ title: 'Resource Monitor' }}
      />
    </ResourcesStack.Navigator>
  );
}

// Settings Navigator
function SettingsNavigator() {
  const { theme } = useTheme();
  
  return (
    <SettingsStack.Navigator
      screenOptions={{
        headerStyle: {
          backgroundColor: theme.colors.background,
        },
        headerTintColor: theme.colors.text,
      }}
    >
      <SettingsStack.Screen 
        name="SettingsList" 
        component={SettingsScreen}
        options={{ title: 'Settings' }}
      />
      <SettingsStack.Screen 
        name="Profile" 
        component={ProfileScreen}
        options={{ title: 'Profile' }}
      />
      <SettingsStack.Screen 
        name="Notifications" 
        component={NotificationsScreen}
        options={{ title: 'Notifications' }}
      />
    </SettingsStack.Navigator>
  );
}

// Main Tab Navigator
function MainNavigator() {
  const { theme } = useTheme();
  
  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        tabBarIcon: ({ focused, color, size }) => {
          let iconName: string;

          switch (route.name) {
            case 'Dashboard':
              iconName = 'dashboard';
              break;
            case 'Jobs':
              iconName = 'work';
              break;
            case 'Marketplace':
              iconName = 'store';
              break;
            case 'Resources':
              iconName = 'memory';
              break;
            case 'Settings':
              iconName = 'settings';
              break;
            default:
              iconName = 'circle';
          }

          return <Icon name={iconName} size={size} color={color} />;
        },
        tabBarActiveTintColor: theme.colors.primary,
        tabBarInactiveTintColor: theme.colors.textSecondary,
        tabBarStyle: {
          backgroundColor: theme.colors.background,
          borderTopColor: theme.colors.border,
        },
        headerShown: false,
      })}
    >
      <Tab.Screen name="Dashboard" component={DashboardScreen} />
      <Tab.Screen name="Jobs" component={JobsNavigator} />
      <Tab.Screen name="Marketplace" component={MarketplaceNavigator} />
      <Tab.Screen name="Resources" component={ResourcesNavigator} />
      <Tab.Screen name="Settings" component={SettingsNavigator} />
    </Tab.Navigator>
  );
}

// Root Navigator
export function AppNavigator() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <SplashScreen />;
  }

  return (
    <RootStack.Navigator screenOptions={{ headerShown: false }}>
      {isAuthenticated ? (
        <RootStack.Screen name="Main" component={MainNavigator} />
      ) : (
        <RootStack.Screen name="Auth" component={AuthNavigator} />
      )}
    </RootStack.Navigator>
  );
} 