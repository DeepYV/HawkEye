/**
 * Scroll Event Observer
 * Throttled to avoid excessive events
 */

import { SessionManager } from '../core/session';
import { EventQueue } from '../transport/queue';
import { Event } from '../types';

export class ScrollObserver {
  private sessionManager: SessionManager;
  private eventQueue: EventQueue;
  private handler: (() => void) | null = null;
  private throttleTimeout: number | null = null;
  private lastScrollY = 0;

  constructor(sessionManager: SessionManager, eventQueue: EventQueue) {
    this.sessionManager = sessionManager;
    this.eventQueue = eventQueue;
  }

  start(): void {
    if (typeof window === 'undefined') {
      return;
    }

    this.lastScrollY = window.scrollY;

    this.handler = () => {
      // Throttle scroll events (max once per 500ms)
      if (this.throttleTimeout !== null) {
        return;
      }

      this.throttleTimeout = window.setTimeout(() => {
        const currentScrollY = window.scrollY;
        const scrollDelta = Math.abs(currentScrollY - this.lastScrollY);

        // Only capture significant scrolls (> 100px)
        if (scrollDelta > 100) {
          const event: Event = {
            eventType: 'scroll',
            timestamp: new Date().toISOString(),
            sessionId: this.sessionManager.getSessionId(),
            route: this.sessionManager.getCurrentRoute(),
            target: {
              type: 'window',
            },
            metadata: {
              scrollY: currentScrollY,
              scrollDelta,
            },
          };

          this.eventQueue.add(event);
          this.lastScrollY = currentScrollY;
        }

        this.throttleTimeout = null;
      }, 500);
    };

    window.addEventListener('scroll', this.handler, { passive: true });
  }

  stop(): void {
    if (this.handler) {
      window.removeEventListener('scroll', this.handler);
      if (this.throttleTimeout !== null) {
        clearTimeout(this.throttleTimeout);
        this.throttleTimeout = null;
      }
      this.handler = null;
    }
  }
}
