let saved: typeof fetch | null = null;

/** Remember the page's fetch before the SDK patches it. */
export function rememberFetch(fn: typeof fetch): void {
  saved = fn;
}

export function getOriginalFetch(): typeof fetch {
  if (saved) return saved;
  if (typeof globalThis.fetch === 'function') return globalThis.fetch.bind(globalThis);
  return (() => Promise.reject(new Error('HawkEye SDK: fetch is unavailable'))) as typeof fetch;
}
