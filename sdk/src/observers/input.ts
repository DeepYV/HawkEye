/**
 * Input Event Observer
 * Captures focus, blur, and form submission events
 */

import { SessionManager } from '../core/session';
import { EventQueue } from '../transport/queue';
import { Event } from '../types';

export class InputObserver {
  private sessionManager: SessionManager;
  private eventQueue: EventQueue;
  private focusHandler: ((e: FocusEvent) => void) | null = null;
  private blurHandler: ((e: FocusEvent) => void) | null = null;
  private submitHandler: ((e: SubmitEvent) => void) | null = null;

  constructor(sessionManager: SessionManager, eventQueue: EventQueue) {
    this.sessionManager = sessionManager;
    this.eventQueue = eventQueue;
  }

  start(): void {
    if (typeof document === 'undefined') {
      return;
    }

    // Focus events
    this.focusHandler = (e: FocusEvent) => {
      const target = e.target as HTMLElement;
      if (!target || (target.tagName !== 'INPUT' && target.tagName !== 'TEXTAREA')) {
        return;
      }

      const event: Event = {
        eventType: 'input_focus',
        timestamp: new Date().toISOString(),
        sessionId: this.sessionManager.getSessionId(),
        route: this.sessionManager.getCurrentRoute(),
        target: {
          type: 'input',
          id: target.id || undefined,
          tagName: target.tagName.toLowerCase(),
        },
        metadata: {
          inputType: (target as HTMLInputElement).type || 'text',
        },
      };

      this.eventQueue.add(event);
    };

    // Blur events
    this.blurHandler = (e: FocusEvent) => {
      const target = e.target as HTMLElement;
      if (!target || (target.tagName !== 'INPUT' && target.tagName !== 'TEXTAREA')) {
        return;
      }

      const input = target as HTMLInputElement;
      const valueLength = input.value?.length || 0;

      const event: Event = {
        eventType: 'input_blur',
        timestamp: new Date().toISOString(),
        sessionId: this.sessionManager.getSessionId(),
        route: this.sessionManager.getCurrentRoute(),
        target: {
          type: 'input',
          id: target.id || undefined,
          tagName: target.tagName.toLowerCase(),
        },
        metadata: {
          inputType: input.type || 'text',
          valueLength, // Only length, not actual value (privacy)
        },
      };

      this.eventQueue.add(event);
    };

    // Form submission
    this.submitHandler = (e: SubmitEvent) => {
      const form = e.target as HTMLFormElement;
      if (!form || form.tagName !== 'FORM') {
        return;
      }

      const event: Event = {
        eventType: 'form_submit',
        timestamp: new Date().toISOString(),
        sessionId: this.sessionManager.getSessionId(),
        route: this.sessionManager.getCurrentRoute(),
        target: {
          type: 'form',
          id: form.id || undefined,
        },
        metadata: {
          fieldCount: form.elements.length,
        },
      };

      this.eventQueue.add(event);
    };

    document.addEventListener('focus', this.focusHandler, true);
    document.addEventListener('blur', this.blurHandler, true);
    document.addEventListener('submit', this.submitHandler, true);
  }

  stop(): void {
    if (this.focusHandler) {
      document.removeEventListener('focus', this.focusHandler, true);
      this.focusHandler = null;
    }
    if (this.blurHandler) {
      document.removeEventListener('blur', this.blurHandler, true);
      this.blurHandler = null;
    }
    if (this.submitHandler) {
      document.removeEventListener('submit', this.submitHandler, true);
      this.submitHandler = null;
    }
  }
}
