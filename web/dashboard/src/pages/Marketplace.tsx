import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Grid,
  Card,
  CardContent,
  Button,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from '@mui/material';
import {
  ShoppingCart,
  LocalOffer,
  TrendingUp,
  AttachMoney,
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';

interface MarketOffer {
  id: string;
  provider_name: string;
  resource_type: string;
  cpu_cores: number;
  memory_gb: number;
  gpu_count: number;
  price_per_hour: number;
  location: string;
  status: string;
}

export default function Marketplace() {
  const { user } = useAuth();
  const [offers, setOffers] = useState<MarketOffer[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchMarketData();
    const interval = setInterval(fetchMarketData, 30000);
    return () => clearInterval(interval);
  }, []);

  const fetchMarketData = async () => {
    try {
      // Mock data - replace with API call
      const mockOffers: MarketOffer[] = [
        {
          id: '1',
          provider_name: 'CloudNode Alpha',
          resource_type: 'GPU',
          cpu_cores: 32,
          memory_gb: 128,
          gpu_count: 4,
          price_per_hour: 4.5,
          location: 'US-East',
          status: 'active',
        },
        {
          id: '2',
          provider_name: 'DataCenter Pro',
          resource_type: 'CPU',
          cpu_cores: 64,
          memory_gb: 256,
          gpu_count: 0,
          price_per_hour: 2.8,
          location: 'EU-West',
          status: 'active',
        },
      ];

      setOffers(mockOffers);
      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch market data:', error);
      setLoading(false);
    }
  };

  return (
    <Box>
      <Typography variant="h4" fontWeight="bold" mb={3}>
        Compute Marketplace
      </Typography>

      <Grid container spacing={3} mb={3}>
        <Grid item xs={12} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Active Offers
                  </Typography>
                  <Typography variant="h4">25</Typography>
                </Box>
                <LocalOffer color="primary" fontSize="large" />
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
                  <Typography variant="h4">18</Typography>
                </Box>
                <ShoppingCart color="secondary" fontSize="large" />
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
                  <Typography variant="h4">$0.80</Typography>
                </Box>
                <AttachMoney color="success" fontSize="large" />
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
                    Avg GPU Price
                  </Typography>
                  <Typography variant="h4">$3.20</Typography>
                </Box>
                <AttachMoney color="warning" fontSize="large" />
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Paper sx={{ p: 3 }}>
        <Typography variant="h6" gutterBottom>
          Available Offers
        </Typography>
        
        <TableContainer>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Provider</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>Specifications</TableCell>
                <TableCell>Price/Hour</TableCell>
                <TableCell>Location</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {offers.map((offer) => (
                <TableRow key={offer.id}>
                  <TableCell>{offer.provider_name}</TableCell>
                  <TableCell>
                    <Chip
                      label={offer.resource_type}
                      color={offer.resource_type === 'GPU' ? 'success' : 'default'}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>
                    CPU: {offer.cpu_cores} cores | RAM: {offer.memory_gb}GB
                    {offer.gpu_count > 0 && ` | GPU: ${offer.gpu_count}`}
                  </TableCell>
                  <TableCell>
                    <Typography color="primary" fontWeight="bold">
                      ${offer.price_per_hour.toFixed(2)}
                    </Typography>
                  </TableCell>
                  <TableCell>{offer.location}</TableCell>
                  <TableCell>
                    <Button size="small" variant="contained">
                      Accept
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </Paper>
    </Box>
  );
}
