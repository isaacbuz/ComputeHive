import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Grid,
  Card,
  CardContent,
  Button,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Chip,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Alert,
  LinearProgress,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tooltip,
  Badge,
  Tab,
  Tabs,
} from '@mui/material';
import {
  ShoppingCart,
  LocalOffer,
  TrendingUp,
  TrendingDown,
  Timer,
  Speed,
  Memory,
  Storage,
  AttachMoney,
  CheckCircle,
  Cancel,
  Info,
  Refresh,
  FilterList,
} from '@mui/icons-material';
import { DataGrid, GridColDef } from '@mui/x-data-grid';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  ResponsiveContainer,
  Legend,
  PieChart,
  Pie,
  Cell,
} from 'recharts';
import { useAuth } from '../contexts/AuthContext';

interface MarketOffer {
  id: string;
  provider_id: string;
  provider_name: string;
  resource_type: string;
  cpu_cores: number;
  memory_gb: number;
  gpu_count: number;
  gpu_model: string;
  storage_gb: number;
  network_bandwidth_gbps: number;
  price_per_hour: number;
  availability: string;
  location: string;
  reputation_score: number;
  sla_uptime: number;
  created_at: string;
  expires_at: string;
  status: string;
}

interface MarketBid {
  id: string;
  consumer_id: string;
  consumer_name: string;
  resource_requirements: {
    cpu_cores: number;
    memory_gb: number;
    gpu_count: number;
    gpu_model?: string;
    storage_gb: number;
  };
  max_price_per_hour: number;
  duration_hours: number;
  deadline: string;
  created_at: string;
  status: string;
  matched_offer_id?: string;
}

interface MarketStats {
  active_offers: number;
  active_bids: number;
  avg_cpu_price: number;
  avg_gpu_price: number;
  total_capacity: {
    cpu_cores: number;
    memory_gb: number;
    gpu_count: number;
  };
  price_trends: Array<{
    time: string;
    cpu_price: number;
    gpu_price: number;
  }>;
}

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`marketplace-tabpanel-${index}`}
      aria-labelledby={`marketplace-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

export default function Marketplace() {
  const { user } = useAuth();
  const [tabValue, setTabValue] = useState(0);
  const [offers, setOffers] = useState<MarketOffer[]>([]);
  const [bids, setBids] = useState<MarketBid[]>([]);
  const [stats, setStats] = useState<MarketStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [createOfferOpen, setCreateOfferOpen] = useState(false);
  const [createBidOpen, setCreateBidOpen] = useState(false);
  const [selectedOffer, setSelectedOffer] = useState<MarketOffer | null>(null);
  const [filterResource, setFilterResource] = useState('all');
  const [sortBy, setSortBy] = useState('price');

  // Mock data - replace with actual API calls
  useEffect(() => {
    fetchMarketData();
    const interval = setInterval(fetchMarketData, 10000); // Refresh every 10 seconds
    return () => clearInterval(interval);
  }, []);

  const fetchMarketData = async () => {
    try {
      // Mock data
      const mockOffers: MarketOffer[] = [
        {
          id: '1',
          provider_id: 'provider-1',
          provider_name: 'CloudNode Alpha',
          resource_type: 'GPU',
          cpu_cores: 32,
          memory_gb: 128,
          gpu_count: 4,
          gpu_model: 'NVIDIA A100',
          storage_gb: 1000,
          network_bandwidth_gbps: 10,
          price_per_hour: 4.5,
          availability: '99.9%',
          location: 'US-East',
          reputation_score: 4.8,
          sla_uptime: 99.9,
          created_at: new Date().toISOString(),
          expires_at: new Date(Date.now() + 86400000).toISOString(),
          status: 'active',
        },
        {
          id: '2',
          provider_id: 'provider-2',
          provider_name: 'DataCenter Pro',
          resource_type: 'CPU',
          cpu_cores: 64,
          memory_gb: 256,
          gpu_count: 0,
          gpu_model: '',
          storage_gb: 2000,
          network_bandwidth_gbps: 25,
          price_per_hour: 2.8,
          availability: '98.5%',
          location: 'EU-West',
          reputation_score: 4.6,
          sla_uptime: 99.5,
          created_at: new Date().toISOString(),
          expires_at: new Date(Date.now() + 86400000).toISOString(),
          status: 'active',
        },
      ];

      const mockBids: MarketBid[] = [
        {
          id: '1',
          consumer_id: 'consumer-1',
          consumer_name: 'AI Research Lab',
          resource_requirements: {
            cpu_cores: 16,
            memory_gb: 64,
            gpu_count: 2,
            gpu_model: 'NVIDIA A100',
            storage_gb: 500,
          },
          max_price_per_hour: 5.0,
          duration_hours: 24,
          deadline: new Date(Date.now() + 7200000).toISOString(),
          created_at: new Date().toISOString(),
          status: 'pending',
        },
      ];

      const mockStats: MarketStats = {
        active_offers: 25,
        active_bids: 18,
        avg_cpu_price: 0.8,
        avg_gpu_price: 3.2,
        total_capacity: {
          cpu_cores: 1024,
          memory_gb: 4096,
          gpu_count: 64,
        },
        price_trends: Array.from({ length: 24 }, (_, i) => ({
          time: `${i}:00`,
          cpu_price: 0.8 + Math.random() * 0.2,
          gpu_price: 3.2 + Math.random() * 0.5,
        })),
      };

      setOffers(mockOffers);
      setBids(mockBids);
      setStats(mockStats);
      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch market data:', error);
      setLoading(false);
    }
  };

  const handleCreateOffer = async (formData: any) => {
    try {
      // API call to create offer
      console.log('Creating offer:', formData);
      setCreateOfferOpen(false);
      fetchMarketData();
    } catch (error) {
      console.error('Failed to create offer:', error);
    }
  };

  const handleCreateBid = async (formData: any) => {
    try {
      // API call to create bid
      console.log('Creating bid:', formData);
      setCreateBidOpen(false);
      fetchMarketData();
    } catch (error) {
      console.error('Failed to create bid:', error);
    }
  };

  const handleAcceptOffer = async (offerId: string) => {
    try {
      // API call to accept offer
      console.log('Accepting offer:', offerId);
      fetchMarketData();
    } catch (error) {
      console.error('Failed to accept offer:', error);
    }
  };

  const offerColumns: GridColDef[] = [
    {
      field: 'provider_name',
      headerName: 'Provider',
      width: 180,
      renderCell: (params) => (
        <Box display="flex" alignItems="center">
          <Typography variant="body2">{params.value}</Typography>
          <Chip
            size="small"
            label={`${params.row.reputation_score}★`}
            color="primary"
            sx={{ ml: 1 }}
          />
        </Box>
      ),
    },
    {
      field: 'resource_type',
      headerName: 'Type',
      width: 100,
      renderCell: (params) => (
        <Chip
          label={params.value}
          color={params.value === 'GPU' ? 'success' : 'default'}
          size="small"
        />
      ),
    },
    {
      field: 'specs',
      headerName: 'Specifications',
      width: 300,
      renderCell: (params) => (
        <Box>
          <Typography variant="caption">
            CPU: {params.row.cpu_cores} cores | RAM: {params.row.memory_gb}GB
          </Typography>
          {params.row.gpu_count > 0 && (
            <Typography variant="caption" display="block">
              GPU: {params.row.gpu_count}x {params.row.gpu_model}
            </Typography>
          )}
        </Box>
      ),
    },
    {
      field: 'price_per_hour',
      headerName: 'Price/Hour',
      width: 120,
      renderCell: (params) => (
        <Typography variant="body2" color="primary" fontWeight="bold">
          ${params.value.toFixed(2)}
        </Typography>
      ),
    },
    {
      field: 'location',
      headerName: 'Location',
      width: 100,
    },
    {
      field: 'availability',
      headerName: 'Availability',
      width: 100,
      renderCell: (params) => (
        <Typography variant="body2" color="success.main">
          {params.value}
        </Typography>
      ),
    },
    {
      field: 'actions',
      headerName: 'Actions',
      width: 150,
      renderCell: (params) => (
        <Box>
          <Button
            size="small"
            variant="contained"
            onClick={() => handleAcceptOffer(params.row.id)}
            sx={{ mr: 1 }}
          >
            Accept
          </Button>
          <IconButton
            size="small"
            onClick={() => setSelectedOffer(params.row)}
          >
            <Info />
          </IconButton>
        </Box>
      ),
    },
  ];

  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" fontWeight="bold">
          Compute Marketplace
        </Typography>
        <Box>
          <Button
            variant="outlined"
            startIcon={<Refresh />}
            onClick={fetchMarketData}
            sx={{ mr: 2 }}
          >
            Refresh
          </Button>
          <Button
            variant="contained"
            startIcon={<LocalOffer />}
            onClick={() => setCreateOfferOpen(true)}
            sx={{ mr: 1 }}
          >
            Create Offer
          </Button>
          <Button
            variant="contained"
            startIcon={<ShoppingCart />}
            onClick={() => setCreateBidOpen(true)}
          >
            Create Bid
          </Button>
        </Box>
      </Box>

      {/* Market Statistics */}
      <Grid container spacing={3} mb={3}>
        <Grid item xs={12} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Active Offers
                  </Typography>
                  <Typography variant="h4">
                    {stats?.active_offers || 0}
                  </Typography>
                </Box>
                <LocalOffer color="primary" fontSize="large" />
              </Box>
              <Box display="flex" alignItems="center" mt={1}>
                <TrendingUp color="success" fontSize="small" />
                <Typography variant="caption" color="success.main" ml={0.5}>
                  +12% from last hour
                </Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Active Bids
                  </Typography>
                  <Typography variant="h4">
                    {stats?.active_bids || 0}
                  </Typography>
                </Box>
                <ShoppingCart color="secondary" fontSize="large" />
              </Box>
              <Box display="flex" alignItems="center" mt={1}>
                <TrendingDown color="error" fontSize="small" />
                <Typography variant="caption" color="error.main" ml={0.5}>
                  -5% from last hour
                </Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Avg CPU Price
                  </Typography>
                  <Typography variant="h4">
                    ${stats?.avg_cpu_price.toFixed(2) || '0.00'}
                  </Typography>
                </Box>
                <AttachMoney color="success" fontSize="large" />
              </Box>
              <Typography variant="caption" color="textSecondary">
                Per core per hour
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Avg GPU Price
                  </Typography>
                  <Typography variant="h4">
                    ${stats?.avg_gpu_price.toFixed(2) || '0.00'}
                  </Typography>
                </Box>
                <AttachMoney color="warning" fontSize="large" />
              </Box>
              <Typography variant="caption" color="textSecondary">
                Per GPU per hour
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Price Trends Chart */}
      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" gutterBottom>
          24-Hour Price Trends
        </Typography>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={stats?.price_trends || []}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="time" />
            <YAxis />
            <RechartsTooltip />
            <Legend />
            <Line
              type="monotone"
              dataKey="cpu_price"
              stroke="#8884d8"
              name="CPU Price ($/core/hr)"
            />
            <Line
              type="monotone"
              dataKey="gpu_price"
              stroke="#82ca9d"
              name="GPU Price ($/GPU/hr)"
            />
          </LineChart>
        </ResponsiveContainer>
      </Paper>

      {/* Market Tabs */}
      <Paper sx={{ width: '100%' }}>
        <Tabs
          value={tabValue}
          onChange={(e, newValue) => setTabValue(newValue)}
          aria-label="marketplace tabs"
        >
          <Tab label="Available Offers" />
          <Tab label="Active Bids" />
          <Tab label="My Listings" />
          <Tab label="Transaction History" />
        </Tabs>

        <TabPanel value={tabValue} index={0}>
          {/* Filter Controls */}
          <Box display="flex" gap={2} mb={2}>
            <FormControl size="small" sx={{ minWidth: 120 }}>
              <InputLabel>Resource Type</InputLabel>
              <Select
                value={filterResource}
                onChange={(e) => setFilterResource(e.target.value)}
                label="Resource Type"
              >
                <MenuItem value="all">All</MenuItem>
                <MenuItem value="cpu">CPU Only</MenuItem>
                <MenuItem value="gpu">GPU</MenuItem>
              </Select>
            </FormControl>
            <FormControl size="small" sx={{ minWidth: 120 }}>
              <InputLabel>Sort By</InputLabel>
              <Select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value)}
                label="Sort By"
              >
                <MenuItem value="price">Price</MenuItem>
                <MenuItem value="reputation">Reputation</MenuItem>
                <MenuItem value="availability">Availability</MenuItem>
              </Select>
            </FormControl>
          </Box>

          {/* Offers Grid */}
          <DataGrid
            rows={offers}
            columns={offerColumns}
            pageSize={10}
            rowsPerPageOptions={[10, 25, 50]}
            disableSelectionOnClick
            autoHeight
            loading={loading}
          />
        </TabPanel>

        <TabPanel value={tabValue} index={1}>
          {/* Active Bids */}
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Consumer</TableCell>
                  <TableCell>Requirements</TableCell>
                  <TableCell>Max Price/Hour</TableCell>
                  <TableCell>Duration</TableCell>
                  <TableCell>Deadline</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {bids.map((bid) => (
                  <TableRow key={bid.id}>
                    <TableCell>{bid.consumer_name}</TableCell>
                    <TableCell>
                      <Typography variant="caption">
                        CPU: {bid.resource_requirements.cpu_cores} | 
                        RAM: {bid.resource_requirements.memory_gb}GB
                        {bid.resource_requirements.gpu_count > 0 && (
                          <> | GPU: {bid.resource_requirements.gpu_count}x {bid.resource_requirements.gpu_model}</>
                        )}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Typography color="primary" fontWeight="bold">
                        ${bid.max_price_per_hour.toFixed(2)}
                      </Typography>
                    </TableCell>
                    <TableCell>{bid.duration_hours}h</TableCell>
                    <TableCell>
                      <Chip
                        icon={<Timer />}
                        label={new Date(bid.deadline).toLocaleString()}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={bid.status}
                        color={bid.status === 'matched' ? 'success' : 'default'}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>
                      <Button size="small" variant="outlined">
                        Make Offer
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </TabPanel>

        <TabPanel value={tabValue} index={2}>
          {/* My Listings */}
          <Alert severity="info">
            Your active offers and bids will appear here
          </Alert>
        </TabPanel>

        <TabPanel value={tabValue} index={3}>
          {/* Transaction History */}
          <Alert severity="info">
            Your completed transactions will appear here
          </Alert>
        </TabPanel>
      </Paper>

      {/* Create Offer Dialog */}
      <Dialog
        open={createOfferOpen}
        onClose={() => setCreateOfferOpen(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>Create Resource Offer</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="CPU Cores"
                type="number"
                defaultValue={8}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Memory (GB)"
                type="number"
                defaultValue={32}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="GPU Count"
                type="number"
                defaultValue={0}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="GPU Model"
                placeholder="e.g., NVIDIA A100"
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Storage (GB)"
                type="number"
                defaultValue={100}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Network Bandwidth (Gbps)"
                type="number"
                defaultValue={1}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Price per Hour ($)"
                type="number"
                step="0.01"
                defaultValue={1.0}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <FormControl fullWidth>
                <InputLabel>Location</InputLabel>
                <Select defaultValue="us-east">
                  <MenuItem value="us-east">US-East</MenuItem>
                  <MenuItem value="us-west">US-West</MenuItem>
                  <MenuItem value="eu-west">EU-West</MenuItem>
                  <MenuItem value="asia-pacific">Asia-Pacific</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Duration (hours)"
                type="number"
                defaultValue={24}
                helperText="How long will this offer be available?"
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateOfferOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={() => handleCreateOffer({})}>
            Create Offer
          </Button>
        </DialogActions>
      </Dialog>

      {/* Create Bid Dialog */}
      <Dialog
        open={createBidOpen}
        onClose={() => setCreateBidOpen(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>Create Resource Bid</DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="CPU Cores Needed"
                type="number"
                defaultValue={4}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Memory Needed (GB)"
                type="number"
                defaultValue={16}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="GPU Count Needed"
                type="number"
                defaultValue={0}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Preferred GPU Model"
                placeholder="e.g., NVIDIA A100"
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Storage Needed (GB)"
                type="number"
                defaultValue={50}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Max Price per Hour ($)"
                type="number"
                step="0.01"
                defaultValue={2.0}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Duration (hours)"
                type="number"
                defaultValue={8}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Deadline"
                type="datetime-local"
                InputLabelProps={{ shrink: true }}
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateBidOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={() => handleCreateBid({})}>
            Create Bid
          </Button>
        </DialogActions>
      </Dialog>

      {/* Offer Details Dialog */}
      <Dialog
        open={!!selectedOffer}
        onClose={() => setSelectedOffer(null)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>Offer Details</DialogTitle>
        <DialogContent>
          {selectedOffer && (
            <Box>
              <Typography variant="h6" gutterBottom>
                {selectedOffer.provider_name}
              </Typography>
              <Grid container spacing={2}>
                <Grid item xs={6}>
                  <Typography variant="body2" color="textSecondary">
                    Resource Type
                  </Typography>
                  <Typography variant="body1">
                    {selectedOffer.resource_type}
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="textSecondary">
                    Location
                  </Typography>
                  <Typography variant="body1">
                    {selectedOffer.location}
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="textSecondary">
                    CPU Cores
                  </Typography>
                  <Typography variant="body1">
                    {selectedOffer.cpu_cores}
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="textSecondary">
                    Memory
                  </Typography>
                  <Typography variant="body1">
                    {selectedOffer.memory_gb} GB
                  </Typography>
                </Grid>
                {selectedOffer.gpu_count > 0 && (
                  <>
                    <Grid item xs={6}>
                      <Typography variant="body2" color="textSecondary">
                        GPU Count
                      </Typography>
                      <Typography variant="body1">
                        {selectedOffer.gpu_count}
                      </Typography>
                    </Grid>
                    <Grid item xs={6}>
                      <Typography variant="body2" color="textSecondary">
                        GPU Model
                      </Typography>
                      <Typography variant="body1">
                        {selectedOffer.gpu_model}
                      </Typography>
                    </Grid>
                  </>
                )}
                <Grid item xs={6}>
                  <Typography variant="body2" color="textSecondary">
                    Storage
                  </Typography>
                  <Typography variant="body1">
                    {selectedOffer.storage_gb} GB
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="textSecondary">
                    Network
                  </Typography>
                  <Typography variant="body1">
                    {selectedOffer.network_bandwidth_gbps} Gbps
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="textSecondary">
                    SLA Uptime
                  </Typography>
                  <Typography variant="body1">
                    {selectedOffer.sla_uptime}%
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="textSecondary">
                    Reputation Score
                  </Typography>
                  <Typography variant="body1">
                    {selectedOffer.reputation_score}/5.0
                  </Typography>
                </Grid>
                <Grid item xs={12}>
                  <Typography variant="body2" color="textSecondary">
                    Price per Hour
                  </Typography>
                  <Typography variant="h5" color="primary">
                    ${selectedOffer.price_per_hour.toFixed(2)}
                  </Typography>
                </Grid>
              </Grid>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSelectedOffer(null)}>Close</Button>
          <Button
            variant="contained"
            onClick={() => {
              handleAcceptOffer(selectedOffer!.id);
              setSelectedOffer(null);
            }}
          >
            Accept Offer
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
} 