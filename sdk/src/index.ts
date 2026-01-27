/**
 * HawkEye Observer SDK
 * Main entry point
 */

export { initFrustrationObserver, getSessionId, captureEvent, flushEvents, teardown } from './core/init';
export type { SDKConfig, Event, EventTarget, IngestRequest, IngestResponse } from './types';
