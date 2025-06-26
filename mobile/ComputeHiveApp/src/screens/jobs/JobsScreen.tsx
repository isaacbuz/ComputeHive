import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  RefreshControl,
  Alert,
  TextInput,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Ionicons } from '@expo/vector-icons';

interface Job {
  id: string;
  name: string;
  status: 'running' | 'completed' | 'failed' | 'pending' | 'cancelled';
  type: string;
  priority: 'low' | 'medium' | 'high';
  progress: number;
  startTime: string;
  endTime?: string;
  earnings: number;
  resourceUsage: {
    cpu: number;
    memory: number;
    gpu: number;
  };
}

const JobsScreen: React.FC = () => {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [filteredJobs, setFilteredJobs] = useState<Job[]>([]);
  const [refreshing, setRefreshing] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [sortBy, setSortBy] = useState<'name' | 'status' | 'startTime' | 'earnings'>('startTime');

  const statusFilters = [
    { key: 'all', label: 'All', icon: 'list' },
    { key: 'running', label: 'Running', icon: 'play' },
    { key: 'completed', label: 'Completed', icon: 'checkmark' },
    { key: 'failed', label: 'Failed', icon: 'close' },
    { key: 'pending', label: 'Pending', icon: 'time' },
  ];

  useEffect(() => {
    loadJobs();
  }, []);

  useEffect(() => {
    filterAndSortJobs();
  }, [jobs, searchQuery, statusFilter, sortBy]);

  const loadJobs = async () => {
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const mockJobs: Job[] = [
        {
          id: '1',
          name: 'ML Model Training',
          status: 'running',
          type: 'Machine Learning',
          priority: 'high',
          progress: 75,
          startTime: '2024-01-15T10:30:00Z',
          earnings: 45.67,
          resourceUsage: { cpu: 85, memory: 60, gpu: 90 },
        },
        {
          id: '2',
          name: 'Data Processing Pipeline',
          status: 'completed',
          type: 'Data Processing',
          priority: 'medium',
          progress: 100,
          startTime: '2024-01-15T08:00:00Z',
          endTime: '2024-01-15T09:30:00Z',
          earnings: 23.45,
          resourceUsage: { cpu: 0, memory: 0, gpu: 0 },
        },
        {
          id: '3',
          name: 'Image Recognition',
          status: 'failed',
          type: 'Computer Vision',
          priority: 'high',
          progress: 30,
          startTime: '2024-01-15T07:00:00Z',
          endTime: '2024-01-15T07:45:00Z',
          earnings: 0,
          resourceUsage: { cpu: 0, memory: 0, gpu: 0 },
        },
        {
          id: '4',
          name: 'Blockchain Mining',
          status: 'pending',
          type: 'Cryptocurrency',
          priority: 'low',
          progress: 0,
          startTime: '2024-01-15T12:00:00Z',
          earnings: 0,
          resourceUsage: { cpu: 0, memory: 0, gpu: 0 },
        },
      ];
      
      setJobs(mockJobs);
    } catch (error) {
      Alert.alert('Error', 'Failed to load jobs');
    }
  };

  const filterAndSortJobs = () => {
    let filtered = jobs;

    // Apply search filter
    if (searchQuery) {
      filtered = filtered.filter(job =>
        job.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        job.type.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    // Apply status filter
    if (statusFilter !== 'all') {
      filtered = filtered.filter(job => job.status === statusFilter);
    }

    // Apply sorting
    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.name.localeCompare(b.name);
        case 'status':
          return a.status.localeCompare(b.status);
        case 'startTime':
          return new Date(b.startTime).getTime() - new Date(a.startTime).getTime();
        case 'earnings':
          return b.earnings - a.earnings;
        default:
          return 0;
      }
    });

    setFilteredJobs(filtered);
  };

  const onRefresh = async () => {
    setRefreshing(true);
    await loadJobs();
    setRefreshing(false);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
        return '#059669';
      case 'completed':
        return '#2563eb';
      case 'failed':
        return '#dc2626';
      case 'pending':
        return '#d97706';
      case 'cancelled':
        return '#6b7280';
      default:
        return '#6b7280';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running':
        return 'play-circle';
      case 'completed':
        return 'checkmark-circle';
      case 'failed':
        return 'close-circle';
      case 'pending':
        return 'time';
      case 'cancelled':
        return 'stop-circle';
      default:
        return 'help-circle';
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high':
        return '#dc2626';
      case 'medium':
        return '#d97706';
      case 'low':
        return '#059669';
      default:
        return '#6b7280';
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const renderJobItem = ({ item }: { item: Job }) => (
    <TouchableOpacity
      style={styles.jobCard}
      onPress={() => Alert.alert('Job Details', `Viewing details for ${item.name}`)}
      activeOpacity={0.7}
    >
      <View style={styles.jobHeader}>
        <View style={styles.jobTitleContainer}>
          <Text style={styles.jobName} numberOfLines={1}>
            {item.name}
          </Text>
          <View style={[styles.priorityBadge, { backgroundColor: getPriorityColor(item.priority) }]}>
            <Text style={styles.priorityText}>{item.priority.toUpperCase()}</Text>
          </View>
        </View>
        <View style={styles.statusContainer}>
          <Ionicons
            name={getStatusIcon(item.status) as any}
            size={20}
            color={getStatusColor(item.status)}
          />
          <Text style={[styles.statusText, { color: getStatusColor(item.status) }]}>
            {item.status.toUpperCase()}
          </Text>
        </View>
      </View>

      <View style={styles.jobDetails}>
        <Text style={styles.jobType}>{item.type}</Text>
        <Text style={styles.jobTime}>
          Started: {formatDate(item.startTime)}
        </Text>
        {item.endTime && (
          <Text style={styles.jobTime}>
            Ended: {formatDate(item.endTime)}
          </Text>
        )}
      </View>

      {item.status === 'running' && (
        <View style={styles.progressContainer}>
          <View style={styles.progressBar}>
            <View
              style={[
                styles.progressFill,
                { width: `${item.progress}%` },
              ]}
            />
          </View>
          <Text style={styles.progressText}>{item.progress}%</Text>
        </View>
      )}

      <View style={styles.jobFooter}>
        <View style={styles.resourceUsage}>
          <Text style={styles.resourceText}>CPU: {item.resourceUsage.cpu}%</Text>
          <Text style={styles.resourceText}>RAM: {item.resourceUsage.memory}%</Text>
          <Text style={styles.resourceText}>GPU: {item.resourceUsage.gpu}%</Text>
        </View>
        <View style={styles.earningsContainer}>
          <Ionicons name="cash" size={16} color="#d97706" />
          <Text style={styles.earningsText}>${item.earnings.toFixed(2)}</Text>
        </View>
      </View>
    </TouchableOpacity>
  );

  const renderStatusFilter = ({ item }: { item: typeof statusFilters[0] }) => (
    <TouchableOpacity
      style={[
        styles.filterButton,
        statusFilter === item.key && styles.filterButtonActive,
      ]}
      onPress={() => setStatusFilter(item.key)}
    >
      <Ionicons
        name={item.icon as any}
        size={16}
        color={statusFilter === item.key ? '#ffffff' : '#6b7280'}
      />
      <Text
        style={[
          styles.filterButtonText,
          statusFilter === item.key && styles.filterButtonTextActive,
        ]}
      >
        {item.label}
      </Text>
    </TouchableOpacity>
  );

  return (
    <SafeAreaView style={styles.container}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>Jobs</Text>
        <TouchableOpacity
          style={styles.addButton}
          onPress={() => Alert.alert('New Job', 'Create new job')}
        >
          <Ionicons name="add" size={24} color="#ffffff" />
        </TouchableOpacity>
      </View>

      {/* Search Bar */}
      <View style={styles.searchContainer}>
        <Ionicons name="search" size={20} color="#6b7280" />
        <TextInput
          style={styles.searchInput}
          placeholder="Search jobs..."
          value={searchQuery}
          onChangeText={setSearchQuery}
        />
        {searchQuery.length > 0 && (
          <TouchableOpacity onPress={() => setSearchQuery('')}>
            <Ionicons name="close-circle" size={20} color="#6b7280" />
          </TouchableOpacity>
        )}
      </View>

      {/* Status Filters */}
      <View style={styles.filtersContainer}>
        <FlatList
          data={statusFilters}
          renderItem={renderStatusFilter}
          keyExtractor={(item) => item.key}
          horizontal
          showsHorizontalScrollIndicator={false}
          contentContainerStyle={styles.filtersList}
        />
      </View>

      {/* Sort Options */}
      <View style={styles.sortContainer}>
        <Text style={styles.sortLabel}>Sort by:</Text>
        <TouchableOpacity
          style={styles.sortButton}
          onPress={() => {
            const options = ['name', 'status', 'startTime', 'earnings'];
            const currentIndex = options.indexOf(sortBy);
            const nextIndex = (currentIndex + 1) % options.length;
            setSortBy(options[nextIndex] as any);
          }}
        >
          <Text style={styles.sortButtonText}>
            {sortBy.charAt(0).toUpperCase() + sortBy.slice(1)}
          </Text>
          <Ionicons name="chevron-down" size={16} color="#6b7280" />
        </TouchableOpacity>
      </View>

      {/* Jobs List */}
      <FlatList
        data={filteredJobs}
        renderItem={renderJobItem}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.jobsList}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
        }
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Ionicons name="briefcase-outline" size={64} color="#9ca3af" />
            <Text style={styles.emptyText}>No jobs found</Text>
            <Text style={styles.emptySubtext}>
              Try adjusting your search or filters
            </Text>
          </View>
        }
      />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f9fafb',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingVertical: 16,
    backgroundColor: '#ffffff',
    borderBottomWidth: 1,
    borderBottomColor: '#e5e7eb',
  },
  headerTitle: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#111827',
  },
  addButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: '#2563eb',
    justifyContent: 'center',
    alignItems: 'center',
  },
  searchContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#ffffff',
    marginHorizontal: 20,
    marginTop: 16,
    marginBottom: 12,
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderRadius: 12,
    borderWidth: 1,
    borderColor: '#e5e7eb',
  },
  searchInput: {
    flex: 1,
    marginLeft: 12,
    fontSize: 16,
    color: '#111827',
  },
  filtersContainer: {
    marginBottom: 12,
  },
  filtersList: {
    paddingHorizontal: 20,
  },
  filterButton: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 8,
    marginRight: 8,
    borderRadius: 20,
    backgroundColor: '#f3f4f6',
  },
  filterButtonActive: {
    backgroundColor: '#2563eb',
  },
  filterButtonText: {
    marginLeft: 4,
    fontSize: 14,
    fontWeight: '500',
    color: '#6b7280',
  },
  filterButtonTextActive: {
    color: '#ffffff',
  },
  sortContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 20,
    marginBottom: 12,
  },
  sortLabel: {
    fontSize: 14,
    color: '#6b7280',
    marginRight: 8,
  },
  sortButton: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 12,
    paddingVertical: 6,
    backgroundColor: '#ffffff',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#e5e7eb',
  },
  sortButtonText: {
    fontSize: 14,
    fontWeight: '500',
    color: '#374151',
    marginRight: 4,
  },
  jobsList: {
    paddingHorizontal: 20,
    paddingBottom: 20,
  },
  jobCard: {
    backgroundColor: '#ffffff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: {
      width: 0,
      height: 2,
    },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  jobHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 12,
  },
  jobTitleContainer: {
    flex: 1,
    marginRight: 12,
  },
  jobName: {
    fontSize: 16,
    fontWeight: '600',
    color: '#111827',
    marginBottom: 4,
  },
  priorityBadge: {
    alignSelf: 'flex-start',
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 4,
  },
  priorityText: {
    fontSize: 10,
    fontWeight: '600',
    color: '#ffffff',
  },
  statusContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  statusText: {
    fontSize: 12,
    fontWeight: '600',
    marginLeft: 4,
  },
  jobDetails: {
    marginBottom: 12,
  },
  jobType: {
    fontSize: 14,
    color: '#6b7280',
    marginBottom: 4,
  },
  jobTime: {
    fontSize: 12,
    color: '#9ca3af',
  },
  progressContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 12,
  },
  progressBar: {
    flex: 1,
    height: 6,
    backgroundColor: '#e5e7eb',
    borderRadius: 3,
    marginRight: 12,
  },
  progressFill: {
    height: '100%',
    backgroundColor: '#2563eb',
    borderRadius: 3,
  },
  progressText: {
    fontSize: 12,
    fontWeight: '500',
    color: '#6b7280',
    minWidth: 30,
  },
  jobFooter: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  resourceUsage: {
    flexDirection: 'row',
  },
  resourceText: {
    fontSize: 12,
    color: '#6b7280',
    marginRight: 12,
  },
  earningsContainer: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  earningsText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#d97706',
    marginLeft: 4,
  },
  emptyContainer: {
    alignItems: 'center',
    paddingVertical: 60,
  },
  emptyText: {
    fontSize: 18,
    fontWeight: '600',
    color: '#6b7280',
    marginTop: 16,
  },
  emptySubtext: {
    fontSize: 14,
    color: '#9ca3af',
    marginTop: 4,
  },
});

export default JobsScreen; 