/**
 * Telemetry Client
 */

import { AxiosInstance } from 'axios';
import { Metric, Alert } from './types';

export class TelemetryClient {
  constructor(private http: AxiosInstance) {}

  async sendMetrics(metrics: Metric[]): Promise<void> {
    await this.http.post('/telemetry/metrics', metrics);
  }

  async queryMetrics(params: {
    metric: string;
    start?: string;
    end?: string;
    agent_id?: string;
  }): Promise<Metric[]> {
    const response = await this.http.get('/telemetry/metrics/query', { params });
    return response.data;
  }

  async createAlert(alert: Partial<Alert>): Promise<Alert> {
    const response = await this.http.post('/telemetry/alerts', alert);
    return response.data;
  }

  async getAlerts(): Promise<Alert[]> {
    const response = await this.http.get('/telemetry/alerts');
    return response.data;
  }
} 