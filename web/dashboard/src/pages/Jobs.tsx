import React, { useState } from 'react';
import {
  Box,
  Paper,
  Typography,
  Button,
  Chip,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Grid,
  Card,
  CardContent,
  CardActions,
  Tooltip,
  LinearProgress,
  Tab,
  Tabs,
  Alert,
  Slider,
  FormHelperText,
  Autocomplete,
} from '@mui/material';
import {
  Add as AddIcon,
  Refresh as RefreshIcon,
  FilterList as FilterIcon,
  Delete as DeleteIcon,
  Stop as StopIcon,
  PlayArrow as PlayIcon,
  Info as InfoIcon,
  Download as DownloadIcon,
  Terminal as TerminalIcon,
  Schedule as ScheduleIcon,
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Cancel as CancelIcon,
  CloudUpload as CloudUploadIcon,
} from '@mui/icons-material';
import { DataGrid, GridColDef, GridRenderCellParams } from '@mui/x-data-grid';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useSnackbar } from 'notistack';
import axios from 'axios';
import { format } from 'date-fns';

// Job interfaces
interface Job {
  id: string;
  type: string;
  status: string;
  priority: number;
  requirements: {
    cpu_cores: number;
    memory_mb: number;
    gpu_count: number;
    gpu_type?: string;
    storage_mb: number;
  };
  payload: any;
  assigned_agent_id?: string;
  created_at: string;
  scheduled_at?: string;
  started_at?: string;
  completed_at?: string;
  estimated_cost: number;
  actual_cost?: number;
  progress?: number;
  error_message?: string;
}

interface JobFormData {
  type: string;
  priority: number;
  cpu_cores: number;
  memory_gb: number;
  gpu_count: number;
  gpu_type: string;
  storage_gb: number;
  docker_image?: string;
  command?: string;
  script_content?: string;
  script_language?: string;
  timeout: number;
  max_retries: number;
}

const jobStatusColors: Record<string, 'default' | 'primary' | 'secondary' | 'error' | 'info' | 'success' | 'warning'> = {
  pending: 'default',
  scheduled: 'info',
  running: 'primary',
  completed: 'success',
  failed: 'error',
  cancelled: 'warning',
};

const getJobStatusIcon = (status: string) => {
  switch (status) {
    case 'pending':
      return <ScheduleIcon fontSize="small" />;
    case 'scheduled':
      return <ScheduleIcon fontSize="small" color="info" />;
    case 'running':
      return <PlayIcon fontSize="small" color="primary" />;
    case 'completed':
      return <CheckCircleIcon fontSize="small" color="success" />;
    case 'failed':
      return <ErrorIcon fontSize="small" color="error" />;
    case 'cancelled':
      return <CancelIcon fontSize="small" color="warning" />;
    default:
      return <InfoIcon fontSize="small" />;
  }
};

export default function Jobs() {
  const [selectedTab, setSelectedTab] = useState(0);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [selectedJob, setSelectedJob] = useState<Job | null>(null);
  const [formData, setFormData] = useState<JobFormData>({
    type: 'docker',
    priority: 5,
    cpu_cores: 1,
    memory_gb: 1,
    gpu_count: 0,
    gpu_type: '',
    storage_gb: 10,
    timeout: 3600,
    max_retries: 3,
  });
  
  const queryClient = useQueryClient();
  const { enqueueSnackbar } = useSnackbar();

  // Fetch jobs
  const { data: jobs, isLoading, refetch } = useQuery<Job[]>({
    queryKey: ['jobs'],
    queryFn: async () => {
      const response = await axios.get('/api/v1/jobs');
      return response.data;
    },
    refetchInterval: 5000, // Refresh every 5 seconds
  });

  // Create job mutation
  const createJobMutation = useMutation({
    mutationFn: async (data: any) => {
      const response = await axios.post('/api/v1/jobs', data);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['jobs'] });
      enqueueSnackbar('Job created successfully', { variant: 'success' });
      setCreateDialogOpen(false);
      resetForm();
    },
    onError: (error: any) => {
      enqueueSnackbar(error.response?.data?.message || 'Failed to create job', { variant: 'error' });
    },
  });

  // Cancel job mutation
  const cancelJobMutation = useMutation({
    mutationFn: async (jobId: string) => {
      await axios.post(`/api/v1/jobs/${jobId}/cancel`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['jobs'] });
      enqueueSnackbar('Job cancelled', { variant: 'info' });
    },
    onError: (error: any) => {
      enqueueSnackbar(error.response?.data?.message || 'Failed to cancel job', { variant: 'error' });
    },
  });

  const resetForm = () => {
    setFormData({
      type: 'docker',
      priority: 5,
      cpu_cores: 1,
      memory_gb: 1,
      gpu_count: 0,
      gpu_type: '',
      storage_gb: 10,
      timeout: 3600,
      max_retries: 3,
    });
  };

  const handleCreateJob = () => {
    const jobData = {
      type: formData.type,
      priority: formData.priority,
      requirements: {
        cpu_cores: formData.cpu_cores,
        memory_mb: formData.memory_gb * 1024,
        gpu_count: formData.gpu_count,
        gpu_type: formData.gpu_type || undefined,
        storage_mb: formData.storage_gb * 1024,
        network_mbps: 100,
      },
      payload: {},
      timeout: formData.timeout,
      max_retries: formData.max_retries,
    };

    // Add type-specific payload
    if (formData.type === 'docker') {
      jobData.payload = {
        image: formData.docker_image,
        command: formData.command?.split(' '),
      };
    } else if (formData.type === 'script') {
      jobData.payload = {
        script: formData.script_content,
        language: formData.script_language,
      };
    }

    createJobMutation.mutate(jobData);
  };

  const filteredJobs = jobs?.filter(job => {
    if (selectedTab === 0) return true; // All jobs
    if (selectedTab === 1) return ['pending', 'scheduled', 'running'].includes(job.status);
    if (selectedTab === 2) return job.status === 'completed';
    if (selectedTab === 3) return job.status === 'failed';
    return true;
  });

  const columns: GridColDef[] = [
    {
      field: 'id',
      headerName: 'Job ID',
      width: 150,
      renderCell: (params: GridRenderCellParams) => (
        <Tooltip title={params.value}>
          <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
            {params.value.substring(0, 12)}...
          </Typography>
        </Tooltip>
      ),
    },
    {
      field: 'type',
      headerName: 'Type',
      width: 100,
      renderCell: (params: GridRenderCellParams) => (
        <Chip label={params.value} size="small" variant="outlined" />
      ),
    },
    {
      field: 'status',
      headerName: 'Status',
      width: 120,
      renderCell: (params: GridRenderCellParams) => (
        <Box display="flex" alignItems="center" gap={0.5}>
          {getJobStatusIcon(params.value)}
          <Chip
            label={params.value}
            size="small"
            color={jobStatusColors[params.value] || 'default'}
          />
        </Box>
      ),
    },
    {
      field: 'priority',
      headerName: 'Priority',
      width: 80,
      align: 'center',
      renderCell: (params: GridRenderCellParams) => (
        <Chip
          label={params.value}
          size="small"
          color={params.value >= 8 ? 'error' : params.value >= 5 ? 'warning' : 'default'}
        />
      ),
    },
    {
      field: 'requirements',
      headerName: 'Resources',
      width: 200,
      renderCell: (params: GridRenderCellParams) => (
        <Box>
          <Typography variant="caption">
            CPU: {params.value.cpu_cores} • RAM: {params.value.memory_mb / 1024}GB
            {params.value.gpu_count > 0 && ` • GPU: ${params.value.gpu_count}`}
          </Typography>
        </Box>
      ),
    },
    {
      field: 'created_at',
      headerName: 'Created',
      width: 180,
      renderCell: (params: GridRenderCellParams) => (
        <Typography variant="body2">
          {format(new Date(params.value), 'MMM d, HH:mm:ss')}
        </Typography>
      ),
    },
    {
      field: 'estimated_cost',
      headerName: 'Est. Cost',
      width: 100,
      align: 'right',
      renderCell: (params: GridRenderCellParams) => (
        <Typography variant="body2" fontWeight="medium">
          ${params.value.toFixed(2)}
        </Typography>
      ),
    },
    {
      field: 'actions',
      headerName: 'Actions',
      width: 120,
      sortable: false,
      renderCell: (params: GridRenderCellParams) => {
        const job = params.row as Job;
        return (
          <Box>
            <Tooltip title="View details">
              <IconButton size="small" onClick={() => setSelectedJob(job)}>
                <InfoIcon fontSize="small" />
              </IconButton>
            </Tooltip>
            {['pending', 'scheduled', 'running'].includes(job.status) && (
              <Tooltip title="Cancel job">
                <IconButton
                  size="small"
                  color="error"
                  onClick={() => cancelJobMutation.mutate(job.id)}
                >
                  <StopIcon fontSize="small" />
                </IconButton>
              </Tooltip>
            )}
            {job.status === 'completed' && (
              <Tooltip title="Download results">
                <IconButton size="small" color="primary">
                  <DownloadIcon fontSize="small" />
                </IconButton>
              </Tooltip>
            )}
          </Box>
        );
      },
    },
  ];

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Jobs</Typography>
        <Box display="flex" gap={2}>
          <Button
            startIcon={<RefreshIcon />}
            onClick={() => refetch()}
            disabled={isLoading}
          >
            Refresh
          </Button>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setCreateDialogOpen(true)}
          >
            Create Job
          </Button>
        </Box>
      </Box>

      {/* Stats Cards */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Total Jobs
              </Typography>
              <Typography variant="h4">
                {jobs?.length || 0}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Active Jobs
              </Typography>
              <Typography variant="h4" color="primary">
                {jobs?.filter(j => ['pending', 'scheduled', 'running'].includes(j.status)).length || 0}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Success Rate
              </Typography>
              <Typography variant="h4" color="success.main">
                {jobs && jobs.length > 0
                  ? `${((jobs.filter(j => j.status === 'completed').length / jobs.length) * 100).toFixed(1)}%`
                  : '0%'}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Typography color="textSecondary" gutterBottom>
                Total Cost
              </Typography>
              <Typography variant="h4" color="warning.main">
                ${jobs?.reduce((sum, job) => sum + (job.actual_cost || job.estimated_cost), 0).toFixed(2) || '0.00'}
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Tabs */}
      <Paper sx={{ mb: 2 }}>
        <Tabs value={selectedTab} onChange={(_, value) => setSelectedTab(value)}>
          <Tab label={`All (${jobs?.length || 0})`} />
          <Tab label={`Active (${jobs?.filter(j => ['pending', 'scheduled', 'running'].includes(j.status)).length || 0})`} />
          <Tab label={`Completed (${jobs?.filter(j => j.status === 'completed').length || 0})`} />
          <Tab label={`Failed (${jobs?.filter(j => j.status === 'failed').length || 0})`} />
        </Tabs>
      </Paper>

      {/* Jobs Table */}
      <Paper sx={{ height: 600 }}>
        <DataGrid
          rows={filteredJobs || []}
          columns={columns}
          loading={isLoading}
          disableRowSelectionOnClick
          pageSizeOptions={[10, 25, 50]}
          initialState={{
            pagination: {
              paginationModel: { pageSize: 10 },
            },
            sorting: {
              sortModel: [{ field: 'created_at', sort: 'desc' }],
            },
          }}
        />
      </Paper>

      {/* Create Job Dialog */}
      <Dialog open={createDialogOpen} onClose={() => setCreateDialogOpen(false)} maxWidth="md" fullWidth>
        <DialogTitle>Create New Job</DialogTitle>
        <DialogContent>
          <Grid container spacing={3} sx={{ mt: 1 }}>
            <Grid item xs={12} sm={6}>
              <FormControl fullWidth>
                <InputLabel>Job Type</InputLabel>
                <Select
                  value={formData.type}
                  label="Job Type"
                  onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                >
                  <MenuItem value="docker">Docker Container</MenuItem>
                  <MenuItem value="script">Script</MenuItem>
                  <MenuItem value="binary">Binary</MenuItem>
                  <MenuItem value="wasm">WebAssembly</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            
            <Grid item xs={12} sm={6}>
              <FormControl fullWidth>
                <Typography gutterBottom>Priority</Typography>
                <Slider
                  value={formData.priority}
                  onChange={(_, value) => setFormData({ ...formData, priority: value as number })}
                  min={0}
                  max={10}
                  marks
                  valueLabelDisplay="auto"
                />
                <FormHelperText>Higher priority jobs are scheduled first</FormHelperText>
              </FormControl>
            </Grid>

            {/* Type-specific fields */}
            {formData.type === 'docker' && (
              <>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="Docker Image"
                    placeholder="e.g., python:3.9-slim"
                    value={formData.docker_image || ''}
                    onChange={(e) => setFormData({ ...formData, docker_image: e.target.value })}
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="Command"
                    placeholder="e.g., python script.py --arg value"
                    value={formData.command || ''}
                    onChange={(e) => setFormData({ ...formData, command: e.target.value })}
                  />
                </Grid>
              </>
            )}

            {formData.type === 'script' && (
              <>
                <Grid item xs={12} sm={6}>
                  <FormControl fullWidth>
                    <InputLabel>Language</InputLabel>
                    <Select
                      value={formData.script_language || ''}
                      label="Language"
                      onChange={(e) => setFormData({ ...formData, script_language: e.target.value })}
                    >
                      <MenuItem value="python">Python</MenuItem>
                      <MenuItem value="javascript">JavaScript</MenuItem>
                      <MenuItem value="bash">Bash</MenuItem>
                      <MenuItem value="r">R</MenuItem>
                    </Select>
                  </FormControl>
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    multiline
                    rows={6}
                    label="Script Content"
                    value={formData.script_content || ''}
                    onChange={(e) => setFormData({ ...formData, script_content: e.target.value })}
                  />
                </Grid>
              </>
            )}

            {/* Resource Requirements */}
            <Grid item xs={12}>
              <Typography variant="h6" gutterBottom>
                Resource Requirements
              </Typography>
            </Grid>

            <Grid item xs={12} sm={4}>
              <TextField
                fullWidth
                type="number"
                label="CPU Cores"
                value={formData.cpu_cores}
                onChange={(e) => setFormData({ ...formData, cpu_cores: parseInt(e.target.value) || 1 })}
                inputProps={{ min: 1, max: 64 }}
              />
            </Grid>

            <Grid item xs={12} sm={4}>
              <TextField
                fullWidth
                type="number"
                label="Memory (GB)"
                value={formData.memory_gb}
                onChange={(e) => setFormData({ ...formData, memory_gb: parseInt(e.target.value) || 1 })}
                inputProps={{ min: 1, max: 512 }}
              />
            </Grid>

            <Grid item xs={12} sm={4}>
              <TextField
                fullWidth
                type="number"
                label="Storage (GB)"
                value={formData.storage_gb}
                onChange={(e) => setFormData({ ...formData, storage_gb: parseInt(e.target.value) || 10 })}
                inputProps={{ min: 1, max: 1000 }}
              />
            </Grid>

            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                type="number"
                label="GPU Count"
                value={formData.gpu_count}
                onChange={(e) => setFormData({ ...formData, gpu_count: parseInt(e.target.value) || 0 })}
                inputProps={{ min: 0, max: 8 }}
              />
            </Grid>

            {formData.gpu_count > 0 && (
              <Grid item xs={12} sm={6}>
                <Autocomplete
                  freeSolo
                  options={['NVIDIA A100', 'NVIDIA V100', 'NVIDIA T4', 'NVIDIA RTX 3090', 'NVIDIA RTX 4090']}
                  value={formData.gpu_type}
                  onChange={(_, value) => setFormData({ ...formData, gpu_type: value || '' })}
                  renderInput={(params) => (
                    <TextField {...params} label="GPU Type" placeholder="Optional" />
                  )}
                />
              </Grid>
            )}

            {/* Advanced Settings */}
            <Grid item xs={12}>
              <Typography variant="h6" gutterBottom>
                Advanced Settings
              </Typography>
            </Grid>

            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                type="number"
                label="Timeout (seconds)"
                value={formData.timeout}
                onChange={(e) => setFormData({ ...formData, timeout: parseInt(e.target.value) || 3600 })}
                inputProps={{ min: 60, max: 86400 }}
              />
            </Grid>

            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                type="number"
                label="Max Retries"
                value={formData.max_retries}
                onChange={(e) => setFormData({ ...formData, max_retries: parseInt(e.target.value) || 3 })}
                inputProps={{ min: 0, max: 10 }}
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={handleCreateJob}
            variant="contained"
            disabled={createJobMutation.isPending}
          >
            {createJobMutation.isPending ? <CircularProgress size={24} /> : 'Create Job'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Job Details Dialog */}
      {selectedJob && (
        <Dialog open={!!selectedJob} onClose={() => setSelectedJob(null)} maxWidth="md" fullWidth>
          <DialogTitle>
            Job Details - {selectedJob.id}
          </DialogTitle>
          <DialogContent>
            <Grid container spacing={2} sx={{ mt: 1 }}>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle2" color="textSecondary">Type</Typography>
                <Typography variant="body1">{selectedJob.type}</Typography>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle2" color="textSecondary">Status</Typography>
                <Box display="flex" alignItems="center" gap={1}>
                  {getJobStatusIcon(selectedJob.status)}
                  <Chip
                    label={selectedJob.status}
                    size="small"
                    color={jobStatusColors[selectedJob.status] || 'default'}
                  />
                </Box>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle2" color="textSecondary">Priority</Typography>
                <Typography variant="body1">{selectedJob.priority}/10</Typography>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle2" color="textSecondary">Assigned Agent</Typography>
                <Typography variant="body1">{selectedJob.assigned_agent_id || 'Not assigned'}</Typography>
              </Grid>
              <Grid item xs={12}>
                <Typography variant="subtitle2" color="textSecondary">Resource Requirements</Typography>
                <Typography variant="body1">
                  CPU: {selectedJob.requirements.cpu_cores} cores • 
                  Memory: {selectedJob.requirements.memory_mb / 1024}GB • 
                  Storage: {selectedJob.requirements.storage_mb / 1024}GB
                  {selectedJob.requirements.gpu_count > 0 && (
                    <> • GPU: {selectedJob.requirements.gpu_count} {selectedJob.requirements.gpu_type}</>
                  )}
                </Typography>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle2" color="textSecondary">Created At</Typography>
                <Typography variant="body1">
                  {format(new Date(selectedJob.created_at), 'PPpp')}
                </Typography>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle2" color="textSecondary">Estimated Cost</Typography>
                <Typography variant="body1" fontWeight="medium">
                  ${selectedJob.estimated_cost.toFixed(2)}
                </Typography>
              </Grid>
              {selectedJob.error_message && (
                <Grid item xs={12}>
                  <Alert severity="error">
                    <Typography variant="body2">{selectedJob.error_message}</Typography>
                  </Alert>
                </Grid>
              )}
              {selectedJob.progress !== undefined && (
                <Grid item xs={12}>
                  <Typography variant="subtitle2" color="textSecondary" gutterBottom>
                    Progress
                  </Typography>
                  <LinearProgress variant="determinate" value={selectedJob.progress} />
                  <Typography variant="body2" align="right" sx={{ mt: 1 }}>
                    {selectedJob.progress}%
                  </Typography>
                </Grid>
              )}
            </Grid>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setSelectedJob(null)}>Close</Button>
            {['pending', 'scheduled', 'running'].includes(selectedJob.status) && (
              <Button
                color="error"
                onClick={() => {
                  cancelJobMutation.mutate(selectedJob.id);
                  setSelectedJob(null);
                }}
              >
                Cancel Job
              </Button>
            )}
          </DialogActions>
        </Dialog>
      )}
    </Box>
  );
}

import { CircularProgress } from '@mui/material'; 