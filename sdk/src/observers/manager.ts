/**
 * Observer Manager
 * Manages all event observers
 */

import { SessionManager } from '../core/session';
import { EventQueue } from '../transport/queue';
import { ConfigManager } from '../core/config';
import { ClickObserver } from './clicks';
import { ScrollObserver } from './scroll';
import { InputObserver } from './input';
import { ErrorObserver } from './errors';
import { NavigationObserver } from './navigation';

export class ObserverManager {
  private sessionManager: SessionManager;
  private eventQueue: EventQueue;
  private config: ConfigManager;
  private observers: Array<{ stop: () => void }> = [];

  constructor(
    sessionManager: SessionManager,
    eventQueue: EventQueue,
    config: ConfigManager
  ) {
    this.sessionManager = sessionManager;
    this.eventQueue = eventQueue;
    this.config = config;
  }

  /**
   * Start all observers
   */
  start(): void {
    if (typeof window === 'undefined') {
      return; // Server-side rendering, skip
    }

    try {
      // Click observer
      const clickObserver = new ClickObserver(
        this.sessionManager,
        this.eventQueue
      );
      clickObserver.start();
      this.observers.push(clickObserver);

      // Scroll observer
      const scrollObserver = new ScrollObserver(
        this.sessionManager,
        this.eventQueue
      );
      scrollObserver.start();
      this.observers.push(scrollObserver);

      // Input observer
      const inputObserver = new InputObserver(
        this.sessionManager,
        this.eventQueue
      );
      inputObserver.start();
      this.observers.push(inputObserver);

      // Error observer
      const errorObserver = new ErrorObserver(
        this.sessionManager,
        this.eventQueue
      );
      errorObserver.start();
      this.observers.push(errorObserver);

      // Navigation observer
      const navigationObserver = new NavigationObserver(
        this.sessionManager,
        this.eventQueue
      );
      navigationObserver.start();
      this.observers.push(navigationObserver);

      // Handle page unload
      window.addEventListener('beforeunload', () => {
        this.eventQueue.flush().catch(() => {
          // Ignore errors during unload
        });
      });
    } catch (error) {
      if (this.config.isDebugEnabled()) {
        console.error('HawkEye SDK: Failed to start observers', error);
      }
    }
  }

  /**
   * Stop all observers
   */
  stop(): void {
    this.observers.forEach((observer) => {
      try {
        observer.stop();
      } catch (error) {
        if (this.config.isDebugEnabled()) {
          console.error('HawkEye SDK: Error stopping observer', error);
        }
      }
    });
    this.observers = [];
  }
}
