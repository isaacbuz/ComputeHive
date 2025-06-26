import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Grid,
  TextField,
  Button,
  Switch,
  FormControlLabel,
  Divider,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
  ListItemSecondaryAction,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Alert,
  Chip,
  Avatar,
  Tabs,
  Tab,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Slider,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Badge,
  Tooltip,
  AlertTitle,
  LinearProgress
} from '@mui/material';
import {
  Person,
  Security,
  Notifications,
  Payment,
  Storage,
  Api,
  Delete,
  Edit,
  Save,
  Cancel,
  Visibility,
  VisibilityOff,
  ExpandMore,
  Warning,
  CheckCircle,
  Error,
  Info,
  Key,
  Email,
  Phone,
  LocationOn,
  Business,
  Language,
  DarkMode,
  LightMode,
  Refresh,
  Download,
  Upload,
  CloudUpload,
  CloudDownload,
  VpnKey,
  TwoWheeler,
  VerifiedUser,
  Block,
  History,
  Settings as SettingsIcon
} from '@mui/icons-material';

interface UserProfile {
  id: string;
  email: string;
  username: string;
  first_name: string;
  last_name: string;
  avatar?: string;
  phone?: string;
  company?: string;
  location?: string;
  timezone: string;
  language: string;
  created_at: string;
  last_login: string;
  verified: boolean;
  two_factor_enabled: boolean;
  api_keys: ApiKey[];
  preferences: UserPreferences;
}

interface ApiKey {
  id: string;
  name: string;
  key: string;
  created_at: string;
  last_used?: string;
  permissions: string[];
  active: boolean;
}

interface UserPreferences {
  theme: 'light' | 'dark' | 'auto';
  notifications: {
    email: boolean;
    push: boolean;
    sms: boolean;
    job_updates: boolean;
    billing_alerts: boolean;
    security_alerts: boolean;
    marketing: boolean;
  };
  privacy: {
    profile_public: boolean;
    show_usage_stats: boolean;
    allow_analytics: boolean;
  };
  compute: {
    default_region: string;
    auto_scale: boolean;
    cost_optimization: boolean;
    max_concurrent_jobs: number;
  };
}

const Settings: React.FC = () => {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [tabValue, setTabValue] = useState(0);
  const [showApiKeyDialog, setShowApiKeyDialog] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [showPasswordDialog, setShowPasswordDialog] = useState(false);
  const [showTwoFactorDialog, setShowTwoFactorDialog] = useState(false);
  const [editingProfile, setEditingProfile] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  // Mock data
  useEffect(() => {
    const mockProfile: UserProfile = {
      id: 'user-123',
      email: 'john.doe@example.com',
      username: 'johndoe',
      first_name: 'John',
      last_name: 'Doe',
      avatar: 'https://via.placeholder.com/150',
      phone: '+1 (555) 123-4567',
      company: 'TechCorp Inc.',
      location: 'San Francisco, CA',
      timezone: 'America/Los_Angeles',
      language: 'en',
      created_at: '2024-01-15T10:30:00Z',
      last_login: new Date().toISOString(),
      verified: true,
      two_factor_enabled: false,
      api_keys: [
        {
          id: 'key-1',
          name: 'Production API Key',
          key: 'ch_live_1234567890abcdef',
          created_at: '2024-01-20T14:30:00Z',
          last_used: '2024-01-25T09:15:00Z',
          permissions: ['read', 'write', 'jobs', 'payments'],
          active: true
        },
        {
          id: 'key-2',
          name: 'Development API Key',
          key: 'ch_test_abcdef1234567890',
          created_at: '2024-01-22T16:45:00Z',
          permissions: ['read', 'jobs'],
          active: true
        }
      ],
      preferences: {
        theme: 'dark',
        notifications: {
          email: true,
          push: true,
          sms: false,
          job_updates: true,
          billing_alerts: true,
          security_alerts: true,
          marketing: false
        },
        privacy: {
          profile_public: false,
          show_usage_stats: true,
          allow_analytics: true
        },
        compute: {
          default_region: 'us-east-1',
          auto_scale: true,
          cost_optimization: true,
          max_concurrent_jobs: 10
        }
      }
    };

    setProfile(mockProfile);
    setLoading(false);
  }, []);

  const handleSaveProfile = async () => {
    setSaving(true);
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      setSuccess('Profile updated successfully!');
      setEditingProfile(false);
    } catch (err) {
      setError('Failed to update profile');
    } finally {
      setSaving(false);
    }
  };

  const handleGenerateApiKey = async () => {
    setSaving(true);
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      setSuccess('API key generated successfully!');
      setShowApiKeyDialog(false);
    } catch (err) {
      setError('Failed to generate API key');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteAccount = async () => {
    setSaving(true);
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 2000));
      setSuccess('Account deletion initiated. You will receive a confirmation email.');
      setShowDeleteDialog(false);
    } catch (err) {
      setError('Failed to delete account');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <Box sx={{ p: 3 }}>
        <LinearProgress />
      </Box>
    );
  }

  if (!profile) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="error">Failed to load profile</Alert>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1">
          Settings
        </Typography>
        <Box>
          <Button
            variant="outlined"
            startIcon={<Refresh />}
            onClick={() => window.location.reload()}
            sx={{ mr: 1 }}
          >
            Refresh
          </Button>
        </Box>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {success && (
        <Alert severity="success" sx={{ mb: 3 }} onClose={() => setSuccess(null)}>
          {success}
        </Alert>
      )}

      <Grid container spacing={3}>
        {/* Profile Overview Card */}
        <Grid item xs={12} md={4}>
          <Card>
            <CardContent sx={{ textAlign: 'center' }}>
              <Badge
                overlap="circular"
                anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
                badgeContent={
                  <Tooltip title="Verified Account">
                    <CheckCircle color="primary" fontSize="small" />
                  </Tooltip>
                }
              >
                <Avatar
                  src={profile.avatar}
                  sx={{ width: 100, height: 100, mx: 'auto', mb: 2 }}
                >
                  {profile.first_name[0]}{profile.last_name[0]}
                </Avatar>
              </Badge>
              
              <Typography variant="h5" gutterBottom>
                {profile.first_name} {profile.last_name}
              </Typography>
              
              <Typography color="textSecondary" gutterBottom>
                @{profile.username}
              </Typography>
              
              <Typography variant="body2" color="textSecondary" gutterBottom>
                {profile.email}
              </Typography>
              
              {profile.company && (
                <Typography variant="body2" color="textSecondary" gutterBottom>
                  {profile.company}
                </Typography>
              )}
              
              <Box sx={{ mt: 2 }}>
                <Chip
                  icon={profile.verified ? <VerifiedUser /> : <Block />}
                  label={profile.verified ? 'Verified' : 'Unverified'}
                  color={profile.verified ? 'success' : 'error'}
                  size="small"
                  sx={{ mr: 1 }}
                />
                <Chip
                  icon={profile.two_factor_enabled ? <TwoWheeler /> : <VpnKey />}
                  label={profile.two_factor_enabled ? '2FA Enabled' : '2FA Disabled'}
                  color={profile.two_factor_enabled ? 'success' : 'default'}
                  size="small"
                />
              </Box>
              
              <Divider sx={{ my: 2 }} />
              
              <Typography variant="body2" color="textSecondary">
                Member since {new Date(profile.created_at).toLocaleDateString()}
              </Typography>
              
              <Typography variant="body2" color="textSecondary">
                Last login: {new Date(profile.last_login).toLocaleString()}
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        {/* Settings Tabs */}
        <Grid item xs={12} md={8}>
          <Card>
            <CardContent>
              <Tabs value={tabValue} onChange={(e, newValue) => setTabValue(newValue)}>
                <Tab label="Profile" icon={<Person />} />
                <Tab label="Security" icon={<Security />} />
                <Tab label="Notifications" icon={<Notifications />} />
                <Tab label="Preferences" icon={<SettingsIcon />} />
                <Tab label="API Keys" icon={<Api />} />
                <Tab label="Billing" icon={<Payment />} />
              </Tabs>
              
              <Box sx={{ mt: 3 }}>
                {/* Profile Tab */}
                {tabValue === 0 && (
                  <Box>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                      <Typography variant="h6">Profile Information</Typography>
                      <Button
                        variant={editingProfile ? 'outlined' : 'contained'}
                        startIcon={editingProfile ? <Cancel /> : <Edit />}
                        onClick={() => setEditingProfile(!editingProfile)}
                      >
                        {editingProfile ? 'Cancel' : 'Edit Profile'}
                      </Button>
                    </Box>
                    
                    <Grid container spacing={2}>
                      <Grid item xs={12} sm={6}>
                        <TextField
                          fullWidth
                          label="First Name"
                          value={profile.first_name}
                          disabled={!editingProfile}
                          variant={editingProfile ? 'outlined' : 'standard'}
                        />
                      </Grid>
                      <Grid item xs={12} sm={6}>
                        <TextField
                          fullWidth
                          label="Last Name"
                          value={profile.last_name}
                          disabled={!editingProfile}
                          variant={editingProfile ? 'outlined' : 'standard'}
                        />
                      </Grid>
                      <Grid item xs={12}>
                        <TextField
                          fullWidth
                          label="Email"
                          value={profile.email}
                          disabled
                          variant="standard"
                          InputProps={{
                            endAdornment: profile.verified ? (
                              <CheckCircle color="success" />
                            ) : (
                              <Warning color="warning" />
                            )
                          }}
                        />
                      </Grid>
                      <Grid item xs={12}>
                        <TextField
                          fullWidth
                          label="Username"
                          value={profile.username}
                          disabled
                          variant="standard"
                        />
                      </Grid>
                      <Grid item xs={12}>
                        <TextField
                          fullWidth
                          label="Phone"
                          value={profile.phone || ''}
                          disabled={!editingProfile}
                          variant={editingProfile ? 'outlined' : 'standard'}
                        />
                      </Grid>
                      <Grid item xs={12}>
                        <TextField
                          fullWidth
                          label="Company"
                          value={profile.company || ''}
                          disabled={!editingProfile}
                          variant={editingProfile ? 'outlined' : 'standard'}
                        />
                      </Grid>
                      <Grid item xs={12}>
                        <TextField
                          fullWidth
                          label="Location"
                          value={profile.location || ''}
                          disabled={!editingProfile}
                          variant={editingProfile ? 'outlined' : 'standard'}
                        />
                      </Grid>
                    </Grid>
                    
                    {editingProfile && (
                      <Box sx={{ mt: 3, display: 'flex', gap: 2 }}>
                        <Button
                          variant="contained"
                          startIcon={<Save />}
                          onClick={handleSaveProfile}
                          disabled={saving}
                        >
                          {saving ? 'Saving...' : 'Save Changes'}
                        </Button>
                        <Button
                          variant="outlined"
                          onClick={() => setEditingProfile(false)}
                        >
                          Cancel
                        </Button>
                      </Box>
                    )}
                  </Box>
                )}

                {/* Security Tab */}
                {tabValue === 1 && (
                  <Box>
                    <Typography variant="h6" gutterBottom>Security Settings</Typography>
                    
                    <List>
                      <ListItem>
                        <ListItemIcon>
                          <VpnKey />
                        </ListItemIcon>
                        <ListItemText
                          primary="Change Password"
                          secondary="Update your account password"
                        />
                        <ListItemSecondaryAction>
                          <Button
                            variant="outlined"
                            size="small"
                            onClick={() => setShowPasswordDialog(true)}
                          >
                            Change
                          </Button>
                        </ListItemSecondaryAction>
                      </ListItem>
                      
                      <ListItem>
                        <ListItemIcon>
                          <TwoWheeler />
                        </ListItemIcon>
                        <ListItemText
                          primary="Two-Factor Authentication"
                          secondary={profile.two_factor_enabled ? 'Enabled' : 'Add an extra layer of security'}
                        />
                        <ListItemSecondaryAction>
                          <Button
                            variant="outlined"
                            size="small"
                            onClick={() => setShowTwoFactorDialog(true)}
                          >
                            {profile.two_factor_enabled ? 'Manage' : 'Enable'}
                          </Button>
                        </ListItemSecondaryAction>
                      </ListItem>
                      
                      <ListItem>
                        <ListItemIcon>
                          <History />
                        </ListItemIcon>
                        <ListItemText
                          primary="Login History"
                          secondary="View recent login activity"
                        />
                        <ListItemSecondaryAction>
                          <Button variant="outlined" size="small">
                            View
                          </Button>
                        </ListItemSecondaryAction>
                      </ListItem>
                      
                      <ListItem>
                        <ListItemIcon>
                          <Delete color="error" />
                        </ListItemIcon>
                        <ListItemText
                          primary="Delete Account"
                          secondary="Permanently delete your account and all data"
                        />
                        <ListItemSecondaryAction>
                          <Button
                            variant="outlined"
                            color="error"
                            size="small"
                            onClick={() => setShowDeleteDialog(true)}
                          >
                            Delete
                          </Button>
                        </ListItemSecondaryAction>
                      </ListItem>
                    </List>
                  </Box>
                )}

                {/* Notifications Tab */}
                {tabValue === 2 && (
                  <Box>
                    <Typography variant="h6" gutterBottom>Notification Preferences</Typography>
                    
                    <Grid container spacing={2}>
                      <Grid item xs={12}>
                        <FormControlLabel
                          control={
                            <Switch
                              checked={profile.preferences.notifications.email}
                              onChange={(e) => {
                                // Handle change
                              }}
                            />
                          }
                          label="Email Notifications"
                        />
                      </Grid>
                      <Grid item xs={12}>
                        <FormControlLabel
                          control={
                            <Switch
                              checked={profile.preferences.notifications.push}
                              onChange={(e) => {
                                // Handle change
                              }}
                            />
                          }
                          label="Push Notifications"
                        />
                      </Grid>
                      <Grid item xs={12}>
                        <FormControlLabel
                          control={
                            <Switch
                              checked={profile.preferences.notifications.sms}
                              onChange={(e) => {
                                // Handle change
                              }}
                            />
                          }
                          label="SMS Notifications"
                        />
                      </Grid>
                      <Grid item xs={12}>
                        <FormControlLabel
                          control={
                            <Switch
                              checked={profile.preferences.notifications.job_updates}
                              onChange={(e) => {
                                // Handle change
                              }}
                            />
                          }
                          label="Job Status Updates"
                        />
                      </Grid>
                      <Grid item xs={12}>
                        <FormControlLabel
                          control={
                            <Switch
                              checked={profile.preferences.notifications.billing_alerts}
                              onChange={(e) => {
                                // Handle change
                              }}
                            />
                          }
                          label="Billing Alerts"
                        />
                      </Grid>
                      <Grid item xs={12}>
                        <FormControlLabel
                          control={
                            <Switch
                              checked={profile.preferences.notifications.security_alerts}
                              onChange={(e) => {
                                // Handle change
                              }}
                            />
                          }
                          label="Security Alerts"
                        />
                      </Grid>
                      <Grid item xs={12}>
                        <FormControlLabel
                          control={
                            <Switch
                              checked={profile.preferences.notifications.marketing}
                              onChange={(e) => {
                                // Handle change
                              }}
                            />
                          }
                          label="Marketing Communications"
                        />
                      </Grid>
                    </Grid>
                  </Box>
                )}

                {/* Preferences Tab */}
                {tabValue === 3 && (
                  <Box>
                    <Typography variant="h6" gutterBottom>Application Preferences</Typography>
                    
                    <Grid container spacing={3}>
                      <Grid item xs={12} md={6}>
                        <FormControl fullWidth>
                          <InputLabel>Theme</InputLabel>
                          <Select
                            value={profile.preferences.theme}
                            label="Theme"
                            onChange={(e) => {
                              // Handle change
                            }}
                          >
                            <MenuItem value="light">Light</MenuItem>
                            <MenuItem value="dark">Dark</MenuItem>
                            <MenuItem value="auto">Auto</MenuItem>
                          </Select>
                        </FormControl>
                      </Grid>
                      
                      <Grid item xs={12} md={6}>
                        <FormControl fullWidth>
                          <InputLabel>Language</InputLabel>
                          <Select
                            value={profile.preferences.language}
                            label="Language"
                            onChange={(e) => {
                              // Handle change
                            }}
                          >
                            <MenuItem value="en">English</MenuItem>
                            <MenuItem value="es">Spanish</MenuItem>
                            <MenuItem value="fr">French</MenuItem>
                            <MenuItem value="de">German</MenuItem>
                          </Select>
                        </FormControl>
                      </Grid>
                      
                      <Grid item xs={12} md={6}>
                        <FormControl fullWidth>
                          <InputLabel>Default Region</InputLabel>
                          <Select
                            value={profile.preferences.compute.default_region}
                            label="Default Region"
                            onChange={(e) => {
                              // Handle change
                            }}
                          >
                            <MenuItem value="us-east-1">US East (N. Virginia)</MenuItem>
                            <MenuItem value="us-west-1">US West (Oregon)</MenuItem>
                            <MenuItem value="eu-west-1">Europe (Ireland)</MenuItem>
                            <MenuItem value="ap-southeast-1">Asia Pacific (Singapore)</MenuItem>
                          </Select>
                        </FormControl>
                      </Grid>
                      
                      <Grid item xs={12} md={6}>
                        <Typography gutterBottom>Max Concurrent Jobs</Typography>
                        <Slider
                          value={profile.preferences.compute.max_concurrent_jobs}
                          min={1}
                          max={50}
                          marks
                          valueLabelDisplay="auto"
                          onChange={(e, value) => {
                            // Handle change
                          }}
                        />
                      </Grid>
                      
                      <Grid item xs={12}>
                        <FormControlLabel
                          control={
                            <Switch
                              checked={profile.preferences.compute.auto_scale}
                              onChange={(e) => {
                                // Handle change
                              }}
                            />
                          }
                          label="Auto-scaling"
                        />
                      </Grid>
                      
                      <Grid item xs={12}>
                        <FormControlLabel
                          control={
                            <Switch
                              checked={profile.preferences.compute.cost_optimization}
                              onChange={(e) => {
                                // Handle change
                              }}
                            />
                          }
                          label="Cost Optimization"
                        />
                      </Grid>
                    </Grid>
                  </Box>
                )}

                {/* API Keys Tab */}
                {tabValue === 4 && (
                  <Box>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                      <Typography variant="h6">API Keys</Typography>
                      <Button
                        variant="contained"
                        startIcon={<Add />}
                        onClick={() => setShowApiKeyDialog(true)}
                      >
                        Generate New Key
                      </Button>
                    </Box>
                    
                    <List>
                      {profile.api_keys.map((apiKey) => (
                        <ListItem key={apiKey.id} divider>
                          <ListItemIcon>
                            <VpnKey />
                          </ListItemIcon>
                          <ListItemText
                            primary={apiKey.name}
                            secondary={
                              <Box>
                                <Typography variant="body2" color="textSecondary">
                                  {apiKey.key.substring(0, 8)}...{apiKey.key.substring(apiKey.key.length - 4)}
                                </Typography>
                                <Box sx={{ mt: 1 }}>
                                  {apiKey.permissions.map((permission) => (
                                    <Chip
                                      key={permission}
                                      label={permission}
                                      size="small"
                                      sx={{ mr: 0.5, mb: 0.5 }}
                                    />
                                  ))}
                                </Box>
                                <Typography variant="caption" color="textSecondary">
                                  Created: {new Date(apiKey.created_at).toLocaleDateString()}
                                  {apiKey.last_used && ` • Last used: ${new Date(apiKey.last_used).toLocaleDateString()}`}
                                </Typography>
                              </Box>
                            }
                          />
                          <ListItemSecondaryAction>
                            <IconButton edge="end" aria-label="delete">
                              <Delete />
                            </IconButton>
                          </ListItemSecondaryAction>
                        </ListItem>
                      ))}
                    </List>
                  </Box>
                )}

                {/* Billing Tab */}
                {tabValue === 5 && (
                  <Box>
                    <Typography variant="h6" gutterBottom>Billing & Payment</Typography>
                    
                    <Alert severity="info" sx={{ mb: 2 }}>
                      <AlertTitle>Billing Information</AlertTitle>
                      Manage your payment methods, view invoices, and update billing preferences.
                    </Alert>
                    
                    <List>
                      <ListItem>
                        <ListItemIcon>
                          <Payment />
                        </ListItemIcon>
                        <ListItemText
                          primary="Payment Methods"
                          secondary="Manage your payment methods"
                        />
                        <ListItemSecondaryAction>
                          <Button variant="outlined" size="small">
                            Manage
                          </Button>
                        </ListItemSecondaryAction>
                      </ListItem>
                      
                      <ListItem>
                        <ListItemIcon>
                          <Download />
                        </ListItemIcon>
                        <ListItemText
                          primary="Billing History"
                          secondary="View and download invoices"
                        />
                        <ListItemSecondaryAction>
                          <Button variant="outlined" size="small">
                            View
                          </Button>
                        </ListItemSecondaryAction>
                      </ListItem>
                      
                      <ListItem>
                        <ListItemIcon>
                          <Storage />
                        </ListItemIcon>
                        <ListItemText
                          primary="Usage & Billing"
                          secondary="Monitor your usage and costs"
                        />
                        <ListItemSecondaryAction>
                          <Button variant="outlined" size="small">
                            Monitor
                          </Button>
                        </ListItemSecondaryAction>
                      </ListItem>
                    </List>
                  </Box>
                )}
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Generate API Key Dialog */}
      <Dialog open={showApiKeyDialog} onClose={() => setShowApiKeyDialog(false)}>
        <DialogTitle>Generate New API Key</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            label="Key Name"
            placeholder="e.g., Production API Key"
            sx={{ mb: 2, mt: 1 }}
          />
          <Typography variant="subtitle2" gutterBottom>
            Permissions
          </Typography>
          <Grid container spacing={1}>
            {['read', 'write', 'jobs', 'payments', 'admin'].map((permission) => (
              <Grid item xs={6} key={permission}>
                <FormControlLabel
                  control={<Switch defaultChecked={permission === 'read'} />}
                  label={permission.charAt(0).toUpperCase() + permission.slice(1)}
                />
              </Grid>
            ))}
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowApiKeyDialog(false)}>Cancel</Button>
          <Button
            variant="contained"
            onClick={handleGenerateApiKey}
            disabled={saving}
          >
            {saving ? 'Generating...' : 'Generate Key'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Change Password Dialog */}
      <Dialog open={showPasswordDialog} onClose={() => setShowPasswordDialog(false)}>
        <DialogTitle>Change Password</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            type="password"
            label="Current Password"
            sx={{ mb: 2, mt: 1 }}
          />
          <TextField
            fullWidth
            type={showPassword ? 'text' : 'password'}
            label="New Password"
            sx={{ mb: 2 }}
            InputProps={{
              endAdornment: (
                <IconButton
                  onClick={() => setShowPassword(!showPassword)}
                  edge="end"
                >
                  {showPassword ? <VisibilityOff /> : <Visibility />}
                </IconButton>
              )
            }}
          />
          <TextField
            fullWidth
            type="password"
            label="Confirm New Password"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowPasswordDialog(false)}>Cancel</Button>
          <Button variant="contained">Change Password</Button>
        </DialogActions>
      </Dialog>

      {/* Two-Factor Authentication Dialog */}
      <Dialog open={showTwoFactorDialog} onClose={() => setShowTwoFactorDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Two-Factor Authentication</DialogTitle>
        <DialogContent>
          <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
            Add an extra layer of security to your account by enabling two-factor authentication.
          </Typography>
          <Box sx={{ textAlign: 'center', my: 2 }}>
            <img
              src="https://via.placeholder.com/200x200?text=QR+Code"
              alt="QR Code"
              style={{ maxWidth: '200px' }}
            />
          </Box>
          <TextField
            fullWidth
            label="Verification Code"
            placeholder="Enter 6-digit code"
            sx={{ mb: 2 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowTwoFactorDialog(false)}>Cancel</Button>
          <Button variant="contained">Enable 2FA</Button>
        </DialogActions>
      </Dialog>

      {/* Delete Account Dialog */}
      <Dialog open={showDeleteDialog} onClose={() => setShowDeleteDialog(false)}>
        <DialogTitle>Delete Account</DialogTitle>
        <DialogContent>
          <Alert severity="warning" sx={{ mb: 2 }}>
            <AlertTitle>Warning</AlertTitle>
            This action cannot be undone. All your data, jobs, and settings will be permanently deleted.
          </Alert>
          <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
            To confirm deletion, please type "DELETE" in the field below:
          </Typography>
          <TextField
            fullWidth
            label="Type DELETE to confirm"
            placeholder="DELETE"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowDeleteDialog(false)}>Cancel</Button>
          <Button
            variant="contained"
            color="error"
            onClick={handleDeleteAccount}
            disabled={saving}
          >
            {saving ? 'Deleting...' : 'Delete Account'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Settings; 