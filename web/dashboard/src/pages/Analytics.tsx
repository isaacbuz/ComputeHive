import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Grid,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Button,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  LinearProgress,
  Alert,
  Tabs,
  Tab,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Divider,
  Tooltip,
  IconButton
} from '@mui/material';
import {
  TrendingUp,
  TrendingDown,
  BarChart,
  PieChart,
  Timeline,
  Download,
  Refresh,
  ExpandMore,
  AttachMoney,
  Schedule,
  Memory,
  Speed,
  CheckCircle,
  Error,
  Warning,
  Info,
  CloudDownload,
  CloudUpload,
  Storage,
  NetworkCheck,
  Speed as SpeedIcon,
  Assessment,
  Analytics as AnalyticsIcon,
  Insights,
  DataUsage,
  ShowChart,
  Timeline as TimelineIcon
} from '@mui/icons-material';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  ResponsiveContainer,
  BarChart as RechartsBarChart,
  Bar,
  PieChart as RechartsPieChart,
  Pie,
  Cell,
  AreaChart,
  Area,
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  Radar,
  ComposedChart,
  Legend
} from 'recharts';

interface AnalyticsData {
  timeRange: string;
  jobs: JobAnalytics;
  resources: ResourceAnalytics;
  costs: CostAnalytics;
  performance: PerformanceAnalytics;
  trends: TrendData[];
}

interface JobAnalytics {
  total: number;
  completed: number;
  failed: number;
  running: number;
  pending: number;
  averageDuration: number;
  successRate: number;
  byType: { type: string; count: number }[];
  byStatus: { status: string; count: number }[];
  dailyStats: { date: string; completed: number; failed: number }[];
}

interface ResourceAnalytics {
  totalResources: number;
  available: number;
  allocated: number;
  utilization: number;
  byType: { type: string; count: number; utilization: number }[];
  topUtilized: { name: string; utilization: number; type: string }[];
  performance: { metric: string; value: number; trend: number }[];
}

interface CostAnalytics {
  totalSpent: number;
  thisMonth: number;
  lastMonth: number;
  averagePerJob: number;
  byResourceType: { type: string; cost: number }[];
  dailyCosts: { date: string; cost: number }[];
  costBreakdown: { category: string; amount: number; percentage: number }[];
}

interface PerformanceAnalytics {
  averageResponseTime: number;
  throughput: number;
  errorRate: number;
  availability: number;
  metrics: { timestamp: string; cpu: number; memory: number; gpu: number; network: number }[];
}

interface TrendData {
  date: string;
  jobs: number;
  costs: number;
  utilization: number;
  performance: number;
}

const Analytics: React.FC = () => {
  const [data, setData] = useState<AnalyticsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [timeRange, setTimeRange] = useState('30d');
  const [tabValue, setTabValue] = useState(0);
  const [refreshKey, setRefreshKey] = useState(0);

  // Mock data
  useEffect(() => {
    const generateMockData = (): AnalyticsData => {
      const now = new Date();
      const days = timeRange === '7d' ? 7 : timeRange === '30d' ? 30 : 90;
      
      const trendData: TrendData[] = Array.from({ length: days }, (_, i) => ({
        date: new Date(now.getTime() - (days - i - 1) * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        jobs: Math.floor(Math.random() * 50) + 10,
        costs: Math.random() * 100 + 20,
        utilization: Math.random() * 30 + 60,
        performance: Math.random() * 20 + 80
      }));

      const dailyStats = Array.from({ length: days }, (_, i) => ({
        date: new Date(now.getTime() - (days - i - 1) * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        completed: Math.floor(Math.random() * 30) + 5,
        failed: Math.floor(Math.random() * 5) + 1
      }));

      const dailyCosts = Array.from({ length: days }, (_, i) => ({
        date: new Date(now.getTime() - (days - i - 1) * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        cost: Math.random() * 50 + 10
      }));

      const metrics = Array.from({ length: 24 }, (_, i) => ({
        timestamp: new Date(now.getTime() - (23 - i) * 60 * 60 * 1000).toISOString(),
        cpu: Math.random() * 100,
        memory: Math.random() * 100,
        gpu: Math.random() * 100,
        network: Math.random() * 100
      }));

      return {
        timeRange,
        jobs: {
          total: 1250,
          completed: 1180,
          failed: 45,
          running: 15,
          pending: 10,
          averageDuration: 45.5,
          successRate: 94.4,
          byType: [
            { type: 'Batch Processing', count: 450 },
            { type: 'ML Training', count: 320 },
            { type: 'Data Analysis', count: 280 },
            { type: 'Real-time Processing', count: 200 }
          ],
          byStatus: [
            { status: 'Completed', count: 1180 },
            { status: 'Failed', count: 45 },
            { status: 'Running', count: 15 },
            { status: 'Pending', count: 10 }
          ],
          dailyStats
        },
        resources: {
          totalResources: 85,
          available: 45,
          allocated: 40,
          utilization: 47.1,
          byType: [
            { type: 'GPU', count: 25, utilization: 78.5 },
            { type: 'CPU', count: 40, utilization: 45.2 },
            { type: 'Storage', count: 15, utilization: 32.1 },
            { type: 'Network', count: 5, utilization: 28.4 }
          ],
          topUtilized: [
            { name: 'GPU-Cluster-01', utilization: 95.2, type: 'GPU' },
            { name: 'CPU-Farm-03', utilization: 87.6, type: 'CPU' },
            { name: 'Storage-Array-02', utilization: 82.1, type: 'Storage' },
            { name: 'GPU-Cluster-02', utilization: 79.8, type: 'GPU' }
          ],
          performance: [
            { metric: 'CPU Utilization', value: 45.2, trend: 2.1 },
            { metric: 'Memory Usage', value: 67.8, trend: -1.5 },
            { metric: 'GPU Utilization', value: 78.5, trend: 5.2 },
            { metric: 'Network Throughput', value: 34.2, trend: 0.8 }
          ]
        },
        costs: {
          totalSpent: 2847.50,
          thisMonth: 892.30,
          lastMonth: 756.80,
          averagePerJob: 2.28,
          byResourceType: [
            { type: 'GPU', cost: 1456.80 },
            { type: 'CPU', cost: 892.40 },
            { type: 'Storage', cost: 345.20 },
            { type: 'Network', cost: 153.10 }
          ],
          dailyCosts,
          costBreakdown: [
            { category: 'Compute Resources', amount: 2349.20, percentage: 82.5 },
            { category: 'Storage', amount: 345.20, percentage: 12.1 },
            { category: 'Network', amount: 153.10, percentage: 5.4 }
          ]
        },
        performance: {
          averageResponseTime: 2.3,
          throughput: 156.7,
          errorRate: 3.6,
          availability: 99.8,
          metrics
        },
        trends: trendData
      };
    };

    setLoading(true);
    setTimeout(() => {
      setData(generateMockData());
      setLoading(false);
    }, 1000);
  }, [timeRange, refreshKey]);

  const handleRefresh = () => {
    setRefreshKey(prev => prev + 1);
  };

  const getTrendIcon = (trend: number) => {
    if (trend > 0) return <TrendingUp color="success" />;
    if (trend < 0) return <TrendingDown color="error" />;
    return <Timeline color="info" />;
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'completed': return 'success';
      case 'failed': return 'error';
      case 'running': return 'warning';
      case 'pending': return 'info';
      default: return 'default';
    }
  };

  if (loading) {
    return (
      <Box sx={{ p: 3 }}>
        <LinearProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="error">{error}</Alert>
      </Box>
    );
  }

  if (!data) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="info">No analytics data available</Alert>
      </Box>
    );
  }

  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8'];

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          Analytics & Reports
        </Typography>
        <Box>
          <FormControl size="small" sx={{ mr: 2, minWidth: 120 }}>
            <InputLabel>Time Range</InputLabel>
            <Select
              value={timeRange}
              label="Time Range"
              onChange={(e) => setTimeRange(e.target.value)}
            >
              <MenuItem value="7d">Last 7 Days</MenuItem>
              <MenuItem value="30d">Last 30 Days</MenuItem>
              <MenuItem value="90d">Last 90 Days</MenuItem>
            </Select>
          </FormControl>
          <Button
            variant="outlined"
            startIcon={<Refresh />}
            onClick={handleRefresh}
            sx={{ mr: 1 }}
          >
            Refresh
          </Button>
          <Button
            variant="contained"
            startIcon={<Download />}
          >
            Export Report
          </Button>
        </Box>
      </Box>

      {/* Key Metrics Cards */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Total Jobs
                  </Typography>
                  <Typography variant="h4">{data.jobs.total.toLocaleString()}</Typography>
                  <Typography variant="body2" color="success.main">
                    +{((data.jobs.completed / data.jobs.total) * 100).toFixed(1)}% Success Rate
                  </Typography>
                </Box>
                <Assessment color="primary" sx={{ fontSize: 40 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Resource Utilization
                  </Typography>
                  <Typography variant="h4">{data.resources.utilization.toFixed(1)}%</Typography>
                  <Typography variant="body2" color="info.main">
                    {data.resources.allocated}/{data.resources.totalResources} Resources Active
                  </Typography>
                </Box>
                <DataUsage color="primary" sx={{ fontSize: 40 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Total Cost
                  </Typography>
                  <Typography variant="h4">${data.costs.totalSpent.toFixed(2)}</Typography>
                  <Typography variant="body2" color="success.main">
                    ${data.costs.averagePerJob.toFixed(2)} avg per job
                  </Typography>
                </Box>
                <AttachMoney color="primary" sx={{ fontSize: 40 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    System Availability
                  </Typography>
                  <Typography variant="h4">{data.performance.availability}%</Typography>
                  <Typography variant="body2" color="success.main">
                    {data.performance.averageResponseTime}s avg response
                  </Typography>
                </Box>
                <CheckCircle color="primary" sx={{ fontSize: 40 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Analytics Tabs */}
      <Card>
        <CardContent>
          <Tabs value={tabValue} onChange={(e, newValue) => setTabValue(newValue)}>
            <Tab label="Overview" icon={<ShowChart />} />
            <Tab label="Jobs" icon={<Assessment />} />
            <Tab label="Resources" icon={<Memory />} />
            <Tab label="Costs" icon={<AttachMoney />} />
            <Tab label="Performance" icon={<Speed />} />
            <Tab label="Trends" icon={<TimelineIcon />} />
          </Tabs>
          
          <Box sx={{ mt: 3 }}>
            {/* Overview Tab */}
            {tabValue === 0 && (
              <Grid container spacing={3}>
                <Grid item xs={12} md={8}>
                  <Typography variant="h6" gutterBottom>Performance Trends</Typography>
                  <ResponsiveContainer width="100%" height={400}>
                    <ComposedChart data={data.trends}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="date" />
                      <YAxis yAxisId="left" />
                      <YAxis yAxisId="right" orientation="right" />
                      <RechartsTooltip />
                      <Legend />
                      <Line
                        yAxisId="left"
                        type="monotone"
                        dataKey="jobs"
                        stroke="#8884d8"
                        name="Jobs"
                      />
                      <Bar
                        yAxisId="right"
                        dataKey="costs"
                        fill="#82ca9d"
                        name="Costs ($)"
                      />
                    </ComposedChart>
                  </ResponsiveContainer>
                </Grid>
                <Grid item xs={12} md={4}>
                  <Typography variant="h6" gutterBottom>Resource Distribution</Typography>
                  <ResponsiveContainer width="100%" height={300}>
                    <RechartsPieChart>
                      <Pie
                        data={data.resources.byType}
                        cx="50%"
                        cy="50%"
                        labelLine={false}
                        label={({ type, percentage }) => `${type} ${percentage}%`}
                        outerRadius={80}
                        fill="#8884d8"
                        dataKey="count"
                      >
                        {data.resources.byType.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <RechartsTooltip />
                    </RechartsPieChart>
                  </ResponsiveContainer>
                </Grid>
                
                <Grid item xs={12}>
                  <Typography variant="h6" gutterBottom>Key Insights</Typography>
                  <Grid container spacing={2}>
                    <Grid item xs={12} md={6}>
                      <Alert severity="success" icon={<CheckCircle />}>
                        <Typography variant="subtitle2">High Success Rate</Typography>
                        <Typography variant="body2">
                          Job success rate is {data.jobs.successRate}%, which is excellent for production workloads.
                        </Typography>
                      </Alert>
                    </Grid>
                    <Grid item xs={12} md={6}>
                      <Alert severity="info" icon={<Info />}>
                        <Typography variant="subtitle2">Resource Optimization</Typography>
                        <Typography variant="body2">
                          GPU utilization is at {data.resources.byType.find(r => r.type === 'GPU')?.utilization}%, 
                          indicating good resource usage.
                        </Typography>
                      </Alert>
                    </Grid>
                    <Grid item xs={12} md={6}>
                      <Alert severity="warning" icon={<Warning />}>
                        <Typography variant="subtitle2">Cost Management</Typography>
                        <Typography variant="body2">
                          Average cost per job is ${data.costs.averagePerJob.toFixed(2)}. 
                          Consider optimizing resource allocation to reduce costs.
                        </Typography>
                      </Alert>
                    </Grid>
                    <Grid item xs={12} md={6}>
                      <Alert severity="info" icon={<Insights />}>
                        <Typography variant="subtitle2">Performance Metrics</Typography>
                        <Typography variant="body2">
                          System availability is {data.performance.availability}% with 
                          {data.performance.averageResponseTime}s average response time.
                        </Typography>
                      </Alert>
                    </Grid>
                  </Grid>
                </Grid>
              </Grid>
            )}

            {/* Jobs Tab */}
            {tabValue === 1 && (
              <Grid container spacing={3}>
                <Grid item xs={12} md={6}>
                  <Typography variant="h6" gutterBottom>Job Status Distribution</Typography>
                  <ResponsiveContainer width="100%" height={300}>
                    <RechartsPieChart>
                      <Pie
                        data={data.jobs.byStatus}
                        cx="50%"
                        cy="50%"
                        labelLine={false}
                        label={({ status, percentage }) => `${status} ${percentage}%`}
                        outerRadius={80}
                        fill="#8884d8"
                        dataKey="count"
                      >
                        {data.jobs.byStatus.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <RechartsTooltip />
                    </RechartsPieChart>
                  </ResponsiveContainer>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Typography variant="h6" gutterBottom>Job Types</Typography>
                  <ResponsiveContainer width="100%" height={300}>
                    <RechartsBarChart data={data.jobs.byType}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="type" />
                      <YAxis />
                      <RechartsTooltip />
                      <Bar dataKey="count" fill="#8884d8" />
                    </RechartsBarChart>
                  </ResponsiveContainer>
                </Grid>
                <Grid item xs={12}>
                  <Typography variant="h6" gutterBottom>Daily Job Statistics</Typography>
                  <ResponsiveContainer width="100%" height={300}>
                    <AreaChart data={data.jobs.dailyStats}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="date" />
                      <YAxis />
                      <RechartsTooltip />
                      <Area type="monotone" dataKey="completed" stackId="1" stroke="#82ca9d" fill="#82ca9d" />
                      <Area type="monotone" dataKey="failed" stackId="1" stroke="#ff8042" fill="#ff8042" />
                    </AreaChart>
                  </ResponsiveContainer>
                </Grid>
              </Grid>
            )}

            {/* Resources Tab */}
            {tabValue === 2 && (
              <Grid container spacing={3}>
                <Grid item xs={12} md={6}>
                  <Typography variant="h6" gutterBottom>Resource Performance</Typography>
                  <ResponsiveContainer width="100%" height={300}>
                    <RadarChart data={data.resources.performance}>
                      <PolarGrid />
                      <PolarAngleAxis dataKey="metric" />
                      <PolarRadiusAxis />
                      <Radar
                        name="Current"
                        dataKey="value"
                        stroke="#8884d8"
                        fill="#8884d8"
                        fillOpacity={0.6}
                      />
                      <RechartsTooltip />
                    </RadarChart>
                  </ResponsiveContainer>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Typography variant="h6" gutterBottom>Top Utilized Resources</Typography>
                  <TableContainer component={Paper} variant="outlined">
                    <Table size="small">
                      <TableHead>
                        <TableRow>
                          <TableCell>Resource</TableCell>
                          <TableCell>Type</TableCell>
                          <TableCell align="right">Utilization</TableCell>
                        </TableRow>
                      </TableHead>
                      <TableBody>
                        {data.resources.topUtilized.map((resource) => (
                          <TableRow key={resource.name}>
                            <TableCell>{resource.name}</TableCell>
                            <TableCell>
                              <Chip label={resource.type} size="small" />
                            </TableCell>
                            <TableCell align="right">
                              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end' }}>
                                <Typography variant="body2" sx={{ mr: 1 }}>
                                  {resource.utilization.toFixed(1)}%
                                </Typography>
                                <LinearProgress
                                  variant="determinate"
                                  value={resource.utilization}
                                  sx={{ width: 60, height: 6, borderRadius: 3 }}
                                  color={resource.utilization > 80 ? 'error' : 'primary'}
                                />
                              </Box>
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </TableContainer>
                </Grid>
              </Grid>
            )}

            {/* Costs Tab */}
            {tabValue === 3 && (
              <Grid container spacing={3}>
                <Grid item xs={12} md={6}>
                  <Typography variant="h6" gutterBottom>Cost Breakdown</Typography>
                  <ResponsiveContainer width="100%" height={300}>
                    <RechartsPieChart>
                      <Pie
                        data={data.costs.costBreakdown}
                        cx="50%"
                        cy="50%"
                        labelLine={false}
                        label={({ category, percentage }) => `${category} ${percentage}%`}
                        outerRadius={80}
                        fill="#8884d8"
                        dataKey="amount"
                      >
                        {data.costs.costBreakdown.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <RechartsTooltip formatter={(value) => [`$${value}`, 'Amount']} />
                    </RechartsPieChart>
                  </ResponsiveContainer>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Typography variant="h6" gutterBottom>Daily Costs</Typography>
                  <ResponsiveContainer width="100%" height={300}>
                    <LineChart data={data.costs.dailyCosts}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="date" />
                      <YAxis />
                      <RechartsTooltip formatter={(value) => [`$${value}`, 'Cost']} />
                      <Line type="monotone" dataKey="cost" stroke="#8884d8" />
                    </LineChart>
                  </ResponsiveContainer>
                </Grid>
                <Grid item xs={12}>
                  <Typography variant="h6" gutterBottom>Cost by Resource Type</Typography>
                  <ResponsiveContainer width="100%" height={300}>
                    <RechartsBarChart data={data.costs.byResourceType}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="type" />
                      <YAxis />
                      <RechartsTooltip formatter={(value) => [`$${value}`, 'Cost']} />
                      <Bar dataKey="cost" fill="#82ca9d" />
                    </RechartsBarChart>
                  </ResponsiveContainer>
                </Grid>
              </Grid>
            )}

            {/* Performance Tab */}
            {tabValue === 4 && (
              <Grid container spacing={3}>
                <Grid item xs={12}>
                  <Typography variant="h6" gutterBottom>System Performance Metrics</Typography>
                  <ResponsiveContainer width="100%" height={400}>
                    <LineChart data={data.performance.metrics}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis
                        dataKey="timestamp"
                        tickFormatter={(value) => new Date(value).toLocaleTimeString()}
                      />
                      <YAxis />
                      <RechartsTooltip
                        labelFormatter={(value) => new Date(value).toLocaleString()}
                      />
                      <Legend />
                      <Line type="monotone" dataKey="cpu" stroke="#8884d8" name="CPU %" />
                      <Line type="monotone" dataKey="memory" stroke="#82ca9d" name="Memory %" />
                      <Line type="monotone" dataKey="gpu" stroke="#ffc658" name="GPU %" />
                      <Line type="monotone" dataKey="network" stroke="#ff7300" name="Network %" />
                    </LineChart>
                  </ResponsiveContainer>
                </Grid>
                
                <Grid item xs={12} md={6}>
                  <Card>
                    <CardContent>
                      <Typography variant="h6" gutterBottom>Performance Summary</Typography>
                      <List>
                        <ListItem>
                          <ListItemIcon><SpeedIcon /></ListItemIcon>
                          <ListItemText
                            primary="Average Response Time"
                            secondary={`${data.performance.averageResponseTime} seconds`}
                          />
                        </ListItem>
                        <ListItem>
                          <ListItemIcon><NetworkCheck /></ListItemIcon>
                          <ListItemText
                            primary="Throughput"
                            secondary={`${data.performance.throughput} jobs/hour`}
                          />
                        </ListItem>
                        <ListItem>
                          <ListItemIcon><Error /></ListItemIcon>
                          <ListItemText
                            primary="Error Rate"
                            secondary={`${data.performance.errorRate}%`}
                          />
                        </ListItem>
                        <ListItem>
                          <ListItemIcon><CheckCircle /></ListItemIcon>
                          <ListItemText
                            primary="System Availability"
                            secondary={`${data.performance.availability}%`}
                          />
                        </ListItem>
                      </List>
                    </CardContent>
                  </Card>
                </Grid>
              </Grid>
            )}

            {/* Trends Tab */}
            {tabValue === 5 && (
              <Grid container spacing={3}>
                <Grid item xs={12}>
                  <Typography variant="h6" gutterBottom>Trend Analysis</Typography>
                  <ResponsiveContainer width="100%" height={400}>
                    <ComposedChart data={data.trends}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="date" />
                      <YAxis yAxisId="left" />
                      <YAxis yAxisId="right" orientation="right" />
                      <RechartsTooltip />
                      <Legend />
                      <Line
                        yAxisId="left"
                        type="monotone"
                        dataKey="jobs"
                        stroke="#8884d8"
                        name="Jobs"
                      />
                      <Line
                        yAxisId="left"
                        type="monotone"
                        dataKey="utilization"
                        stroke="#82ca9d"
                        name="Utilization %"
                      />
                      <Bar
                        yAxisId="right"
                        dataKey="costs"
                        fill="#ffc658"
                        name="Costs ($)"
                      />
                    </ComposedChart>
                  </ResponsiveContainer>
                </Grid>
                
                <Grid item xs={12}>
                  <Typography variant="h6" gutterBottom>Trend Insights</Typography>
                  <Grid container spacing={2}>
                    <Grid item xs={12} md={4}>
                      <Alert severity="success">
                        <Typography variant="subtitle2">Job Volume</Typography>
                        <Typography variant="body2">
                          Job volume has increased by 15% over the selected period, 
                          indicating growing usage of the platform.
                        </Typography>
                      </Alert>
                    </Grid>
                    <Grid item xs={12} md={4}>
                      <Alert severity="info">
                        <Typography variant="subtitle2">Resource Efficiency</Typography>
                        <Typography variant="body2">
                          Resource utilization has remained stable around 60-70%, 
                          showing good resource management.
                        </Typography>
                      </Alert>
                    </Grid>
                    <Grid item xs={12} md={4}>
                      <Alert severity="warning">
                        <Typography variant="subtitle2">Cost Trends</Typography>
                        <Typography variant="body2">
                          Daily costs show some volatility. Consider implementing 
                          cost optimization strategies.
                        </Typography>
                      </Alert>
                    </Grid>
                  </Grid>
                </Grid>
              </Grid>
            )}
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
};

export default Analytics; 