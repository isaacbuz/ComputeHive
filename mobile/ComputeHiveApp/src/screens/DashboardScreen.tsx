import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  RefreshControl,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Ionicons } from '@expo/vector-icons';
import { LinearGradient } from 'expo-linear-gradient';
import { LineChart, BarChart } from 'react-native-chart-kit';
import { Dimensions } from 'react-native';

const { width: screenWidth } = Dimensions.get('window');

interface DashboardStats {
  activeJobs: number;
  completedJobs: number;
  failedJobs: number;
  totalEarnings: number;
  cpuUsage: number;
  memoryUsage: number;
  gpuUsage: number;
  networkUsage: number;
}

const DashboardScreen: React.FC = () => {
  const [stats, setStats] = useState<DashboardStats>({
    activeJobs: 0,
    completedJobs: 0,
    failedJobs: 0,
    totalEarnings: 0,
    cpuUsage: 0,
    memoryUsage: 0,
    gpuUsage: 0,
    networkUsage: 0,
  });
  const [refreshing, setRefreshing] = useState(false);
  const [performanceData, setPerformanceData] = useState({
    labels: ['1h', '2h', '3h', '4h', '5h', '6h'],
    datasets: [
      {
        data: [20, 45, 28, 80, 99, 43],
      },
    ],
  });

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      setStats({
        activeJobs: 3,
        completedJobs: 127,
        failedJobs: 2,
        totalEarnings: 245.67,
        cpuUsage: 45,
        memoryUsage: 62,
        gpuUsage: 78,
        networkUsage: 23,
      });
    } catch (error) {
      Alert.alert('Error', 'Failed to load dashboard data');
    }
  };

  const onRefresh = async () => {
    setRefreshing(true);
    await loadDashboardData();
    setRefreshing(false);
  };

  const StatCard = ({ title, value, icon, color, onPress }: {
    title: string;
    value: string | number;
    icon: string;
    color: string;
    onPress?: () => void;
  }) => (
    <TouchableOpacity
      style={[styles.statCard, { borderLeftColor: color }]}
      onPress={onPress}
      activeOpacity={0.7}
    >
      <View style={styles.statHeader}>
        <Ionicons name={icon as any} size={24} color={color} />
        <Text style={styles.statTitle}>{title}</Text>
      </View>
      <Text style={[styles.statValue, { color }]}>{value}</Text>
    </TouchableOpacity>
  );

  const ResourceCard = ({ title, usage, icon, color }: {
    title: string;
    usage: number;
    icon: string;
    color: string;
  }) => (
    <View style={styles.resourceCard}>
      <View style={styles.resourceHeader}>
        <Ionicons name={icon as any} size={20} color={color} />
        <Text style={styles.resourceTitle}>{title}</Text>
      </View>
      <View style={styles.progressContainer}>
        <View style={[styles.progressBar, { backgroundColor: '#e5e7eb' }]}>
          <View
            style={[
              styles.progressFill,
              { width: `${usage}%`, backgroundColor: color },
            ]}
          />
        </View>
        <Text style={styles.progressText}>{usage}%</Text>
      </View>
    </View>
  );

  const QuickAction = ({ title, icon, onPress }: {
    title: string;
    icon: string;
    onPress: () => void;
  }) => (
    <TouchableOpacity style={styles.quickAction} onPress={onPress}>
      <View style={styles.quickActionIcon}>
        <Ionicons name={icon as any} size={24} color="#2563eb" />
      </View>
      <Text style={styles.quickActionText}>{title}</Text>
    </TouchableOpacity>
  );

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView
        style={styles.scrollView}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
      >
        {/* Header */}
        <View style={styles.header}>
          <LinearGradient
            colors={['#2563eb', '#1d4ed8']}
            style={styles.headerGradient}
          >
            <Text style={styles.headerTitle}>Dashboard</Text>
            <Text style={styles.headerSubtitle}>Welcome back!</Text>
          </LinearGradient>
        </View>

        {/* Stats Grid */}
        <View style={styles.statsGrid}>
          <StatCard
            title="Active Jobs"
            value={stats.activeJobs}
            icon="briefcase"
            color="#059669"
          />
          <StatCard
            title="Completed"
            value={stats.completedJobs}
            icon="checkmark-circle"
            color="#2563eb"
          />
          <StatCard
            title="Failed"
            value={stats.failedJobs}
            icon="close-circle"
            color="#dc2626"
          />
          <StatCard
            title="Earnings"
            value={`$${stats.totalEarnings}`}
            icon="cash"
            color="#d97706"
          />
        </View>

        {/* Performance Chart */}
        <View style={styles.chartContainer}>
          <Text style={styles.sectionTitle}>Performance (6h)</Text>
          <LineChart
            data={performanceData}
            width={screenWidth - 40}
            height={220}
            chartConfig={{
              backgroundColor: '#ffffff',
              backgroundGradientFrom: '#ffffff',
              backgroundGradientTo: '#ffffff',
              decimalPlaces: 0,
              color: (opacity = 1) => `rgba(37, 99, 235, ${opacity})`,
              labelColor: (opacity = 1) => `rgba(107, 114, 128, ${opacity})`,
              style: {
                borderRadius: 16,
              },
              propsForDots: {
                r: '6',
                strokeWidth: '2',
                stroke: '#2563eb',
              },
            }}
            bezier
            style={styles.chart}
          />
        </View>

        {/* Resource Usage */}
        <View style={styles.resourceContainer}>
          <Text style={styles.sectionTitle}>Resource Usage</Text>
          <ResourceCard
            title="CPU"
            usage={stats.cpuUsage}
            icon="hardware-chip"
            color="#059669"
          />
          <ResourceCard
            title="Memory"
            usage={stats.memoryUsage}
            icon="desktop"
            color="#2563eb"
          />
          <ResourceCard
            title="GPU"
            usage={stats.gpuUsage}
            icon="game-controller"
            color="#7c3aed"
          />
          <ResourceCard
            title="Network"
            usage={stats.networkUsage}
            icon="wifi"
            color="#d97706"
          />
        </View>

        {/* Quick Actions */}
        <View style={styles.quickActionsContainer}>
          <Text style={styles.sectionTitle}>Quick Actions</Text>
          <View style={styles.quickActionsGrid}>
            <QuickAction
              title="New Job"
              icon="add-circle"
              onPress={() => Alert.alert('New Job', 'Create new job')}
            />
            <QuickAction
              title="Resources"
              icon="server"
              onPress={() => Alert.alert('Resources', 'View resources')}
            />
            <QuickAction
              title="Marketplace"
              icon="storefront"
              onPress={() => Alert.alert('Marketplace', 'Browse marketplace')}
            />
            <QuickAction
              title="Settings"
              icon="settings"
              onPress={() => Alert.alert('Settings', 'Open settings')}
            />
          </View>
        </View>

        {/* Recent Activity */}
        <View style={styles.activityContainer}>
          <Text style={styles.sectionTitle}>Recent Activity</Text>
          <View style={styles.activityItem}>
            <Ionicons name="checkmark-circle" size={20} color="#059669" />
            <Text style={styles.activityText}>Job "ML Training" completed</Text>
            <Text style={styles.activityTime}>2m ago</Text>
          </View>
          <View style={styles.activityItem}>
            <Ionicons name="add-circle" size={20} color="#2563eb" />
            <Text style={styles.activityText}>New job "Data Processing" started</Text>
            <Text style={styles.activityTime}>5m ago</Text>
          </View>
          <View style={styles.activityItem}>
            <Ionicons name="cash" size={20} color="#d97706" />
            <Text style={styles.activityText}>Earned $12.45 from completed job</Text>
            <Text style={styles.activityTime}>10m ago</Text>
          </View>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f9fafb',
  },
  scrollView: {
    flex: 1,
  },
  header: {
    marginBottom: 20,
  },
  headerGradient: {
    padding: 20,
    paddingTop: 40,
    paddingBottom: 30,
  },
  headerTitle: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#ffffff',
    marginBottom: 4,
  },
  headerSubtitle: {
    fontSize: 16,
    color: '#e0e7ff',
  },
  statsGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    paddingHorizontal: 20,
    marginBottom: 20,
  },
  statCard: {
    width: '48%',
    backgroundColor: '#ffffff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    borderLeftWidth: 4,
    shadowColor: '#000',
    shadowOffset: {
      width: 0,
      height: 2,
    },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  statHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  statTitle: {
    fontSize: 14,
    color: '#6b7280',
    marginLeft: 8,
  },
  statValue: {
    fontSize: 24,
    fontWeight: 'bold',
  },
  chartContainer: {
    backgroundColor: '#ffffff',
    marginHorizontal: 20,
    marginBottom: 20,
    borderRadius: 12,
    padding: 16,
    shadowColor: '#000',
    shadowOffset: {
      width: 0,
      height: 2,
    },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#111827',
    marginBottom: 16,
  },
  chart: {
    marginVertical: 8,
    borderRadius: 16,
  },
  resourceContainer: {
    backgroundColor: '#ffffff',
    marginHorizontal: 20,
    marginBottom: 20,
    borderRadius: 12,
    padding: 16,
    shadowColor: '#000',
    shadowOffset: {
      width: 0,
      height: 2,
    },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  resourceCard: {
    marginBottom: 16,
  },
  resourceHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  resourceTitle: {
    fontSize: 14,
    fontWeight: '500',
    color: '#374151',
    marginLeft: 8,
  },
  progressContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  progressBar: {
    flex: 1,
    height: 8,
    borderRadius: 4,
    marginRight: 12,
  },
  progressFill: {
    height: '100%',
    borderRadius: 4,
  },
  progressText: {
    fontSize: 12,
    fontWeight: '500',
    color: '#6b7280',
    minWidth: 30,
  },
  quickActionsContainer: {
    backgroundColor: '#ffffff',
    marginHorizontal: 20,
    marginBottom: 20,
    borderRadius: 12,
    padding: 16,
    shadowColor: '#000',
    shadowOffset: {
      width: 0,
      height: 2,
    },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  quickActionsGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
  },
  quickAction: {
    width: '25%',
    alignItems: 'center',
    paddingVertical: 12,
  },
  quickActionIcon: {
    width: 48,
    height: 48,
    borderRadius: 24,
    backgroundColor: '#eff6ff',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 8,
  },
  quickActionText: {
    fontSize: 12,
    fontWeight: '500',
    color: '#374151',
    textAlign: 'center',
  },
  activityContainer: {
    backgroundColor: '#ffffff',
    marginHorizontal: 20,
    marginBottom: 20,
    borderRadius: 12,
    padding: 16,
    shadowColor: '#000',
    shadowOffset: {
      width: 0,
      height: 2,
    },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  activityItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 8,
  },
  activityText: {
    flex: 1,
    fontSize: 14,
    color: '#374151',
    marginLeft: 12,
  },
  activityTime: {
    fontSize: 12,
    color: '#9ca3af',
  },
});

export default DashboardScreen; 