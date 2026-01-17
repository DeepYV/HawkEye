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
}

export interface SessionInfo {
  sessionId: string;
  startTime: number;
  route: string;
}
