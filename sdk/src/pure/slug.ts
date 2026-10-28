import { MESSAGE_SLUG_MAX } from './constants.js';

/**
 * Redact a JS message: quoted spans and long hex runs become `*`,
 * digit runs become `#`, whitespace collapses, and the result is cut at 120.
 * Empty becomes `-`.
 */
export function messageSlug(message: string): string {
  let text = typeof message === 'string' ? message : '';
  text = text.replace(/"(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*'/g, '*');
  text = text.replace(/[0-9a-fA-F]{8,}/g, '*');
  text = text.replace(/\d+/g, '#');
  text = text.replace(/\s+/g, ' ').trim();
  if (text.length > MESSAGE_SLUG_MAX) text = text.slice(0, MESSAGE_SLUG_MAX);
  return text === '' ? '-' : text;
}

/** Letters only, at most 40, default `Error`. */
export function jsName(name: unknown): string {
  const letters = String(name ?? '').replace(/[^A-Za-z]/g, '').slice(0, 40);
  return letters || 'Error';
}

/** File basename, at most 40, or `-`. */
export function fileBase(filename: unknown): string {
  if (typeof filename !== 'string' || filename === '') return '-';
  const noQuery = filename.split('?')[0]?.split('#')[0] ?? '';
  const base = noQuery.split(/[/\\]/).pop() ?? '';
  if (base === '') return '-';
  return base.slice(0, 40);
}

/** Function name, at most 40, or `-`. */
export function capFunction(name: string | undefined): string {
  if (!name) return '-';
  const trimmed = name.trim();
  if (trimmed === '' || trimmed === 'anonymous' || trimmed === '<anonymous>') return '-';
  return trimmed.slice(0, 40);
}

/** First meaningful function name from a stack, without keeping the stack. */
export function functionFromStack(stack: string | undefined): string {
  if (!stack) return '-';
  const lines = stack.split('\n');
  for (const line of lines) {
    const trimmed = line.trim();
    if (trimmed === '') continue;
    let fn = '';
    const at = trimmed.match(/^at\s+(?:async\s+)?(.+)$/);
    if (at) {
      const rest = at[1] ?? '';
      const named = rest.match(/^([^\s(/]+)\s*(?:\(|$)/);
      if (named && !named[1]?.includes('/') && !named[1]?.includes('\\')) {
        fn = named[1] ?? '';
      }
    } else {
      const firefox = trimmed.match(/^([^@\s]+)@/);
      if (firefox) fn = firefox[1] ?? '';
    }
    if (fn === '') continue;
    const leaf = fn.split('.').pop() ?? fn;
    const capped = capFunction(leaf);
    if (capped !== '-') return capped;
  }
  return '-';
}
