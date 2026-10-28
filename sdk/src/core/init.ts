import { bumpPageSeq, loadIdentity, randomId, readStoredFlow, writeStoredFlow } from './identity.js';
import { installClicks, type WatchNotice } from '../observe/clicks.js';
import { clearPageUnloading, installErrors, markPageUnloading, resetErrorBudget } from '../observe/errors.js';
import { installNavigation, type RouteChange } from '../observe/navigation.js';
import {
  BUSINESS_NAME_RE,
  FLOW_NAME_RE,
  TOKEN_RE,
  clampFlowWindow,
  sanitizeBusinessProps,
  sanitizeFieldKeys,
  validReason,
} from '../pure/grammar.js';
import { eventsEndpoint, EventQueue } from '../transport/queue.js';
import type { FlowDefinition, FlowHandle, FormOutcomeInput, InitOptions } from '../types/api.js';
import type { WireEvent } from '../types/wire.js';

type FlowDef = {
  name: string;
  steps: string[];
  windowMs: number;
  important: boolean;
};

type Attempt = {
  name: string;
  steps: string[];
  windowMs: number;
  important: boolean;
  attemptId: string;
  furthest: number;
  sentDef: boolean;
};

type Runtime = {
  stopClicks: () => void;
  stopNav: () => void;
  stopErrors: () => void;
  stopPage: () => void;
  queue: EventQueue;
};

const noopHandle: FlowHandle = {
  start() {},
  step() {},
  complete() {},
  fail() {},
  abandon() {},
};

let initialized = false;
let loadSeen = false;
let runtime: Runtime | null = null;
let options: InitOptions | null = null;
let registry = new Map<string, FlowDef>();
let allowedEvents = new Set<string>();
let anonymousId = '';
let sessionId = '';
let pageSeq = 1;
let currentRoute = '/';
let attempt: Attempt | null = null;
let monoOrigin = 0;
let queue: EventQueue | null = null;
let notifyWatch: ((kind: WatchNotice, data?: { formKey?: string; flowId?: string }) => void) | null = null;
let setRouteImpl: ((template: string) => void) | null = null;

function mono(): number {
  if (typeof performance !== 'undefined' && typeof performance.now === 'function') {
    return performance.now() - monoOrigin;
  }
  return Date.now() - monoOrigin;
}

function envelope(): Omit<WireEvent, 'type'> {
  return {
    event_id: randomId(),
    timestamp: Date.now(),
    mono_ms: Math.round(mono()),
    page_seq: pageSeq,
    anonymous_id: anonymousId,
    session_id: sessionId,
    route: currentRoute,
    flow_id: attempt?.attemptId ?? '',
    flow_name: attempt?.name ?? '',
    app_version: options?.appVersion ?? '',
    app_build: options?.appBuild ?? '',
    app_namespace: options?.appNamespace ?? '',
    release_id: options?.releaseId ?? '',
    environment: options?.environment ?? '',
  };
}

function persistAttempt(): void {
  if (!attempt) {
    writeStoredFlow(null);
    return;
  }
  writeStoredFlow({
    name: attempt.name,
    attemptId: attempt.attemptId,
    furthest: attempt.furthest,
    sentDef: attempt.sentDef,
  });
}

function buildRegistry(flows: FlowDefinition[] | undefined): Map<string, FlowDef> {
  const map = new Map<string, FlowDef>();
  if (!flows) return map;
  for (const raw of flows) {
    if (!raw || typeof raw.name !== 'string' || !FLOW_NAME_RE.test(raw.name)) continue;
    if (!Array.isArray(raw.steps) || raw.steps.length === 0) continue;
    const steps: string[] = [];
    const seen = new Set<string>();
    let ok = true;
    for (const step of raw.steps) {
      if (typeof step !== 'string' || !FLOW_NAME_RE.test(step) || seen.has(step)) {
        ok = false;
        break;
      }
      seen.add(step);
      steps.push(step);
    }
    if (!ok) continue;
    map.set(raw.name, {
      name: raw.name,
      steps,
      windowMs: clampFlowWindow(raw.windowMs),
      important: raw.important === true,
    });
  }
  return map;
}

function buildEvents(names: string[] | undefined): Set<string> {
  const set = new Set<string>();
  if (!names) return set;
  for (const name of names) {
    if (typeof name === 'string' && BUSINESS_NAME_RE.test(name)) set.add(name);
  }
  return set;
}

function restoreAttempt(): void {
  if (attempt) return;
  const stored = readStoredFlow();
  if (!stored) return;
  const def = registry.get(stored.name);
  if (!def || stored.attemptId === '') {
    writeStoredFlow(null);
    return;
  }
  const furthest = Math.min(Math.max(stored.furthest, 0), def.steps.length - 1);
  attempt = {
    name: def.name,
    steps: def.steps,
    windowMs: def.windowMs,
    important: def.important,
    attemptId: stored.attemptId,
    furthest,
    sentDef: stored.sentDef,
  };
}

function emit(event: WireEvent): void {
  queue?.add(event);
}

function emitFlowStep(
  status: 'entered' | 'completed' | 'failed' | 'abandoned',
  step: string,
  reason?: string,
): void {
  if (!attempt || !queue) return;
  const includeDef = !attempt.sentDef;
  const event: WireEvent = {
    ...envelope(),
    type: 'flow_step',
    flow_id: attempt.attemptId,
    flow_name: attempt.name,
    step,
    status,
    attempt_id: attempt.attemptId,
  };
  const cleaned = validReason(reason);
  if (cleaned) event.reason = cleaned;
  if (includeDef) {
    event.steps = attempt.steps.slice();
    event.window_ms = attempt.windowMs;
    event.important = attempt.important;
    attempt.sentDef = true;
  }
  const flowId = attempt.attemptId;
  persistAttempt();
  emit(event);
  notifyWatch?.('flow_step', { flowId });
}

function beginAttempt(def: FlowDef): void {
  attempt = {
    name: def.name,
    steps: def.steps,
    windowMs: def.windowMs,
    important: def.important,
    attemptId: randomId(),
    furthest: 0,
    sentDef: false,
  };
  persistAttempt();
  const first = def.steps[0];
  if (first) emitFlowStep('entered', first);
}

function startFlow(def: FlowDef): void {
  if (!attempt || attempt.name !== def.name || attempt.furthest > 0) {
    beginAttempt(def);
    return;
  }
  const first = def.steps[0];
  if (first) emitFlowStep('entered', first);
}

function stepFlow(def: FlowDef, step: string): void {
  if (!attempt || attempt.name !== def.name) {
    if (def.steps[0] === step) startFlow(def);
    return;
  }
  const index = def.steps.indexOf(step);
  if (index < 0) return;
  if (index === 0 && attempt.furthest > 0) {
    beginAttempt(def);
    return;
  }
  if (index !== attempt.furthest && index !== attempt.furthest + 1) return;
  attempt.furthest = index;
  emitFlowStep('entered', step);
}

function endFlow(def: FlowDef, status: 'completed' | 'failed' | 'abandoned', reason?: string): void {
  if (!attempt || attempt.name !== def.name) return;
  if (status === 'completed') {
    const last = def.steps.length - 1;
    if (attempt.furthest !== last) return;
    const step = def.steps[last];
    if (!step) return;
    emitFlowStep('completed', step);
    attempt = null;
    persistAttempt();
    return;
  }
  const step = def.steps[attempt.furthest];
  if (!step) return;
  emitFlowStep(status, step, reason);
  attempt = null;
  persistAttempt();
}

function handleNavigate(change: RouteChange): void {
  currentRoute = change.to;
  if (change.kind === 'load') {
    if (loadSeen) return;
    loadSeen = true;
  } else {
    notifyWatch?.('navigation');
    pageSeq = bumpPageSeq();
    resetErrorBudget();
  }
  emit({
    ...envelope(),
    type: 'navigation',
    route: change.to,
    from_route: change.from,
    to_route: change.to,
    kind: change.kind,
  });
}

function onPageHide(): void {
  markPageUnloading();
  runtime && clicksPageHide?.();
  void queue?.flush('unload');
}

function onPageShow(): void {
  clearPageUnloading();
}

function onVisibility(): void {
  if (typeof document !== 'undefined' && document.visibilityState === 'hidden') {
    void queue?.flush('keepalive');
  }
}

let clicksPageHide: (() => void) | null = null;

/**
 * Start observers. In the browser, missing apiKey or ingestionUrl throws.
 * A second call keeps the first configuration. With no window, this returns.
 */
export function init(initOptions: InitOptions): void {
  if (typeof window === 'undefined') return;
  if (initialized) return;

  const apiKey = typeof initOptions?.apiKey === 'string' ? initOptions.apiKey.trim() : '';
  const ingestionUrl = typeof initOptions?.ingestionUrl === 'string' ? initOptions.ingestionUrl.trim() : '';
  if (apiKey === '') throw new Error('HawkEye SDK: apiKey is required');
  if (ingestionUrl === '') throw new Error('HawkEye SDK: ingestionUrl is required');

  options = { ...initOptions, apiKey, ingestionUrl };
  registry = buildRegistry(initOptions.importantFlows);
  allowedEvents = buildEvents(initOptions.events);
  if (attempt && !registry.has(attempt.name)) {
    attempt = null;
    writeStoredFlow(null);
  }
  restoreAttempt();

  const identity = loadIdentity();
  anonymousId = identity.anonymousId;
  sessionId = identity.sessionId;
  pageSeq = identity.pageSeq;
  monoOrigin = typeof performance !== 'undefined' ? performance.now() : Date.now();

  const endpoint = eventsEndpoint(ingestionUrl);
  queue = new EventQueue(endpoint, apiKey);

  let noteLink: (() => void) | null = null;
  const clicks = installClicks({
    envelope,
    emit,
    mono,
    onSameTabLink: () => noteLink?.(),
  });
  notifyWatch = clicks.notify;
  clicksPageHide = clicks.onPageHide;

  const stopErrors = installErrors({
    envelope,
    emit,
    ingestionUrl: endpoint,
  });

  const navigation = installNavigation((change) => {
    handleNavigate(change);
  });
  noteLink = navigation.noteSameTabLink;
  setRouteImpl = navigation.setRoute;
  currentRoute = navigation.current();

  const pageAc = new AbortController();
  window.addEventListener('pagehide', onPageHide, { capture: true, signal: pageAc.signal });
  window.addEventListener('pageshow', onPageShow, { signal: pageAc.signal });
  document.addEventListener('visibilitychange', onVisibility, { signal: pageAc.signal });

  runtime = {
    queue,
    stopClicks: clicks.stop,
    stopNav: navigation.stop,
    stopErrors,
    stopPage: () => pageAc.abort(),
  };
  initialized = true;
}

/** Drop the event unless `name` was listed in `init({ events })`. */
export function track(name: string, props?: Record<string, string | number | boolean>): void {
  if (!initialized || !queue) return;
  if (typeof name !== 'string' || !allowedEvents.has(name)) return;
  const cleaned = sanitizeBusinessProps(props as Record<string, unknown> | undefined);
  if (cleaned == null) return;
  const flowId = attempt?.attemptId ?? '';
  emit({
    ...envelope(),
    type: 'business',
    name,
    props: cleaned,
  });
  notifyWatch?.('business', { flowId });
}

/** Host form result. A missing or invalid formKey drops the event. */
export function formOutcome(input: FormOutcomeInput): void {
  if (!initialized || !queue || !input) return;
  if (typeof input.formKey !== 'string' || !TOKEN_RE.test(input.formKey)) return;
  if (
    input.result !== 'success' &&
    input.result !== 'validation_error' &&
    input.result !== 'submit_error'
  ) {
    return;
  }
  const event: WireEvent = {
    ...envelope(),
    type: 'form_outcome',
    form_key: input.formKey,
    result: input.result,
    field_keys: sanitizeFieldKeys(input.fieldKeys),
  };
  const code = validReason(input.errorCode);
  if (code) event.error_code = code;
  emit(event);
  notifyWatch?.('form_outcome', { formKey: input.formKey });
}

/**
 * Handle for a registered flow. Unknown names, and calls before init,
 * return a handle whose methods do nothing.
 */
export function flow(name: string): FlowHandle {
  if (!initialized) return noopHandle;
  const def = registry.get(name);
  if (!def) return noopHandle;
  return {
    start() {
      startFlow(def);
    },
    step(step: string) {
      stepFlow(def, step);
    },
    complete() {
      endFlow(def, 'completed');
    },
    fail(reason?: string) {
      endFlow(def, 'failed', reason);
    },
    abandon(reason?: string) {
      endFlow(def, 'abandoned', reason);
    },
  };
}

/** Supply a path template. Templates containing `?`, `#`, `@`, or `%` are ignored. */
export function setRoute(template: string): void {
  if (!initialized) return;
  setRouteImpl?.(template);
}

/**
 * Remove listeners, restore fetch and XHR, and flush the queue.
 * Does not clear localStorage.
 */
export async function teardown(): Promise<void> {
  if (!initialized || !runtime) return;
  const current = runtime;
  initialized = false;
  runtime = null;
  notifyWatch = null;
  setRouteImpl = null;
  clicksPageHide = null;
  current.stopPage();
  current.stopClicks();
  current.stopNav();
  current.stopErrors();
  await current.queue.flush('keepalive');
  current.queue.stop();
  queue = null;
}
