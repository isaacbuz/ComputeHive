/**
 * Payment Client
 */

import { AxiosInstance } from 'axios';
import { Payment, Invoice } from './types';

export class PaymentClient {
  constructor(private http: AxiosInstance) {}

  async getBalance(): Promise<any> {
    const response = await this.http.get('/payments/balance');
    return response.data;
  }

  async getPaymentHistory(): Promise<Payment[]> {
    const response = await this.http.get('/payments');
    return response.data;
  }

  async getInvoices(): Promise<Invoice[]> {
    const response = await this.http.get('/payments/invoices');
    return response.data;
  }

  async deposit(amount: number, currency: string = 'USD'): Promise<Payment> {
    const response = await this.http.post('/payments/deposit', {
      amount,
      currency
    });
    return response.data;
  }
} 