/**
 * ComputeHive JavaScript/TypeScript SDK
 * Main entry point
 */

export { ComputeHiveClient } from './client';
export { Job, JobStatus, ResourceRequirements, JobType } from './types';
export { MarketplaceClient } from './marketplace';
export { AgentClient } from './agent';
export { PaymentClient } from './payment';
export { TelemetryClient } from './telemetry';
export { ComputeHiveError, AuthenticationError, APIError } from './errors';
export { EventEmitter } from './events';

// Re-export commonly used types
export type {
  CreateJobRequest,
  JobResult,
  Agent,
  Bid,
  Offer,
  Payment,
  Invoice,
  Metric,
  Alert,
  AuthResponse,
  UserProfile
} from './types'; 