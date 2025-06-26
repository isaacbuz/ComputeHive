import React from 'react';
import { createStackNavigator } from '@react-navigation/stack';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createDrawerNavigator } from '@react-navigation/drawer';
import { Ionicons } from '@expo/vector-icons';
import { useAuth } from '../contexts/AuthContext';
import { useResponsive } from '../hooks/useResponsive';

// Auth Screens
import LoginScreen from '../screens/auth/LoginScreen';
import RegisterScreen from '../screens/auth/RegisterScreen';
import ForgotPasswordScreen from '../screens/auth/ForgotPasswordScreen';

// Main Screens
import DashboardScreen from '../screens/DashboardScreen';
import JobsScreen from '../screens/jobs/JobsScreen';
import JobDetailsScreen from '../screens/jobs/JobDetailsScreen';
import CreateJobScreen from '../screens/jobs/CreateJobScreen';
import ResourcesScreen from '../screens/resources/ResourcesScreen';
import MarketplaceScreen from '../screens/marketplace/MarketplaceScreen';
import SettingsScreen from '../screens/settings/SettingsScreen';
import ProfileScreen from '../screens/settings/ProfileScreen';
import NotificationsScreen from '../screens/NotificationsScreen';

// Components
import CustomDrawerContent from '../components/navigation/CustomDrawerContent';
import LoadingScreen from '../components/common/LoadingScreen';

const Stack = createStackNavigator();
const Tab = createBottomTabNavigator();
const Drawer = createDrawerNavigator();

// Auth Navigator
const AuthNavigator = () => (
  <Stack.Navigator
    screenOptions={{
      headerShown: false,
      cardStyle: { backgroundColor: '#fff' },
    }}
  >
    <Stack.Screen name="Login" component={LoginScreen} />
    <Stack.Screen name="Register" component={RegisterScreen} />
    <Stack.Screen name="ForgotPassword" component={ForgotPasswordScreen} />
  </Stack.Navigator>
);

// Jobs Navigator
const JobsNavigator = () => (
  <Stack.Navigator
    screenOptions={{
      headerShown: false,
    }}
  >
    <Stack.Screen name="JobsList" component={JobsScreen} />
    <Stack.Screen name="JobDetails" component={JobDetailsScreen} />
    <Stack.Screen name="CreateJob" component={CreateJobScreen} />
  </Stack.Navigator>
);

// Resources Navigator
const ResourcesNavigator = () => (
  <Stack.Navigator
    screenOptions={{
      headerShown: false,
    }}
  >
    <Stack.Screen name="ResourcesList" component={ResourcesScreen} />
  </Stack.Navigator>
);

// Marketplace Navigator
const MarketplaceNavigator = () => (
  <Stack.Navigator
    screenOptions={{
      headerShown: false,
    }}
  >
    <Stack.Screen name="MarketplaceList" component={MarketplaceScreen} />
  </Stack.Navigator>
);

// Settings Navigator
const SettingsNavigator = () => (
  <Stack.Navigator
    screenOptions={{
      headerShown: false,
    }}
  >
    <Stack.Screen name="SettingsList" component={SettingsScreen} />
    <Stack.Screen name="Profile" component={ProfileScreen} />
  </Stack.Navigator>
);

// Bottom Tab Navigator
const TabNavigator = () => {
  const { isMobile } = useResponsive();

  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        tabBarIcon: ({ focused, color, size }) => {
          let iconName: keyof typeof Ionicons.glyphMap;

          switch (route.name) {
            case 'Dashboard':
              iconName = focused ? 'home' : 'home-outline';
              break;
            case 'Jobs':
              iconName = focused ? 'briefcase' : 'briefcase-outline';
              break;
            case 'Resources':
              iconName = focused ? 'server' : 'server-outline';
              break;
            case 'Marketplace':
              iconName = focused ? 'storefront' : 'storefront-outline';
              break;
            case 'Settings':
              iconName = focused ? 'settings' : 'settings-outline';
              break;
            default:
              iconName = 'help-outline';
          }

          return <Ionicons name={iconName} size={size} color={color} />;
        },
        tabBarActiveTintColor: '#2563eb',
        tabBarInactiveTintColor: '#6b7280',
        tabBarStyle: {
          height: isMobile ? 60 : 80,
          paddingBottom: isMobile ? 8 : 16,
          paddingTop: 8,
          backgroundColor: '#ffffff',
          borderTopWidth: 1,
          borderTopColor: '#e5e7eb',
        },
        tabBarLabelStyle: {
          fontSize: 12,
          fontWeight: '500',
        },
        headerShown: false,
      })}
    >
      <Tab.Screen 
        name="Dashboard" 
        component={DashboardScreen}
        options={{ tabBarLabel: 'Dashboard' }}
      />
      <Tab.Screen 
        name="Jobs" 
        component={JobsNavigator}
        options={{ tabBarLabel: 'Jobs' }}
      />
      <Tab.Screen 
        name="Resources" 
        component={ResourcesNavigator}
        options={{ tabBarLabel: 'Resources' }}
      />
      <Tab.Screen 
        name="Marketplace" 
        component={MarketplaceNavigator}
        options={{ tabBarLabel: 'Marketplace' }}
      />
      <Tab.Screen 
        name="Settings" 
        component={SettingsNavigator}
        options={{ tabBarLabel: 'Settings' }}
      />
    </Tab.Navigator>
  );
};

// Drawer Navigator (for tablet)
const DrawerNavigator = () => (
  <Drawer.Navigator
    drawerContent={(props) => <CustomDrawerContent {...props} />}
    screenOptions={{
      headerShown: false,
      drawerStyle: {
        width: 280,
      },
    }}
  >
    <Drawer.Screen name="MainTabs" component={TabNavigator} />
    <Drawer.Screen name="Notifications" component={NotificationsScreen} />
  </Drawer.Navigator>
);

// Main App Navigator
const AppNavigator: React.FC<{ onLayout: () => void }> = ({ onLayout }) => {
  const { isAuthenticated, isLoading } = useAuth();
  const { isTablet, isDesktop } = useResponsive();

  if (isLoading) {
    return <LoadingScreen />;
  }

  return (
    <Stack.Navigator
      screenOptions={{
        headerShown: false,
      }}
      onLayout={onLayout}
    >
      {!isAuthenticated ? (
        <Stack.Screen name="Auth" component={AuthNavigator} />
      ) : (
        <>
          {isTablet || isDesktop ? (
            <Stack.Screen name="Drawer" component={DrawerNavigator} />
          ) : (
            <Stack.Screen name="Tabs" component={TabNavigator} />
          )}
          <Stack.Screen name="Notifications" component={NotificationsScreen} />
        </>
      )}
    </Stack.Navigator>
  );
};

export default AppNavigator;
