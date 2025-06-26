/**
 * ComputeHive SDK Type Definitions
 */

// Job related types
export enum JobStatus {
  PENDING = 'pending',
  RUNNING = 'running',
  COMPLETED = 'completed',
  FAILED = 'failed'
}

export enum JobType {
  BATCH = 'batch',
  STREAMING = 'streaming'
}

export interface ResourceRequirements {
  cpu?: { cores: number };
  memory?: { size: number };
  gpu?: { count: number; model?: string };
}

export interface Job {
  id: string;
  name: string;
  status: JobStatus;
  type: JobType;
  resource_requirements: ResourceRequirements;
}

export interface CreateJobRequest {
  name: string;
  type: JobType;
  docker_image?: string;
  resource_requirements: ResourceRequirements;
}

export interface JobResult {
  exit_code?: number;
  stdout?: string;
  stderr?: string;
}

// Agent related types
export interface Agent {
  id: string;
  name: string;
  status: string;
  resources: any;
}

// Marketplace types
export interface Bid {
  id: string;
  agent_id: string;
  job_id: string;
  price: number;
}

export interface Offer {
  id: string;
  agent_id: string;
  price_per_unit: number;
}

// Payment types
export interface Payment {
  id: string;
  amount: number;
  status: string;
}

export interface Invoice {
  id: string;
  total_amount: number;
  status: string;
}

// Telemetry types
export interface Metric {
  name: string;
  value: number;
  timestamp: string;
}

export interface Alert {
  id: string;
  name: string;
  metric_name: string;
}

// Auth types
export interface AuthResponse {
  access_token: string;
  user: UserProfile;
}

export interface UserProfile {
  id: string;
  email: string;
  username: string;
}

// WebSocket event types
export interface WebSocketEvent {
  type: string;
  data: any;
  timestamp: string;
}

export interface JobUpdateEvent extends WebSocketEvent {
  type: 'job.update';
  data: {
    job_id: string;
    status: JobStatus;
    progress?: number;
    message?: string;
  };
}

export interface MetricStreamEvent extends WebSocketEvent {
  type: 'metrics.stream';
  data: Metric[];
}

export interface MarketplaceEvent extends WebSocketEvent {
  type: 'marketplace.bid' | 'marketplace.offer' | 'marketplace.match';
  data: Bid | Offer;
}

// API response types
export interface ApiResponse<T> {
  data: T;
  status: number;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface ErrorResponse {
  error: string;
  message: string;
  status: number;
  details?: any;
} 