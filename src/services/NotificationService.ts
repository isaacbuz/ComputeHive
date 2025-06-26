import PushNotification, {
  Importance,
  PushNotificationObject,
  ReceivedNotification,
} from 'react-native-push-notification';
import messaging from '@react-native-firebase/messaging';
import { Platform } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';

interface NotificationData {
  title: string;
  message: string;
  data?: any;
  badge?: number;
  sound?: string;
}

class NotificationServiceClass {
  private fcmToken: string | null = null;

  async initialize(): Promise<void> {
    // Configure local notifications
    PushNotification.configure({
      onRegister: (token) => {
        console.log('Local notification token:', token);
      },

      onNotification: (notification: Omit<ReceivedNotification, 'userInfo'>) => {
        console.log('Notification received:', notification);
        
        // Handle notification tap
        if (notification.userInteraction) {
          this.handleNotificationOpen(notification);
        }

        // Required on iOS
        if (Platform.OS === 'ios') {
          notification.finish('backgroundFetchResultNoData');
        }
      },

      permissions: {
        alert: true,
        badge: true,
        sound: true,
      },

      popInitialNotification: true,
      requestPermissions: true,
    });

    // Create notification channels for Android
    if (Platform.OS === 'android') {
      this.createNotificationChannels();
    }

    // Initialize Firebase messaging
    await this.initializeFirebaseMessaging();
  }

  private createNotificationChannels(): void {
    // Job notifications channel
    PushNotification.createChannel(
      {
        channelId: 'jobs',
        channelName: 'Job Notifications',
        channelDescription: 'Notifications about job status and updates',
        playSound: true,
        soundName: 'default',
        importance: Importance.HIGH,
        vibrate: true,
      },
      (created) => console.log(`Job channel created: ${created}`)
    );

    // Marketplace notifications channel
    PushNotification.createChannel(
      {
        channelId: 'marketplace',
        channelName: 'Marketplace Notifications',
        channelDescription: 'Notifications about marketplace activities',
        playSound: true,
        soundName: 'default',
        importance: Importance.DEFAULT,
        vibrate: true,
      },
      (created) => console.log(`Marketplace channel created: ${created}`)
    );

    // System notifications channel
    PushNotification.createChannel(
      {
        channelId: 'system',
        channelName: 'System Notifications',
        channelDescription: 'System alerts and important updates',
        playSound: true,
        soundName: 'default',
        importance: Importance.HIGH,
        vibrate: true,
      },
      (created) => console.log(`System channel created: ${created}`)
    );
  }

  private async initializeFirebaseMessaging(): Promise<void> {
    // Request permission
    const authStatus = await messaging().requestPermission();
    const enabled =
      authStatus === messaging.AuthorizationStatus.AUTHORIZED ||
      authStatus === messaging.AuthorizationStatus.PROVISIONAL;

    if (enabled) {
      console.log('Authorization status:', authStatus);
      
      // Get FCM token
      await this.getFCMToken();

      // Handle token refresh
      messaging().onTokenRefresh(async (token) => {
        console.log('FCM Token refreshed:', token);
        this.fcmToken = token;
        await this.saveFCMToken(token);
      });

      // Handle foreground messages
      messaging().onMessage(async (remoteMessage) => {
        console.log('Foreground message received:', remoteMessage);
        this.showLocalNotification({
          title: remoteMessage.notification?.title || 'ComputeHive',
          message: remoteMessage.notification?.body || '',
          data: remoteMessage.data,
        });
      });

      // Handle background message
      messaging().setBackgroundMessageHandler(async (remoteMessage) => {
        console.log('Background message received:', remoteMessage);
        // Handle background message (e.g., update badge count)
      });
    }
  }

  async getFCMToken(): Promise<string | null> {
    try {
      if (this.fcmToken) {
        return this.fcmToken;
      }

      const token = await messaging().getToken();
      if (token) {
        this.fcmToken = token;
        await this.saveFCMToken(token);
        console.log('FCM Token:', token);
        return token;
      }
    } catch (error) {
      console.error('Failed to get FCM token:', error);
    }
    return null;
  }

  private async saveFCMToken(token: string): Promise<void> {
    try {
      await AsyncStorage.setItem('fcm_token', token);
      // TODO: Send token to backend server
    } catch (error) {
      console.error('Failed to save FCM token:', error);
    }
  }

  showLocalNotification(notification: NotificationData): void {
    const notificationObject: PushNotificationObject = {
      channelId: this.getChannelId(notification.data?.type),
      title: notification.title,
      message: notification.message,
      playSound: true,
      soundName: notification.sound || 'default',
      number: notification.badge,
      data: notification.data,
    };

    if (Platform.OS === 'android') {
      notificationObject.largeIcon = 'ic_launcher';
      notificationObject.smallIcon = 'ic_notification';
      notificationObject.color = '#1976d2';
    }

    PushNotification.localNotification(notificationObject);
  }

  scheduleNotification(notification: NotificationData, date: Date): void {
    PushNotification.localNotificationSchedule({
      channelId: this.getChannelId(notification.data?.type),
      title: notification.title,
      message: notification.message,
      date,
      playSound: true,
      soundName: notification.sound || 'default',
      data: notification.data,
    });
  }

  private getChannelId(type?: string): string {
    switch (type) {
      case 'job':
        return 'jobs';
      case 'marketplace':
        return 'marketplace';
      case 'system':
        return 'system';
      default:
        return 'jobs';
    }
  }

  private handleNotificationOpen(notification: any): void {
    console.log('Notification opened:', notification);
    
    // Navigate based on notification type
    const { data } = notification;
    if (data) {
      switch (data.type) {
        case 'job':
          // Navigate to job details
          if (data.jobId) {
            // NavigationService.navigate('JobDetails', { jobId: data.jobId });
          }
          break;
        case 'marketplace':
          // Navigate to marketplace
          // NavigationService.navigate('Marketplace');
          break;
        default:
          // Navigate to dashboard
          // NavigationService.navigate('Dashboard');
          break;
      }
    }
  }

  // Badge management
  async setBadgeCount(count: number): Promise<void> {
    PushNotification.setApplicationIconBadgeNumber(count);
  }

  async getBadgeCount(): Promise<number> {
    return new Promise((resolve) => {
      PushNotification.getApplicationIconBadgeNumber((number) => {
        resolve(number);
      });
    });
  }

  async clearBadge(): Promise<void> {
    PushNotification.setApplicationIconBadgeNumber(0);
  }

  // Cancel notifications
  cancelAllNotifications(): void {
    PushNotification.cancelAllLocalNotifications();
  }

  cancelNotification(id: string): void {
    PushNotification.cancelLocalNotification(id);
  }

  // Permission check
  async checkPermissions(): Promise<boolean> {
    return new Promise((resolve) => {
      PushNotification.checkPermissions((permissions) => {
        resolve(permissions.alert && permissions.badge && permissions.sound);
      });
    });
  }

  async requestPermissions(): Promise<void> {
    PushNotification.requestPermissions();
  }
}

export const NotificationService = new NotificationServiceClass(); 