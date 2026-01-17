/**
 * HTTP Transport
 * Sends events to the ingestion API
 */

import { IngestRequest, IngestResponse } from '../types';

export class HTTPTransport {
  private ingestionUrl: string;
  private apiKey: string;

  constructor(ingestionUrl: string, apiKey: string) {
    this.ingestionUrl = ingestionUrl;
    this.apiKey = apiKey;
  }

  /**
   * Send events to ingestion API
   */
  async send(request: IngestRequest): Promise<IngestResponse> {
    const url = this.ingestionUrl.endsWith('/v1/events')
      ? this.ingestionUrl
      : `${this.ingestionUrl}/v1/events`;

    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey,
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return response.json();
  }

  /**
   * Get API key (for queue)
   */
  getApiKey(): string {
    return this.apiKey;
  }
}
