/**
 * Type definitions for HawkEye Observer SDK
 */

export interface EventTarget {
  type: string;
  id?: string;
  selector?: string;
  tagName?: string;
}

export interface Event {
  eventType: string;
  timestamp: string;
  sessionId: string;
  route: string;
  target: EventTarget;
  metadata?: Record<string, any>;
  environment?: string;        // "production", "staging", "development"
  idempotencyKey?: string;     // Unique key to prevent duplicate signal processing
}

export interface IngestRequest {
  api_key: string;
  sdk_version?: string;
  app_id?: string;
  events: Event[];
}

export interface IngestResponse {
  success: boolean;
  processed?: number;
  message?: string;
}

export interface SDKConfig {
  apiKey: string;
  ingestionUrl: string;
  enableDebug?: boolean;
  batchSize?: number;
  batchInterval?: number;
  projectId?: string;
  appId?: string;
  environment?: string;       // "production", "staging", "development" - defaults to auto-detect
}

export interface SessionInfo {
  sessionId: string;
  startTime: number;
  route: string;
}
