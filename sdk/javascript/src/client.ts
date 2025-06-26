/**
 * ComputeHive Client
 * Main client for interacting with ComputeHive API
 */

import axios, { AxiosInstance, AxiosError } from 'axios';
import WebSocket from 'ws';
import { EventEmitter } from './events';
import { AuthResponse, Job, CreateJobRequest, JobStatus } from './types';
import { APIError, AuthenticationError, NetworkError } from './errors';

export interface ComputeHiveConfig {
  apiUrl?: string;
  wsUrl?: string;
  apiKey?: string;
  accessToken?: string;
  timeout?: number;
  maxRetries?: number;
  debug?: boolean;
}

export class ComputeHiveClient extends EventEmitter {
  private config: ComputeHiveConfig;
  private http: AxiosInstance;
  private ws?: WebSocket;
  private accessToken?: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectInterval = 1000;

  constructor(config: ComputeHiveConfig = {}) {
    super();
    
    this.config = {
      apiUrl: config.apiUrl || 'https://api.computehive.io',
      wsUrl: config.wsUrl || 'wss://api.computehive.io/ws',
      timeout: config.timeout || 30000,
      maxRetries: config.maxRetries || 3,
      debug: config.debug || false,
      ...config
    };

    this.accessToken = config.accessToken || config.apiKey;

    // Initialize HTTP client
    this.http = axios.create({
      baseURL: this.config.apiUrl,
      timeout: this.config.timeout,
      headers: {
        'Content-Type': 'application/json',
        ...(this.accessToken && { Authorization: `Bearer ${this.accessToken}` })
      }
    });

    // Add request/response interceptors
    this.setupInterceptors();
  }

  /**
   * Authenticate with ComputeHive
   */
  async authenticate(email: string, password: string): Promise<AuthResponse> {
    try {
      const response = await this.http.post<AuthResponse>('/auth/login', {
        email,
        password
      });

      this.accessToken = response.data.access_token;
      this.http.defaults.headers['Authorization'] = `Bearer ${this.accessToken}`;
      
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  /**
   * Connect to WebSocket for real-time updates
   */
  connect(): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      return;
    }

    const wsUrl = `${this.config.wsUrl}${this.accessToken ? `?token=${this.accessToken}` : ''}`;
    
    this.ws = new WebSocket(wsUrl);

    this.ws.on('open', () => {
      this.reconnectAttempts = 0;
      this.emit('connected');
      if (this.config.debug) console.log('WebSocket connected');
    });

    this.ws.on('message', (data: Buffer) => {
      try {
        const message = JSON.parse(data.toString());
        this.handleWebSocketMessage(message);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    });

    this.ws.on('close', (code: number, reason: Buffer) => {
      this.emit('disconnected', reason.toString());
      if (this.config.debug) console.log('WebSocket disconnected:', code, reason.toString());
      this.attemptReconnect();
    });

    this.ws.on('error', (error: Error) => {
      this.emit('error', error);
      if (this.config.debug) console.error('WebSocket error:', error);
    });
  }

  /**
   * Disconnect from WebSocket
   */
  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = undefined;
    }
  }

  /**
   * Create a new job
   */
  async createJob(request: CreateJobRequest): Promise<Job> {
    try {
      const response = await this.http.post<Job>('/jobs', request);
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  /**
   * Get job by ID
   */
  async getJob(jobId: string): Promise<Job> {
    try {
      const response = await this.http.get<Job>(`/jobs/${jobId}`);
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  /**
   * List jobs
   */
  async listJobs(params?: {
    status?: JobStatus;
    limit?: number;
    offset?: number;
  }): Promise<Job[]> {
    try {
      const response = await this.http.get<Job[]>('/jobs', { params });
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  /**
   * Cancel a job
   */
  async cancelJob(jobId: string): Promise<void> {
    try {
      await this.http.post(`/jobs/${jobId}/cancel`);
    } catch (error) {
      throw this.handleError(error);
    }
  }

  /**
   * Get job logs
   */
  async getJobLogs(jobId: string): Promise<string> {
    try {
      const response = await this.http.get<{ logs: string }>(`/jobs/${jobId}/logs`);
      return response.data.logs;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  /**
   * Wait for job completion
   */
  async waitForJob(jobId: string, timeout: number = 300000): Promise<Job> {
    const startTime = Date.now();
    
    return new Promise((resolve, reject) => {
      const checkJob = async () => {
        try {
          const job = await this.getJob(jobId);
          
          if (job.status === JobStatus.COMPLETED) {
            resolve(job);
            return;
          }
          
          if (job.status === JobStatus.FAILED) {
            reject(new Error(`Job ${jobId} failed: ${job.error || 'Unknown error'}`));
            return;
          }
          
          if (Date.now() - startTime > timeout) {
            reject(new Error(`Timeout waiting for job ${jobId}`));
            return;
          }
          
          // Check again in 5 seconds
          setTimeout(checkJob, 5000);
        } catch (error) {
          reject(error);
        }
      };
      
      checkJob();
    });
  }

  /**
   * Get current user profile
   */
  async getProfile(): Promise<any> {
    try {
      const response = await this.http.get('/auth/profile');
      return response.data;
    } catch (error) {
      throw this.handleError(error);
    }
  }

  // Private methods

  private setupInterceptors(): void {
    // Request interceptor
    this.http.interceptors.request.use(
      (config) => {
        if (this.config.debug) {
          console.log(`[ComputeHive] ${config.method?.toUpperCase()} ${config.url}`);
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor
    this.http.interceptors.response.use(
      (response) => response,
      async (error: AxiosError) => {
        if (error.response?.status === 401 && !error.config?.url?.includes('/auth/')) {
          // Token expired, emit event
          this.emit('error', new AuthenticationError('Token expired'));
        }
        
        return Promise.reject(error);
      }
    );
  }

  private handleError(error: any): Error {
    if (axios.isAxiosError(error)) {
      if (error.response) {
        const { status, data } = error.response;
        
        if (status === 401) {
          return new AuthenticationError(data.message || 'Unauthorized');
        }
        
        return new APIError(
          data.message || error.message,
          status,
          data
        );
      } else if (error.request) {
        return new NetworkError('No response from server');
      }
    }
    
    return error;
  }

  private handleWebSocketMessage(message: any): void {
    const { type, data } = message;

    switch (type) {
      case 'job.updated':
        this.emit('job.updated', data);
        break;
      case 'job.completed':
        this.emit('job.completed', data);
        break;
      case 'job.failed':
        this.emit('job.failed', data, data.error);
        break;
      case 'job.progress':
        this.emit('job.progress', data.job_id, data.progress, data.message);
        break;
      case 'metrics.update':
        this.emit('metrics.update', data);
        break;
      case 'marketplace.bid':
        this.emit('marketplace.bid', data);
        break;
      case 'marketplace.offer':
        this.emit('marketplace.offer', data);
        break;
      case 'alert.triggered':
        this.emit('alert.triggered', data);
        break;
      default:
        if (this.config.debug) {
          console.log('Unknown WebSocket message type:', type);
        }
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.emit('error', new Error('Max reconnection attempts reached'));
      return;
    }

    this.reconnectAttempts++;
    this.emit('reconnecting', this.reconnectAttempts);

    setTimeout(() => {
      if (this.config.debug) {
        console.log(`Reconnection attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts}`);
      }
      this.connect();
    }, this.reconnectInterval * this.reconnectAttempts);
  }
} 