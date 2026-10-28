import {
  FLOW_WINDOW_MAX_MS,
  FLOW_WINDOW_MIN_MS,
  FLOW_WINDOW_MS,
  MAX_FIELD_KEYS,
  MAX_PROP_LENGTH,
  MAX_PROPS,
} from './constants.js';

export const FLOW_NAME_RE = /^[a-z][a-z0-9_-]{0,63}$/;
export const BUSINESS_NAME_RE = /^[a-z][a-z0-9_]{0,63}$/;
export const TOKEN_RE = /^[a-z0-9][a-z0-9-]{0,63}$/;
export const FIELD_KEY_RE = /^[a-z][a-z0-9_-]{0,31}$/;
export const REASON_RE = /^[a-z0-9_-]{1,32}$/;

const PROP_DENYLIST = new Set([
  'password',
  'email',
  'token',
  'secret',
  'name',
  'query',
  'html',
  'value',
  'text',
  'dom',
  'href',
  'selector',
  'message',
  'input',
  'content',
]);

/** False when the key is on the business-prop denylist. */
export function businessPropAllowed(key: string): boolean {
  return typeof key === 'string' && !PROP_DENYLIST.has(key.toLowerCase());
}

export function clampFlowWindow(windowMs: number | undefined): number {
  if (typeof windowMs !== 'number' || !Number.isFinite(windowMs)) return FLOW_WINDOW_MS;
  if (windowMs < FLOW_WINDOW_MIN_MS || windowMs > FLOW_WINDOW_MAX_MS) return FLOW_WINDOW_MS;
  return windowMs;
}

export function validReason(reason: string | undefined): string | undefined {
  if (!reason || !REASON_RE.test(reason)) return undefined;
  return reason;
}

/** Keep at most 8 grammatical field keys. Extra and invalid keys are dropped. */
export function sanitizeFieldKeys(keys: readonly unknown[] | undefined): string[] {
  if (!keys) return [];
  const out: string[] = [];
  for (const key of keys) {
    if (out.length >= MAX_FIELD_KEYS) break;
    if (typeof key === 'string' && FIELD_KEY_RE.test(key)) out.push(key);
  }
  return out;
}

/**
 * Returns cleaned props, or null when the event must be dropped
 * (too many props, a denylisted key, or a value that is not a short scalar).
 */
export function sanitizeBusinessProps(
  props: Record<string, unknown> | undefined,
): Record<string, string | number | boolean> | null {
  if (props == null) return {};
  if (typeof props !== 'object' || Array.isArray(props)) return null;
  const keys = Object.keys(props);
  const out: Record<string, string | number | boolean> = {};
  for (const key of keys) {
    if (!businessPropAllowed(key)) return null;
    const value = props[key];
    let kept: string | number | boolean;
    if (typeof value === 'string') {
      if (value.length > MAX_PROP_LENGTH) return null;
      kept = value;
    } else if (typeof value === 'boolean') {
      kept = value;
    } else if (typeof value === 'number' && Number.isFinite(value)) {
      kept = value;
    } else {
      return null;
    }
    if (Object.keys(out).length < MAX_PROPS) out[key] = kept;
  }
  return out;
}
