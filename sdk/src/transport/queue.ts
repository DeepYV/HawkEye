import { BATCH_INTERVAL_MS, BATCH_SIZE, SDK_VERSION } from '../pure/constants.js';
import type { WireEvent } from '../types/wire.js';
import { getOriginalFetch } from './fetch.js';

export type FlushMode = 'normal' | 'keepalive' | 'unload';

export function eventsEndpoint(ingestionUrl: string): string {
  const trimmed = ingestionUrl.replace(/\/+$/, '');
  if (trimmed.endsWith('/v1/events')) return trimmed;
  return `${trimmed}/v1/events`;
}

export class EventQueue {
  private events: WireEvent[] = [];
  private timer: ReturnType<typeof setInterval> | null = null;
  private sending = false;

  constructor(
    private readonly url: string,
    private readonly apiKey: string,
  ) {
    if (typeof window !== 'undefined') {
      this.timer = setInterval(() => {
        void this.flush('normal');
      }, BATCH_INTERVAL_MS);
    }
  }

  add(event: WireEvent): void {
    this.events.push(event);
    if (this.events.length >= BATCH_SIZE) {
      void this.flush('normal');
    }
  }

  async flush(mode: FlushMode): Promise<void> {
    if (this.events.length === 0) return;
    if (this.sending && mode === 'normal') return;
    const batch = this.events.splice(0, this.events.length);
    const body = JSON.stringify({ sdk_version: SDK_VERSION, events: batch });
    if (mode === 'unload') {
      this.postUnload(body);
      return;
    }
    this.sending = true;
    try {
      await this.post(body, mode === 'keepalive');
    } catch {
      this.events = batch.concat(this.events);
    } finally {
      this.sending = false;
    }
  }

  stop(): void {
    if (this.timer !== null) {
      clearInterval(this.timer);
      this.timer = null;
    }
  }

  private async post(body: string, keepalive: boolean): Promise<void> {
    const fetchFn = getOriginalFetch();
    const init: RequestInit = {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey,
      },
      body,
    };
    if (keepalive) init.keepalive = true;
    await fetchFn(this.url, init);
  }

  /** Fetch with keepalive. sendBeacon is only used when that call cannot be started. */
  private postUnload(body: string): void {
    try {
      const fetchFn = getOriginalFetch();
      const pending = fetchFn(this.url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey,
        },
        body,
        keepalive: true,
      });
      void Promise.resolve(pending).catch(() => {
        this.beacon(body);
      });
    } catch {
      this.beacon(body);
    }
  }

  private beacon(body: string): void {
    if (typeof navigator === 'undefined' || typeof navigator.sendBeacon !== 'function') return;
    const join = this.url.includes('?') ? '&' : '?';
    const beaconUrl = `${this.url}${join}api_key=${encodeURIComponent(this.apiKey)}`;
    try {
      navigator.sendBeacon(beaconUrl, new Blob([body], { type: 'application/json' }));
    } catch {
      // The page is going away; there is nowhere to retry.
    }
  }
}
