import { takeErrorBudget, type ErrorBudgetState } from '../pure/budget.js';
import { normalizePath } from '../pure/route.js';
import { fileBase, functionFromStack, jsName, messageSlug } from '../pure/slug.js';
import type { WireEvent } from '../types/wire.js';
import { rememberFetch } from '../transport/fetch.js';

const EXTENSION_PREFIXES = [
  'chrome-extension://',
  'moz-extension://',
  'safari-web-extension://',
  'safari-extension://',
];

type XhrMeta = { method: string; url: string };
const xhrMeta = new WeakMap<XMLHttpRequest, XhrMeta>();

let unloading = false;
let budget: ErrorBudgetState = {};
let ingestUrl = '';

export function markPageUnloading(): void {
  unloading = true;
}

export function clearPageUnloading(): void {
  unloading = false;
}

export function resetErrorBudget(): void {
  budget = {};
}

function isExtension(filename: string): boolean {
  return EXTENSION_PREFIXES.some((prefix) => filename.includes(prefix));
}

function isSdkMessage(message: string): boolean {
  const stripped = message.replace(/^Uncaught\s+/, '');
  return message.startsWith('HawkEye SDK:') || stripped.startsWith('HawkEye SDK:');
}

function requestUrl(raw: string): URL | null {
  try {
    return new URL(raw, window.location.href);
  } catch {
    return null;
  }
}

function isIngest(raw: string): boolean {
  const url = requestUrl(raw);
  if (!url || ingestUrl === '') return false;
  try {
    const target = new URL(ingestUrl);
    const path = url.pathname.replace(/\/+$/, '') || '/';
    const targetPath = target.pathname.replace(/\/+$/, '') || '/';
    return url.origin === target.origin && path === targetPath;
  } catch {
    return false;
  }
}

function httpPath(raw: string): string {
  const url = requestUrl(raw);
  if (!url) return normalizePath(raw);
  return normalizePath(url.pathname);
}

export function installErrors(deps: {
  envelope: () => Omit<WireEvent, 'type'>;
  emit: (event: WireEvent) => void;
  ingestionUrl: string;
}): () => void {
  ingestUrl = deps.ingestionUrl;
  unloading = false;
  budget = {};

  const originalFetch = window.fetch;
  rememberFetch(originalFetch.bind(window));

  const emitHttp = (rawUrl: string, method: string, status: number): void => {
    if (status !== 0 && status < 400) return;
    if (status === 0 && unloading) return;
    if (isIngest(rawUrl)) return;
    const path = httpPath(rawUrl);
    const upper = (method || 'GET').toUpperCase();
    const signature = `http:${upper}:${status}:${path}`;
    const decision = takeErrorBudget(budget, signature, Date.now());
    if (!decision.send) return;
    const base = deps.envelope();
    deps.emit({
      ...base,
      type: 'error',
      kind: 'http',
      method: upper,
      status,
      path,
      truncated: decision.truncated,
    });
  };

  const emitJs = (input: {
    name: unknown;
    message: string;
    filename: string;
    stack?: string;
    line: number;
    column: number;
  }): void => {
    if (isSdkMessage(input.message)) return;
    if (input.filename && isExtension(input.filename)) return;
    const name = jsName(input.name);
    const slug = messageSlug(input.message);
    const file = fileBase(input.filename);
    const fn = functionFromStack(input.stack);
    const signature = `js:${name}:${slug}:${file}:${fn}`;
    const decision = takeErrorBudget(budget, signature, Date.now());
    if (!decision.send) return;
    const base = deps.envelope();
    deps.emit({
      ...base,
      type: 'error',
      kind: 'js',
      name,
      message_slug: slug,
      file,
      function: fn,
      line: input.line,
      column: input.column,
      truncated: decision.truncated,
    });
  };

  const patchedFetch: typeof fetch = function patched(input, init) {
    const rawUrl =
      typeof input === 'string'
        ? input
        : input instanceof URL
          ? input.href
          : input.url;
    const method = (
      init?.method ||
      (!(typeof input === 'string') && !(input instanceof URL) ? input.method : '') ||
      'GET'
    ).toUpperCase();
    return Promise.resolve(originalFetch.call(window, input, init)).then(
      (response) => {
        emitHttp(rawUrl, method, response.status);
        return response;
      },
      (error: unknown) => {
        emitHttp(rawUrl, method, 0);
        throw error;
      },
    );
  };

  window.fetch = patchedFetch;

  const originalOpen = XMLHttpRequest.prototype.open;
  const originalSend = XMLHttpRequest.prototype.send;

  XMLHttpRequest.prototype.open = function open(
    this: XMLHttpRequest,
    method: string,
    url: string | URL,
    async?: boolean,
    username?: string | null,
    password?: string | null,
  ): void {
    xhrMeta.set(this, { method: String(method || 'GET').toUpperCase(), url: String(url) });
    return originalOpen.call(this, method, url, async ?? true, username, password);
  };

  XMLHttpRequest.prototype.send = function send(
    this: XMLHttpRequest,
    body?: Document | XMLHttpRequestBodyInit | null,
  ): void {
    const meta = xhrMeta.get(this);
    this.addEventListener('loadend', () => {
      if (!meta) return;
      emitHttp(meta.url, meta.method, this.status);
    });
    return originalSend.call(this, body);
  };

  const onError = (event: ErrorEvent): void => {
    const error = event.error as { name?: string; message?: string; stack?: string } | null;
    emitJs({
      name: error?.name || 'Error',
      message: event.message || error?.message || '',
      filename: event.filename || '',
      stack: error?.stack,
      line: typeof event.lineno === 'number' ? event.lineno : 0,
      column: typeof event.colno === 'number' ? event.colno : 0,
    });
  };

  const onRejection = (event: PromiseRejectionEvent): void => {
    const reason: unknown = event.reason;
    if (typeof reason === 'string') {
      emitJs({ name: 'Error', message: reason, filename: '', line: 0, column: 0 });
      return;
    }
    if (reason instanceof Error) {
      emitJs({
        name: reason.name,
        message: reason.message,
        filename: '',
        stack: reason.stack,
        line: 0,
        column: 0,
      });
      return;
    }
    emitJs({ name: 'Error', message: 'unhandled rejection', filename: '', line: 0, column: 0 });
  };

  window.addEventListener('error', onError);
  window.addEventListener('unhandledrejection', onRejection);

  return () => {
    window.fetch = originalFetch;
    rememberFetch(originalFetch.bind(window));
    XMLHttpRequest.prototype.open = originalOpen;
    XMLHttpRequest.prototype.send = originalSend;
    window.removeEventListener('error', onError);
    window.removeEventListener('unhandledrejection', onRejection);
    unloading = false;
    budget = {};
  };
}
