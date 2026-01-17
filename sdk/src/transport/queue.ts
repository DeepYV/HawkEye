/**
 * Event Queue
 * Batches events and sends them periodically
 */

import { Event, IngestRequest } from '../types';
import { HTTPTransport } from './http';

export class EventQueue {
  private queue: Event[] = [];
  private transport: HTTPTransport;
  private batchSize: number;
  private batchInterval: number;
  private intervalId: number | null = null;
  private debugEnabled: boolean;

  constructor(
    transport: HTTPTransport,
    batchSize: number,
    batchInterval: number,
    debugEnabled: boolean = false
  ) {
    this.transport = transport;
    this.batchSize = batchSize;
    this.batchInterval = batchInterval;
    this.debugEnabled = debugEnabled;
    this.startBatchTimer();
  }

  /**
   * Add event to queue
   */
  add(event: Event): void {
    this.queue.push(event);

    // Send immediately if batch size reached
    if (this.queue.length >= this.batchSize) {
      this.flush();
    }
  }

  /**
   * Flush queue (send all pending events)
   */
  async flush(): Promise<void> {
    if (this.queue.length === 0) {
      return;
    }

    const events = [...this.queue];
    this.queue = [];

    try {
      const request: IngestRequest = {
        api_key: this.transport.getApiKey(),
        sdk_version: '1.0.0',
        events,
      };

      await this.transport.send(request);
    } catch (error) {
      // On error, put events back in queue (simple retry)
      // In production, you might want more sophisticated retry logic
      this.queue.unshift(...events);

      if (this.debugEnabled) {
        console.error('HawkEye SDK: Failed to send events', error);
      }
    }
  }

  /**
   * Start batch timer
   */
  private startBatchTimer(): void {
    if (typeof window === 'undefined') {
      return;
    }

    this.intervalId = window.setInterval(() => {
      this.flush().catch(() => {
        // Ignore errors in timer
      });
    }, this.batchInterval);
  }

  /**
   * Stop batch timer
   */
  stop(): void {
    if (this.intervalId !== null) {
      if (typeof window !== 'undefined') {
        clearInterval(this.intervalId);
      }
      this.intervalId = null;
    }
  }
}
