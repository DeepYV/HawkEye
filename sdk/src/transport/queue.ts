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
  private environment: string;
  private idempotencyCounter: number = 0;
  private sessionId: string;

  constructor(
    transport: HTTPTransport,
    batchSize: number,
    batchInterval: number,
    debugEnabled: boolean = false,
    environment: string = 'production',
    sessionId: string = ''
  ) {
    this.transport = transport;
    this.batchSize = batchSize;
    this.batchInterval = batchInterval;
    this.debugEnabled = debugEnabled;
    this.environment = environment;
    this.sessionId = sessionId;
    this.startBatchTimer();
  }

  /**
   * Generate unique idempotency key for event
   * Format: sessionId-timestamp-counter
   */
  private generateIdempotencyKey(timestamp: string): string {
    this.idempotencyCounter++;
    return `${this.sessionId}-${timestamp}-${this.idempotencyCounter}`;
  }

  /**
   * Add event to queue
   * Enriches event with environment and idempotency key
   */
  add(event: Event): void {
    // Enrich event with environment and idempotency key
    const enrichedEvent: Event = {
      ...event,
      environment: this.environment,
      idempotencyKey: this.generateIdempotencyKey(event.timestamp),
    };

    this.queue.push(enrichedEvent);

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
