const ANON_KEY = 'hawkeye_anonymous_id';
const SESSION_KEY = 'hawkeye_session_id';
const PAGE_KEY = 'hawkeye_page_seq';
const ATTEMPT_KEY = 'hawkeye_attempt_id';
const FLOW_KEY = 'hawkeye_flow_state';

export type StoredFlow = {
  name: string;
  attemptId: string;
  furthest: number;
  sentDef: boolean;
};

const fallback = new Map<string, string>();
let localStore: Storage | null | undefined;
let sessionStore: Storage | null | undefined;
let booted = false;
let pageSeq = 1;

export function randomId(): string {
  const cryptoObj = globalThis.crypto;
  if (cryptoObj && typeof cryptoObj.randomUUID === 'function') return cryptoObj.randomUUID();
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (ch) => {
    const r = (Math.random() * 16) | 0;
    const v = ch === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

function openStore(kind: 'local' | 'session'): Storage | null {
  try {
    const store = kind === 'local' ? window.localStorage : window.sessionStorage;
    const probe = '__hawkeye_probe__';
    store.setItem(probe, '1');
    store.removeItem(probe);
    return store;
  } catch {
    return null;
  }
}

function storeFor(kind: 'local' | 'session'): Storage | null {
  if (kind === 'local') {
    if (localStore === undefined) localStore = openStore('local');
    return localStore;
  }
  if (sessionStore === undefined) sessionStore = openStore('session');
  return sessionStore;
}

function read(kind: 'local' | 'session', key: string): string | null {
  const store = storeFor(kind);
  if (store) {
    try {
      return store.getItem(key);
    } catch {
      return fallback.get(`${kind}:${key}`) ?? null;
    }
  }
  return fallback.get(`${kind}:${key}`) ?? null;
}

function write(kind: 'local' | 'session', key: string, value: string): void {
  const store = storeFor(kind);
  if (store) {
    try {
      store.setItem(key, value);
      return;
    } catch {
      // Fall through to memory when storage is full or blocked.
    }
  }
  fallback.set(`${kind}:${key}`, value);
}

function remove(kind: 'local' | 'session', key: string): void {
  const store = storeFor(kind);
  if (store) {
    try {
      store.removeItem(key);
    } catch {
      // Ignore storage failures.
    }
  }
  fallback.delete(`${kind}:${key}`);
}

function bootPageSeq(): number {
  if (booted) return pageSeq;
  booted = true;
  const raw = read('session', PAGE_KEY);
  if (raw == null || raw === '') {
    pageSeq = 1;
  } else {
    const parsed = Number.parseInt(raw, 10);
    pageSeq = Number.isFinite(parsed) && parsed >= 1 ? parsed + 1 : 1;
  }
  write('session', PAGE_KEY, String(pageSeq));
  return pageSeq;
}

export function loadIdentity(): { anonymousId: string; sessionId: string; pageSeq: number } {
  let anonymousId = read('local', ANON_KEY);
  if (!anonymousId) {
    anonymousId = randomId();
    write('local', ANON_KEY, anonymousId);
  }
  let sessionId = read('session', SESSION_KEY);
  if (!sessionId) {
    sessionId = randomId();
    write('session', SESSION_KEY, sessionId);
  }
  return { anonymousId, sessionId, pageSeq: bootPageSeq() };
}

export function bumpPageSeq(): number {
  pageSeq += 1;
  write('session', PAGE_KEY, String(pageSeq));
  return pageSeq;
}

export function readStoredFlow(): StoredFlow | null {
  const raw = read('session', FLOW_KEY);
  const attemptId = read('session', ATTEMPT_KEY);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as Partial<StoredFlow>;
    if (!parsed || typeof parsed.name !== 'string' || typeof parsed.attemptId !== 'string') return null;
    return {
      name: parsed.name,
      attemptId: parsed.attemptId || attemptId || '',
      furthest: typeof parsed.furthest === 'number' ? parsed.furthest : 0,
      sentDef: parsed.sentDef !== false,
    };
  } catch {
    return null;
  }
}

export function writeStoredFlow(state: StoredFlow | null): void {
  if (!state) {
    remove('session', FLOW_KEY);
    remove('session', ATTEMPT_KEY);
    return;
  }
  write('session', ATTEMPT_KEY, state.attemptId);
  write('session', FLOW_KEY, JSON.stringify(state));
}
