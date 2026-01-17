/**
 * Navigation Observer
 * Tracks route changes (for SPAs)
 */

import { SessionManager } from '../core/session';
import { EventQueue } from '../transport/queue';
import { Event } from '../types';

export class NavigationObserver {
  private sessionManager: SessionManager;
  private eventQueue: EventQueue;
  private popstateHandler: (() => void) | null = null;
  private pushstateHandler: (() => void) | null = null;
  private lastRoute: string;

  constructor(sessionManager: SessionManager, eventQueue: EventQueue) {
    this.sessionManager = sessionManager;
    this.eventQueue = eventQueue;
    this.lastRoute = this.getCurrentRoute();
  }

  start(): void {
    if (typeof window === 'undefined') {
      return;
    }

    // Track popstate (browser back/forward)
    this.popstateHandler = () => {
      this.handleRouteChange();
    };

    // Track pushstate/replacestate (SPA navigation)
    const originalPushState = history.pushState;
    const originalReplaceState = history.replaceState;

    history.pushState = (...args) => {
      originalPushState.apply(history, args);
      this.handleRouteChange();
    };

    history.replaceState = (...args) => {
      originalReplaceState.apply(history, args);
      this.handleRouteChange();
    };

    window.addEventListener('popstate', this.popstateHandler);

    // Initial route
    this.handleRouteChange();
  }

  stop(): void {
    if (this.popstateHandler) {
      window.removeEventListener('popstate', this.popstateHandler);
      this.popstateHandler = null;
    }
  }

  private handleRouteChange(): void {
    const newRoute = this.getCurrentRoute();
    if (newRoute !== this.lastRoute) {
      this.sessionManager.updateRoute(newRoute);

      const event: Event = {
        eventType: 'navigation',
        timestamp: new Date().toISOString(),
        sessionId: this.sessionManager.getSessionId(),
        route: newRoute,
        target: {
          type: 'route',
        },
        metadata: {
          from: this.lastRoute,
          to: newRoute,
        },
      };

      this.eventQueue.add(event);
      this.lastRoute = newRoute;
    }
  }

  private getCurrentRoute(): string {
    if (typeof window !== 'undefined') {
      return window.location.pathname + window.location.search;
    }
    return '/';
  }
}
