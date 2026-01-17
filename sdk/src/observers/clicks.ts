/**
 * Click Event Observer
 */

import { SessionManager } from '../core/session';
import { EventQueue } from '../transport/queue';
import { Event } from '../types';

export class ClickObserver {
  private sessionManager: SessionManager;
  private eventQueue: EventQueue;
  private handler: ((e: MouseEvent) => void) | null = null;

  constructor(sessionManager: SessionManager, eventQueue: EventQueue) {
    this.sessionManager = sessionManager;
    this.eventQueue = eventQueue;
  }

  start(): void {
    if (typeof document === 'undefined') {
      return;
    }

    this.handler = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      if (!target) return;

      // Skip if target is an input/button that will trigger other events
      if (target.tagName === 'INPUT' || target.tagName === 'BUTTON') {
        return; // Will be captured by input observer
      }

      const event: Event = {
        eventType: 'click',
        timestamp: new Date().toISOString(),
        sessionId: this.sessionManager.getSessionId(),
        route: this.sessionManager.getCurrentRoute(),
        target: {
          type: 'element',
          id: target.id || undefined,
          selector: this.getSelector(target),
          tagName: target.tagName.toLowerCase(),
        },
        metadata: {
          clientX: e.clientX,
          clientY: e.clientY,
        },
      };

      this.eventQueue.add(event);
    };

    document.addEventListener('click', this.handler, true);
  }

  stop(): void {
    if (this.handler) {
      document.removeEventListener('click', this.handler, true);
      this.handler = null;
    }
  }

  private getSelector(element: HTMLElement): string {
    if (element.id) {
      return `#${element.id}`;
    }
    if (element.className && typeof element.className === 'string') {
      const classes = element.className.split(' ').filter(Boolean);
      if (classes.length > 0) {
        return `.${classes[0]}`;
      }
    }
    return element.tagName.toLowerCase();
  }
}
