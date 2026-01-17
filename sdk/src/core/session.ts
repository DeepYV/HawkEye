/**
 * Session Management
 * Generates and manages session IDs
 */

export class SessionManager {
  private sessionId: string;
  private startTime: number;
  private currentRoute: string;

  constructor() {
    this.sessionId = this.generateSessionId();
    this.startTime = Date.now();
    this.currentRoute = this.getCurrentRoute();
  }

  /**
   * Generate a UUID v4 session ID
   */
  private generateSessionId(): string {
    // UUID v4 format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
      const r = (Math.random() * 16) | 0;
      const v = c === 'x' ? r : (r & 0x3) | 0x8;
      return v.toString(16);
    });
  }

  /**
   * Get current route/path
   */
  private getCurrentRoute(): string {
    if (typeof window !== 'undefined') {
      return window.location.pathname + window.location.search;
    }
    return '/';
  }

  /**
   * Get current session ID
   */
  getSessionId(): string {
    return this.sessionId;
  }

  /**
   * Get session start time
   */
  getStartTime(): number {
    return this.startTime;
  }

  /**
   * Get current route
   */
  getCurrentRoute(): string {
    return this.currentRoute;
  }

  /**
   * Update current route (called on navigation)
   */
  updateRoute(route: string): void {
    this.currentRoute = route;
  }

  /**
   * Reset session (for testing or explicit reset)
   */
  reset(): void {
    this.sessionId = this.generateSessionId();
    this.startTime = Date.now();
    this.currentRoute = this.getCurrentRoute();
  }
}
