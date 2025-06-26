/**
 * ComputeHive SDK Event System
 */

import { EventEmitter as EE } from 'eventemitter3';
import { Job, JobStatus, Metric, Bid, Offer } from './types';

export interface ComputeHiveEvents {
  // Connection events
  'connected': () => void;
  'disconnected': (reason?: string) => void;
  'reconnecting': (attempt: number) => void;
  'error': (error: Error) => void;

  // Job events
  'job.created': (job: Job) => void;
  'job.updated': (job: Job) => void;
  'job.completed': (job: Job) => void;
  'job.failed': (job: Job, error: string) => void;
  'job.progress': (jobId: string, progress: number, message?: string) => void;

  // Marketplace events
  'marketplace.bid': (bid: Bid) => void;
  'marketplace.offer': (offer: Offer) => void;
  'marketplace.match': (match: any) => void;

  // Telemetry events
  'metrics.update': (metrics: Metric[]) => void;
  'alert.triggered': (alert: any) => void;
  'alert.resolved': (alert: any) => void;

  // Agent events
  'agent.online': (agentId: string) => void;
  'agent.offline': (agentId: string) => void;
  'agent.busy': (agentId: string) => void;
}

export class EventEmitter extends EE<ComputeHiveEvents> {
  constructor() {
    super();
    this.setMaxListeners(100); // Allow many listeners for production use
  }
} 