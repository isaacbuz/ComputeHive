/**
 * Agent Client
 */

import { AxiosInstance } from 'axios';
import { Agent } from './types';

export class AgentClient {
  constructor(private http: AxiosInstance) {}

  async listAgents(): Promise<Agent[]> {
    const response = await this.http.get('/agents');
    return response.data;
  }

  async getAgent(agentId: string): Promise<Agent> {
    const response = await this.http.get(`/agents/${agentId}`);
    return response.data;
  }

  async getAgentMetrics(agentId: string): Promise<any> {
    const response = await this.http.get(`/agents/${agentId}/metrics`);
    return response.data;
  }
} 