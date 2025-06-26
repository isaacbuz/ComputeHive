/**
 * Marketplace Client
 */

import { AxiosInstance } from 'axios';
import { Bid, Offer } from './types';

export class MarketplaceClient {
  constructor(private http: AxiosInstance) {}

  async getBids(jobId?: string): Promise<Bid[]> {
    const response = await this.http.get('/marketplace/bids', {
      params: { job_id: jobId }
    });
    return response.data;
  }

  async createBid(bid: Partial<Bid>): Promise<Bid> {
    const response = await this.http.post('/marketplace/bids', bid);
    return response.data;
  }

  async getOffers(): Promise<Offer[]> {
    const response = await this.http.get('/marketplace/offers');
    return response.data;
  }

  async createOffer(offer: Partial<Offer>): Promise<Offer> {
    const response = await this.http.post('/marketplace/offers', offer);
    return response.data;
  }
} 