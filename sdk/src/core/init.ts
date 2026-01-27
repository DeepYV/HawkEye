/**
 * SDK Initialization
 * Main entry point for the HawkEye Observer SDK
 */

import { SDKConfig } from '../types';
import { ConfigManager } from './config';
import { SessionManager } from './session';
import { EventQueue } from '../transport/queue';
import { HTTPTransport } from '../transport/http';
import { ObserverManager } from '../observers/manager';

let isInitialized = false;
let configManager: ConfigManager | null = null;
let sessionManager: SessionManager | null = null;
let eventQueue: EventQueue | null = null;
let transport: HTTPTransport | null = null;
let observerManager: ObserverManager | null = null;

/**
 * Initialize the HawkEye Observer SDK
 */
export function initFrustrationObserver(config: SDKConfig): void {
  // Prevent double initialization
  if (isInitialized) {
    if (configManager?.isDebugEnabled()) {
      console.warn('HawkEye SDK: Already initialized, ignoring duplicate init call');
    }
    return;
  }

  try {
    // Validate and store configuration
    configManager = new ConfigManager(config);

    // Initialize session manager
    sessionManager = new SessionManager();

    // Initialize transport
    transport = new HTTPTransport(
      configManager.getIngestionUrl(),
      configManager.getApiKey()
    );

    // Initialize event queue with environment and session ID for idempotency keys
    eventQueue = new EventQueue(
      transport,
      configManager.getBatchSize(),
      configManager.getBatchInterval(),
      configManager.isDebugEnabled(),
      configManager.getEnvironment(),
      sessionManager.getSessionId()
    );

    // Initialize observers
    observerManager = new ObserverManager(
      sessionManager,
      eventQueue,
      configManager
    );

    // Start observing
    observerManager.start();

    isInitialized = true;

    if (configManager.isDebugEnabled()) {
      console.log('HawkEye SDK: Initialized successfully', {
        sessionId: sessionManager.getSessionId(),
        ingestionUrl: configManager.getIngestionUrl(),
      });
    }
  } catch (error) {
    // Fail silently in production, log in debug mode
    if (config.enableDebug) {
      console.error('HawkEye SDK: Initialization failed', error);
    }
  }
}

/**
 * Get current session ID (for testing/debugging)
 */
export function getSessionId(): string | null {
  return sessionManager?.getSessionId() || null;
}

/**
 * Manually capture an event (for custom events)
 */
export function captureEvent(eventType: string, target: any, metadata?: Record<string, any>): void {
  if (!isInitialized || !sessionManager || !eventQueue) {
    if (configManager?.isDebugEnabled()) {
      console.warn('HawkEye SDK: Not initialized, cannot capture event');
    }
    return;
  }

  const route = sessionManager.getCurrentRoute();
  const event = {
    eventType,
    timestamp: new Date().toISOString(),
    sessionId: sessionManager.getSessionId(),
    route,
    target: {
      type: target?.type || 'unknown',
      id: target?.id,
      selector: target?.selector,
      tagName: target?.tagName,
    },
    metadata: metadata || {},
  };

  eventQueue.add(event);
}

/**
 * Flush events immediately (for testing or before page unload)
 */
export function flushEvents(): Promise<void> {
  if (!eventQueue) {
    return Promise.resolve();
  }
  return eventQueue.flush();
}

/**
 * Teardown the SDK (cleanup resources, stop observers)
 */
export async function teardown(): Promise<void> {
  if (!isInitialized) {
    return;
  }

  try {
    // Flush any pending events
    if (eventQueue) {
      await eventQueue.flush();
      eventQueue.stop();
    }

    // Stop all observers
    if (observerManager) {
      observerManager.stop();
    }

    // Clear references
    isInitialized = false;
    configManager = null;
    sessionManager = null;
    eventQueue = null;
    transport = null;
    observerManager = null;

    if (configManager?.isDebugEnabled()) {
      console.log('HawkEye SDK: Teardown complete');
    }
  } catch (error) {
    if (configManager?.isDebugEnabled()) {
      console.error('HawkEye SDK: Teardown failed', error);
    }
  }
}
