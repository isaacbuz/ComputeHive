import React from 'react';
import {
  Grid,
  Paper,
  Typography,
  Box,
  Card,
  CardContent,
  LinearProgress,
  Chip,
  List,
  ListItem,
  ListItemText,
  ListItemAvatar,
  Avatar,
  IconButton,
  Tooltip,
} from '@mui/material';
import {
  TrendingUp as TrendingUpIcon,
  Computer as ComputerIcon,
  Work as WorkIcon,
  AttachMoney as MoneyIcon,
  Speed as SpeedIcon,
  Memory as MemoryIcon,
  Storage as StorageIcon,
  NetworkCheck as NetworkIcon,
  MoreVert as MoreVertIcon,
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Schedule as ScheduleIcon,
} from '@mui/icons-material';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { useQuery } from '@tanstack/react-query';
import axios from 'axios';
import { format } from 'date-fns';

// Interfaces
interface DashboardStats {
  totalJobs: number;
  activeJobs: number;
  completedJobs: number;
  failedJobs: number;
  totalAgents: number;
  activeAgents: number;
  totalEarnings: number;
  pendingEarnings: number;
  cpuUtilization: number;
  memoryUtilization: number;
  storageUtilization: number;
  networkUtilization: number;
}

interface RecentJob {
  id: string;
  type: string;
  status: string;
  createdAt: string;
  completedAt?: string;
  cost: number;
}

interface ResourceUsage {
  timestamp: string;
  cpu: number;
  memory: number;
  network: number;
}

// Mock data for charts
const resourceUsageData = Array.from({ length: 24 }, (_, i) => ({
  timestamp: format(new Date().setHours(new Date().getHours() - (23 - i)), 'HH:mm'),
  cpu: Math.random() * 100,
  memory: Math.random() * 100,
  network: Math.random() * 50,
}));

const jobDistributionData = [
  { name: 'Completed', value: 75, color: '#4caf50' },
  { name: 'Running', value: 15, color: '#2196f3' },
  { name: 'Failed', value: 10, color: '#f44336' },
];

const StatCard: React.FC<{
  title: string;
  value: string | number;
  subtitle?: string;
  icon: React.ReactNode;
  color: string;
  trend?: number;
}> = ({ title, value, subtitle, icon, color, trend }) => (
  <Card>
    <CardContent>
      <Box display="flex" justifyContent="space-between" alignItems="flex-start">
        <Box>
          <Typography color="textSecondary" gutterBottom variant="body2">
            {title}
          </Typography>
          <Typography variant="h4" component="div" sx={{ mb: 1 }}>
            {value}
          </Typography>
          {subtitle && (
            <Typography variant="body2" color="textSecondary">
              {subtitle}
            </Typography>
          )}
          {trend !== undefined && (
            <Box display="flex" alignItems="center" mt={1}>
              <TrendingUpIcon
                fontSize="small"
                sx={{ color: trend > 0 ? 'success.main' : 'error.main' }}
              />
              <Typography
                variant="body2"
                sx={{ color: trend > 0 ? 'success.main' : 'error.main', ml: 0.5 }}
              >
                {Math.abs(trend)}%
              </Typography>
            </Box>
          )}
        </Box>
        <Avatar sx={{ bgcolor: color, width: 56, height: 56 }}>
          {icon}
        </Avatar>
      </Box>
    </CardContent>
  </Card>
);

const ResourceCard: React.FC<{
  title: string;
  usage: number;
  icon: React.ReactNode;
  color: string;
}> = ({ title, usage, icon, color }) => (
  <Card>
    <CardContent>
      <Box display="flex" alignItems="center" mb={2}>
        <Avatar sx={{ bgcolor: color, width: 40, height: 40, mr: 2 }}>
          {icon}
        </Avatar>
        <Typography variant="h6">{title}</Typography>
      </Box>
      <Box>
        <Box display="flex" justifyContent="space-between" mb={1}>
          <Typography variant="body2" color="textSecondary">
            Usage
          </Typography>
          <Typography variant="body2" fontWeight="bold">
            {usage}%
          </Typography>
        </Box>
        <LinearProgress
          variant="determinate"
          value={usage}
          sx={{
            height: 8,
            borderRadius: 4,
            backgroundColor: 'action.hover',
            '& .MuiLinearProgress-bar': {
              backgroundColor: color,
              borderRadius: 4,
            },
          }}
        />
      </Box>
    </CardContent>
  </Card>
);

export default function Dashboard() {
  // Fetch dashboard stats
  const { data: stats, isLoading: statsLoading } = useQuery<DashboardStats>({
    queryKey: ['dashboard-stats'],
    queryFn: async () => {
      const response = await axios.get('/api/v1/dashboard/stats');
      return response.data;
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  // Fetch recent jobs
  const { data: recentJobs } = useQuery<RecentJob[]>({
    queryKey: ['recent-jobs'],
    queryFn: async () => {
      const response = await axios.get('/api/v1/jobs?limit=5&sort=createdAt:desc');
      return response.data;
    },
  });

  const getJobStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircleIcon sx={{ color: 'success.main' }} />;
      case 'failed':
        return <ErrorIcon sx={{ color: 'error.main' }} />;
      case 'running':
        return <ScheduleIcon sx={{ color: 'info.main' }} />;
      default:
        return <ScheduleIcon sx={{ color: 'text.secondary' }} />;
    }
  };

  if (statsLoading) {
    return (
      <Box sx={{ width: '100%', mt: 4 }}>
        <LinearProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Dashboard
      </Typography>
      
      {/* Stats Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Total Jobs"
            value={stats?.totalJobs || 0}
            subtitle={`${stats?.activeJobs || 0} active`}
            icon={<WorkIcon />}
            color="#2196f3"
            trend={12}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Active Agents"
            value={stats?.activeAgents || 0}
            subtitle={`of ${stats?.totalAgents || 0} total`}
            icon={<ComputerIcon />}
            color="#4caf50"
            trend={5}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Total Earnings"
            value={`$${stats?.totalEarnings?.toFixed(2) || '0.00'}`}
            subtitle={`$${stats?.pendingEarnings?.toFixed(2) || '0.00'} pending`}
            icon={<MoneyIcon />}
            color="#ff9800"
            trend={23}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Success Rate"
            value={`${stats ? ((stats.completedJobs / stats.totalJobs) * 100).toFixed(1) : 0}%`}
            subtitle="Last 30 days"
            icon={<TrendingUpIcon />}
            color="#9c27b0"
            trend={-2}
          />
        </Grid>
      </Grid>

      {/* Resource Usage */}
      <Typography variant="h5" gutterBottom sx={{ mt: 4, mb: 2 }}>
        Resource Utilization
      </Typography>
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <ResourceCard
            title="CPU"
            usage={stats?.cpuUtilization || 0}
            icon={<SpeedIcon />}
            color="#2196f3"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <ResourceCard
            title="Memory"
            usage={stats?.memoryUtilization || 0}
            icon={<MemoryIcon />}
            color="#4caf50"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <ResourceCard
            title="Storage"
            usage={stats?.storageUtilization || 0}
            icon={<StorageIcon />}
            color="#ff9800"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <ResourceCard
            title="Network"
            usage={stats?.networkUtilization || 0}
            icon={<NetworkIcon />}
            color="#9c27b0"
          />
        </Grid>
      </Grid>

      {/* Charts */}
      <Grid container spacing={3}>
        <Grid item xs={12} md={8}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="h6" gutterBottom>
              Resource Usage Over Time
            </Typography>
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={resourceUsageData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="timestamp" />
                <YAxis />
                <RechartsTooltip />
                <Legend />
                <Area
                  type="monotone"
                  dataKey="cpu"
                  stackId="1"
                  stroke="#2196f3"
                  fill="#2196f3"
                  fillOpacity={0.6}
                />
                <Area
                  type="monotone"
                  dataKey="memory"
                  stackId="1"
                  stroke="#4caf50"
                  fill="#4caf50"
                  fillOpacity={0.6}
                />
                <Area
                  type="monotone"
                  dataKey="network"
                  stackId="1"
                  stroke="#ff9800"
                  fill="#ff9800"
                  fillOpacity={0.6}
                />
              </AreaChart>
            </ResponsiveContainer>
          </Paper>
        </Grid>
        
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 2 }}>
            <Typography variant="h6" gutterBottom>
              Job Distribution
            </Typography>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={jobDistributionData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {jobDistributionData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <RechartsTooltip />
              </PieChart>
            </ResponsiveContainer>
          </Paper>
        </Grid>
      </Grid>

      {/* Recent Jobs */}
      <Paper sx={{ mt: 3, p: 2 }}>
        <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
          <Typography variant="h6">Recent Jobs</Typography>
          <Tooltip title="View all jobs">
            <IconButton size="small">
              <MoreVertIcon />
            </IconButton>
          </Tooltip>
        </Box>
        <List>
          {recentJobs?.map((job, index) => (
            <ListItem
              key={job.id}
              divider={index < recentJobs.length - 1}
              secondaryAction={
                <Chip
                  label={`$${job.cost.toFixed(2)}`}
                  size="small"
                  color="primary"
                  variant="outlined"
                />
              }
            >
              <ListItemAvatar>
                {getJobStatusIcon(job.status)}
              </ListItemAvatar>
              <ListItemText
                primary={
                  <Box display="flex" alignItems="center" gap={1}>
                    <Typography variant="body1">{job.type}</Typography>
                    <Chip
                      label={job.status}
                      size="small"
                      color={
                        job.status === 'completed'
                          ? 'success'
                          : job.status === 'failed'
                          ? 'error'
                          : 'default'
                      }
                    />
                  </Box>
                }
                secondary={`Created ${format(new Date(job.createdAt), 'MMM d, HH:mm')}`}
              />
            </ListItem>
          ))}
        </List>
      </Paper>
    </Box>
  );
} 