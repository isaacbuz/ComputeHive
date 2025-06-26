import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Grid,
  Chip,
  IconButton,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Alert,
  LinearProgress,
  Tooltip,
  Switch,
  FormControlLabel,
  Divider,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  Badge,
  Tabs,
  Tab,
  Accordion,
  AccordionSummary,
  AccordionDetails
} from '@mui/material';
import {
  Memory,
  Storage,
  Speed,
  Computer,
  Add,
  Edit,
  Delete,
  Refresh,
  Visibility,
  VisibilityOff,
  ExpandMore,
  Warning,
  CheckCircle,
  Error,
  Info,
  TrendingUp,
  TrendingDown,
  Schedule,
  LocationOn,
  AttachMoney,
  Star,
  StarBorder
} from '@mui/icons-material';
import { DataGrid, GridColDef, GridValueGetterParams } from '@mui/x-data-grid';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, ResponsiveContainer, BarChart, Bar, PieChart, Pie, Cell } from 'recharts';

interface Resource {
  id: string;
  name: string;
  type: 'cpu' | 'gpu' | 'storage' | 'network';
  status: 'available' | 'allocated' | 'maintenance' | 'offline';
  agent_id: string;
  agent_name: string;
  location: string;
  total_capacity: {
    cpu_cores?: number;
    memory_gb?: number;
    gpu_count?: number;
    gpu_memory_gb?: number;
    storage_gb?: number;
    bandwidth_mbps?: number;
  };
  allocated_capacity: {
    cpu_cores?: number;
    memory_gb?: number;
    gpu_count?: number;
    gpu_memory_gb?: number;
    storage_gb?: number;
    bandwidth_mbps?: number;
  };
  price_per_hour: number;
  reputation_score: number;
  uptime_percentage: number;
  last_heartbeat: string;
  tags: string[];
  capabilities: string[];
}

interface ResourceMetrics {
  timestamp: string;
  cpu_usage: number;
  memory_usage: number;
  gpu_usage: number;
  network_usage: number;
  temperature: number;
  power_consumption: number;
}

const Resources: React.FC = () => {
  const [resources, setResources] = useState<Resource[]>([]);
  const [selectedResource, setSelectedResource] = useState<Resource | null>(null);
  const [metrics, setMetrics] = useState<ResourceMetrics[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tabValue, setTabValue] = useState(0);
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [showMetricsDialog, setShowMetricsDialog] = useState(false);
  const [filterType, setFilterType] = useState<string>('all');
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [sortBy, setSortBy] = useState<string>('name');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');

  // Mock data
  useEffect(() => {
    const mockResources: Resource[] = [
      {
        id: '1',
        name: 'High-Performance GPU Cluster',
        type: 'gpu',
        status: 'available',
        agent_id: 'agent-001',
        agent_name: 'GPU-Farm-01',
        location: 'US East (N. Virginia)',
        total_capacity: {
          cpu_cores: 32,
          memory_gb: 128,
          gpu_count: 4,
          gpu_memory_gb: 24,
          storage_gb: 2000,
          bandwidth_mbps: 10000
        },
        allocated_capacity: {
          cpu_cores: 8,
          memory_gb: 32,
          gpu_count: 1,
          gpu_memory_gb: 6,
          storage_gb: 500,
          bandwidth_mbps: 2500
        },
        price_per_hour: 2.50,
        reputation_score: 4.8,
        uptime_percentage: 99.9,
        last_heartbeat: new Date().toISOString(),
        tags: ['gpu', 'ai', 'ml', 'high-performance'],
        capabilities: ['CUDA', 'TensorFlow', 'PyTorch', 'Docker']
      },
      {
        id: '2',
        name: 'CPU Compute Farm',
        type: 'cpu',
        status: 'allocated',
        agent_id: 'agent-002',
        agent_name: 'CPU-Farm-01',
        location: 'US West (Oregon)',
        total_capacity: {
          cpu_cores: 64,
          memory_gb: 256,
          storage_gb: 4000,
          bandwidth_mbps: 5000
        },
        allocated_capacity: {
          cpu_cores: 32,
          memory_gb: 128,
          storage_gb: 2000,
          bandwidth_mbps: 2500
        },
        price_per_hour: 1.20,
        reputation_score: 4.6,
        uptime_percentage: 99.5,
        last_heartbeat: new Date().toISOString(),
        tags: ['cpu', 'batch-processing', 'web-services'],
        capabilities: ['Docker', 'Kubernetes', 'Load Balancing']
      },
      {
        id: '3',
        name: 'Storage Array',
        type: 'storage',
        status: 'available',
        agent_id: 'agent-003',
        agent_name: 'Storage-01',
        location: 'Europe (Frankfurt)',
        total_capacity: {
          storage_gb: 10000,
          bandwidth_mbps: 2000
        },
        allocated_capacity: {
          storage_gb: 3000,
          bandwidth_mbps: 600
        },
        price_per_hour: 0.50,
        reputation_score: 4.9,
        uptime_percentage: 99.8,
        last_heartbeat: new Date().toISOString(),
        tags: ['storage', 'backup', 'archive'],
        capabilities: ['S3 Compatible', 'Backup', 'Archive']
      }
    ];

    const mockMetrics: ResourceMetrics[] = Array.from({ length: 24 }, (_, i) => ({
      timestamp: new Date(Date.now() - (23 - i) * 3600000).toISOString(),
      cpu_usage: Math.random() * 100,
      memory_usage: Math.random() * 100,
      gpu_usage: Math.random() * 100,
      network_usage: Math.random() * 100,
      temperature: 40 + Math.random() * 30,
      power_consumption: 200 + Math.random() * 300
    }));

    setResources(mockResources);
    setMetrics(mockMetrics);
    setLoading(false);
  }, []);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'available': return 'success';
      case 'allocated': return 'warning';
      case 'maintenance': return 'info';
      case 'offline': return 'error';
      default: return 'default';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'cpu': return <Computer />;
      case 'gpu': return <Memory />;
      case 'storage': return <Storage />;
      case 'network': return <Speed />;
      default: return <Computer />;
    }
  };

  const getUtilizationPercentage = (resource: Resource) => {
    const total = resource.total_capacity;
    const allocated = resource.allocated_capacity;
    
    if (resource.type === 'gpu') {
      return {
        cpu: (allocated.cpu_cores || 0) / (total.cpu_cores || 1) * 100,
        memory: (allocated.memory_gb || 0) / (total.memory_gb || 1) * 100,
        gpu: (allocated.gpu_count || 0) / (total.gpu_count || 1) * 100,
        storage: (allocated.storage_gb || 0) / (total.storage_gb || 1) * 100
      };
    } else if (resource.type === 'cpu') {
      return {
        cpu: (allocated.cpu_cores || 0) / (total.cpu_cores || 1) * 100,
        memory: (allocated.memory_gb || 0) / (total.memory_gb || 1) * 100,
        storage: (allocated.storage_gb || 0) / (total.storage_gb || 1) * 100
      };
    } else {
      return {
        storage: (allocated.storage_gb || 0) / (total.storage_gb || 1) * 100
      };
    }
  };

  const columns: GridColDef[] = [
    {
      field: 'name',
      headerName: 'Resource Name',
      width: 250,
      renderCell: (params) => (
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          {getTypeIcon(params.row.type)}
          <Box>
            <Typography variant="body2" fontWeight="bold">
              {params.value}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {params.row.agent_name}
            </Typography>
          </Box>
        </Box>
      )
    },
    {
      field: 'type',
      headerName: 'Type',
      width: 120,
      renderCell: (params) => (
        <Chip
          label={params.value.toUpperCase()}
          size="small"
          color={params.value === 'gpu' ? 'primary' : 'default'}
        />
      )
    },
    {
      field: 'status',
      headerName: 'Status',
      width: 130,
      renderCell: (params) => (
        <Chip
          label={params.value}
          size="small"
          color={getStatusColor(params.value) as any}
        />
      )
    },
    {
      field: 'utilization',
      headerName: 'Utilization',
      width: 200,
      renderCell: (params) => {
        const util = getUtilizationPercentage(params.row);
        const avgUtil = Object.values(util).reduce((a, b) => a + b, 0) / Object.values(util).length;
        
        return (
          <Box sx={{ width: '100%' }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
              <Typography variant="caption">{avgUtil.toFixed(1)}%</Typography>
            </Box>
            <LinearProgress
              variant="determinate"
              value={avgUtil}
              color={avgUtil > 80 ? 'error' : avgUtil > 60 ? 'warning' : 'success'}
              sx={{ height: 6, borderRadius: 3 }}
            />
          </Box>
        );
      }
    },
    {
      field: 'location',
      headerName: 'Location',
      width: 150,
      renderCell: (params) => (
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
          <LocationOn fontSize="small" />
          <Typography variant="body2">{params.value}</Typography>
        </Box>
      )
    },
    {
      field: 'price_per_hour',
      headerName: 'Price/Hour',
      width: 120,
      renderCell: (params) => (
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
          <AttachMoney fontSize="small" />
          <Typography variant="body2">${params.value}</Typography>
        </Box>
      )
    },
    {
      field: 'reputation_score',
      headerName: 'Reputation',
      width: 120,
      renderCell: (params) => (
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
          {params.value >= 4.5 ? <Star color="primary" /> : <StarBorder />}
          <Typography variant="body2">{params.value}</Typography>
        </Box>
      )
    },
    {
      field: 'actions',
      headerName: 'Actions',
      width: 150,
      sortable: false,
      renderCell: (params) => (
        <Box>
          <Tooltip title="View Metrics">
            <IconButton
              size="small"
              onClick={() => {
                setSelectedResource(params.row);
                setShowMetricsDialog(true);
              }}
            >
              <Visibility />
            </IconButton>
          </Tooltip>
          <Tooltip title="Edit Resource">
            <IconButton size="small">
              <Edit />
            </IconButton>
          </Tooltip>
          <Tooltip title="Delete Resource">
            <IconButton size="small" color="error">
              <Delete />
            </IconButton>
          </Tooltip>
        </Box>
      )
    }
  ];

  const filteredResources = resources.filter(resource => {
    if (filterType !== 'all' && resource.type !== filterType) return false;
    if (filterStatus !== 'all' && resource.status !== filterStatus) return false;
    return true;
  });

  const getResourceStats = () => {
    const total = resources.length;
    const available = resources.filter(r => r.status === 'available').length;
    const allocated = resources.filter(r => r.status === 'allocated').length;
    const maintenance = resources.filter(r => r.status === 'maintenance').length;
    const offline = resources.filter(r => r.status === 'offline').length;

    return { total, available, allocated, maintenance, offline };
  };

  const stats = getResourceStats();

  const resourceTypeData = [
    { name: 'GPU', value: resources.filter(r => r.type === 'gpu').length, color: '#8884d8' },
    { name: 'CPU', value: resources.filter(r => r.type === 'cpu').length, color: '#82ca9d' },
    { name: 'Storage', value: resources.filter(r => r.type === 'storage').length, color: '#ffc658' },
    { name: 'Network', value: resources.filter(r => r.type === 'network').length, color: '#ff7300' }
  ];

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

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          Resources
        </Typography>
        <Box>
          <Button
            variant="contained"
            startIcon={<Add />}
            onClick={() => setShowAddDialog(true)}
            sx={{ mr: 1 }}
          >
            Add Resource
          </Button>
          <Button
            variant="outlined"
            startIcon={<Refresh />}
            onClick={() => window.location.reload()}
          >
            Refresh
          </Button>
        </Box>
      </Box>

      {/* Stats Cards */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Total Resources
                  </Typography>
                  <Typography variant="h4">{stats.total}</Typography>
                </Box>
                <Computer color="primary" sx={{ fontSize: 40 }} />
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
                    Available
                  </Typography>
                  <Typography variant="h4" color="success.main">{stats.available}</Typography>
                </Box>
                <CheckCircle color="success" sx={{ fontSize: 40 }} />
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
                    Allocated
                  </Typography>
                  <Typography variant="h4" color="warning.main">{stats.allocated}</Typography>
                </Box>
                <Schedule color="warning" sx={{ fontSize: 40 }} />
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
                    Maintenance
                  </Typography>
                  <Typography variant="h4" color="info.main">{stats.maintenance}</Typography>
                </Box>
                <Warning color="info" sx={{ fontSize: 40 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Charts */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Resource Types Distribution
              </Typography>
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={resourceTypeData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                    outerRadius={80}
                    fill="#8884d8"
                    dataKey="value"
                  >
                    {resourceTypeData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <RechartsTooltip />
                </PieChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Resource Utilization Trends
              </Typography>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={metrics.slice(-12)}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis
                    dataKey="timestamp"
                    tickFormatter={(value) => new Date(value).toLocaleTimeString()}
                  />
                  <YAxis />
                  <RechartsTooltip
                    labelFormatter={(value) => new Date(value).toLocaleString()}
                  />
                  <Line type="monotone" dataKey="cpu_usage" stroke="#8884d8" name="CPU" />
                  <Line type="monotone" dataKey="memory_usage" stroke="#82ca9d" name="Memory" />
                  <Line type="monotone" dataKey="gpu_usage" stroke="#ffc658" name="GPU" />
                </LineChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Filters */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Grid container spacing={2} alignItems="center">
            <Grid item>
              <FormControl size="small" sx={{ minWidth: 120 }}>
                <InputLabel>Type</InputLabel>
                <Select
                  value={filterType}
                  label="Type"
                  onChange={(e) => setFilterType(e.target.value)}
                >
                  <MenuItem value="all">All Types</MenuItem>
                  <MenuItem value="cpu">CPU</MenuItem>
                  <MenuItem value="gpu">GPU</MenuItem>
                  <MenuItem value="storage">Storage</MenuItem>
                  <MenuItem value="network">Network</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item>
              <FormControl size="small" sx={{ minWidth: 120 }}>
                <InputLabel>Status</InputLabel>
                <Select
                  value={filterStatus}
                  label="Status"
                  onChange={(e) => setFilterStatus(e.target.value)}
                >
                  <MenuItem value="all">All Status</MenuItem>
                  <MenuItem value="available">Available</MenuItem>
                  <MenuItem value="allocated">Allocated</MenuItem>
                  <MenuItem value="maintenance">Maintenance</MenuItem>
                  <MenuItem value="offline">Offline</MenuItem>
                </Select>
              </FormControl>
            </Grid>
          </Grid>
        </CardContent>
      </Card>

      {/* Resources Table */}
      <Card>
        <CardContent>
          <DataGrid
            rows={filteredResources}
            columns={columns}
            pageSize={10}
            rowsPerPageOptions={[10, 25, 50]}
            disableSelectionOnClick
            autoHeight
            sx={{
              '& .MuiDataGrid-cell:focus': {
                outline: 'none',
              },
            }}
          />
        </CardContent>
      </Card>

      {/* Metrics Dialog */}
      <Dialog
        open={showMetricsDialog}
        onClose={() => setShowMetricsDialog(false)}
        maxWidth="lg"
        fullWidth
      >
        <DialogTitle>
          Resource Metrics - {selectedResource?.name}
        </DialogTitle>
        <DialogContent>
          {selectedResource && (
            <Box>
              <Tabs value={tabValue} onChange={(e, newValue) => setTabValue(newValue)}>
                <Tab label="Overview" />
                <Tab label="Performance" />
                <Tab label="Details" />
              </Tabs>
              
              <Box sx={{ mt: 2 }}>
                {tabValue === 0 && (
                  <Grid container spacing={3}>
                    <Grid item xs={12} md={6}>
                      <Typography variant="h6" gutterBottom>Resource Information</Typography>
                      <List dense>
                        <ListItem>
                          <ListItemIcon><Computer /></ListItemIcon>
                          <ListItemText
                            primary="Type"
                            secondary={selectedResource.type.toUpperCase()}
                          />
                        </ListItem>
                        <ListItem>
                          <ListItemIcon><LocationOn /></ListItemIcon>
                          <ListItemText
                            primary="Location"
                            secondary={selectedResource.location}
                          />
                        </ListItem>
                        <ListItem>
                          <ListItemIcon><AttachMoney /></ListItemIcon>
                          <ListItemText
                            primary="Price per Hour"
                            secondary={`$${selectedResource.price_per_hour}`}
                          />
                        </ListItem>
                        <ListItem>
                          <ListItemIcon><Star /></ListItemIcon>
                          <ListItemText
                            primary="Reputation Score"
                            secondary={selectedResource.reputation_score}
                          />
                        </ListItem>
                      </List>
                    </Grid>
                    <Grid item xs={12} md={6}>
                      <Typography variant="h6" gutterBottom>Utilization</Typography>
                      {Object.entries(getUtilizationPercentage(selectedResource)).map(([key, value]) => (
                        <Box key={key} sx={{ mb: 2 }}>
                          <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                            <Typography variant="body2" textTransform="capitalize">
                              {key.replace('_', ' ')}
                            </Typography>
                            <Typography variant="body2">{value.toFixed(1)}%</Typography>
                          </Box>
                          <LinearProgress
                            variant="determinate"
                            value={value}
                            color={value > 80 ? 'error' : value > 60 ? 'warning' : 'success'}
                          />
                        </Box>
                      ))}
                    </Grid>
                  </Grid>
                )}
                
                {tabValue === 1 && (
                  <Box>
                    <Typography variant="h6" gutterBottom>Performance Metrics (Last 24 Hours)</Typography>
                    <ResponsiveContainer width="100%" height={400}>
                      <LineChart data={metrics}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis
                          dataKey="timestamp"
                          tickFormatter={(value) => new Date(value).toLocaleTimeString()}
                        />
                        <YAxis />
                        <RechartsTooltip
                          labelFormatter={(value) => new Date(value).toLocaleString()}
                        />
                        <Line type="monotone" dataKey="cpu_usage" stroke="#8884d8" name="CPU Usage %" />
                        <Line type="monotone" dataKey="memory_usage" stroke="#82ca9d" name="Memory Usage %" />
                        <Line type="monotone" dataKey="gpu_usage" stroke="#ffc658" name="GPU Usage %" />
                        <Line type="monotone" dataKey="temperature" stroke="#ff7300" name="Temperature °C" />
                      </LineChart>
                    </ResponsiveContainer>
                  </Box>
                )}
                
                {tabValue === 2 && (
                  <Box>
                    <Typography variant="h6" gutterBottom>Detailed Information</Typography>
                    <Grid container spacing={3}>
                      <Grid item xs={12} md={6}>
                        <Typography variant="subtitle1" gutterBottom>Capabilities</Typography>
                        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                          {selectedResource.capabilities.map((capability) => (
                            <Chip key={capability} label={capability} size="small" />
                          ))}
                        </Box>
                      </Grid>
                      <Grid item xs={12} md={6}>
                        <Typography variant="subtitle1" gutterBottom>Tags</Typography>
                        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                          {selectedResource.tags.map((tag) => (
                            <Chip key={tag} label={tag} size="small" variant="outlined" />
                          ))}
                        </Box>
                      </Grid>
                      <Grid item xs={12}>
                        <Typography variant="subtitle1" gutterBottom>Capacity Details</Typography>
                        <TableContainer component={Paper} variant="outlined">
                          <Table size="small">
                            <TableHead>
                              <TableRow>
                                <TableCell>Resource</TableCell>
                                <TableCell align="right">Total</TableCell>
                                <TableCell align="right">Allocated</TableCell>
                                <TableCell align="right">Available</TableCell>
                                <TableCell align="right">Utilization</TableCell>
                              </TableRow>
                            </TableHead>
                            <TableBody>
                              {Object.entries(selectedResource.total_capacity).map(([key, total]) => {
                                const allocated = selectedResource.allocated_capacity[key as keyof typeof selectedResource.allocated_capacity] || 0;
                                const available = (total as number) - (allocated as number);
                                const utilization = ((allocated as number) / (total as number)) * 100;
                                
                                return (
                                  <TableRow key={key}>
                                    <TableCell component="th" scope="row">
                                      {key.replace('_', ' ').toUpperCase()}
                                    </TableCell>
                                    <TableCell align="right">{total}</TableCell>
                                    <TableCell align="right">{allocated}</TableCell>
                                    <TableCell align="right">{available}</TableCell>
                                    <TableCell align="right">
                                      <Chip
                                        label={`${utilization.toFixed(1)}%`}
                                        size="small"
                                        color={utilization > 80 ? 'error' : utilization > 60 ? 'warning' : 'success'}
                                      />
                                    </TableCell>
                                  </TableRow>
                                );
                              })}
                            </TableBody>
                          </Table>
                        </TableContainer>
                      </Grid>
                    </Grid>
                  </Box>
                )}
              </Box>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowMetricsDialog(false)}>Close</Button>
        </DialogActions>
      </Dialog>

      {/* Add Resource Dialog */}
      <Dialog
        open={showAddDialog}
        onClose={() => setShowAddDialog(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>Add New Resource</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Resource Name"
                variant="outlined"
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <FormControl fullWidth>
                <InputLabel>Resource Type</InputLabel>
                <Select label="Resource Type">
                  <MenuItem value="cpu">CPU</MenuItem>
                  <MenuItem value="gpu">GPU</MenuItem>
                  <MenuItem value="storage">Storage</MenuItem>
                  <MenuItem value="network">Network</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Price per Hour ($)"
                type="number"
                variant="outlined"
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Location"
                variant="outlined"
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Tags (comma-separated)"
                variant="outlined"
                placeholder="gpu, ai, ml, high-performance"
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowAddDialog(false)}>Cancel</Button>
          <Button variant="contained">Add Resource</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Resources; 