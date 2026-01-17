/**
 * Error Event Observer
 * Captures JavaScript errors and unhandled promise rejections
 */

import { SessionManager } from '../core/session';
import { EventQueue } from '../transport/queue';
import { Event } from '../types';

export class ErrorObserver {
  private sessionManager: SessionManager;
  private eventQueue: EventQueue;
  private errorHandler: ((e: ErrorEvent) => void) | null = null;
  private unhandledRejectionHandler: ((e: PromiseRejectionEvent) => void) | null = null;

  constructor(sessionManager: SessionManager, eventQueue: EventQueue) {
    this.sessionManager = sessionManager;
    this.eventQueue = eventQueue;
  }

  start(): void {
    if (typeof window === 'undefined') {
      return;
    }

    // JavaScript errors
    this.errorHandler = (e: ErrorEvent) => {
      const event: Event = {
        eventType: 'error',
        timestamp: new Date().toISOString(),
        sessionId: this.sessionManager.getSessionId(),
        route: this.sessionManager.getCurrentRoute(),
        target: {
          type: 'error',
        },
        metadata: {
          message: e.message || 'Unknown error',
          filename: e.filename || undefined,
          lineno: e.lineno || undefined,
          colno: e.colno || undefined,
          // Don't include stack trace in metadata (too verbose)
        },
      };

      this.eventQueue.add(event);
    };

    // Unhandled promise rejections
    this.unhandledRejectionHandler = (e: PromiseRejectionEvent) => {
      const reason = e.reason;
      const message = reason?.message || String(reason) || 'Unhandled promise rejection';

      const event: Event = {
        eventType: 'unhandled_rejection',
        timestamp: new Date().toISOString(),
        sessionId: this.sessionManager.getSessionId(),
        route: this.sessionManager.getCurrentRoute(),
        target: {
          type: 'promise',
        },
        metadata: {
          message,
        },
      };

      this.eventQueue.add(event);
    };

    window.addEventListener('error', this.errorHandler);
    window.addEventListener('unhandledrejection', this.unhandledRejectionHandler);
  }

  stop(): void {
    if (this.errorHandler) {
      window.removeEventListener('error', this.errorHandler);
      this.errorHandler = null;
    }
    if (this.unhandledRejectionHandler) {
      window.removeEventListener('unhandledrejection', this.unhandledRejectionHandler);
      this.unhandledRejectionHandler = null;
    }
  }
}
