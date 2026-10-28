import { SIGNATURE_MAX } from './constants.js';
import { TOKEN_RE } from './grammar.js';

const CONTROL_ROLES = new Set([
  'button',
  'link',
  'tab',
  'menuitem',
  'checkbox',
  'radio',
  'switch',
]);

const TEXT_INPUT_TYPES = new Set([
  'text',
  'search',
  'email',
  'password',
  'url',
  'tel',
  'number',
  'date',
  'datetime-local',
  'time',
  'month',
  'week',
  'color',
  'file',
]);

const SUBMIT_INPUT_TYPES = new Set(['button', 'submit', 'reset', 'image']);

/** Stepper words compared to the whole accepted label, case-insensitive. */
const STEPPER_LABELS = new Set([
  'next',
  'previous',
  'prev',
  '+',
  '-',
  '−',
  '–',
  '→',
  '←',
  '›',
  '‹',
  '»',
  '«',
  '▶',
  '◀',
  '❯',
  '❮',
]);

const LABEL_RE = /^[A-Za-z0-9][A-Za-z0-9 ._+-]{0,39}$/;
const ID_RE = /^[A-Za-z][A-Za-z0-9-]{0,31}$/;

export type SignatureParts = {
  ignored?: boolean;
  hawkeye?: string;
  tag?: string;
  role?: string;
  ariaLabel?: string;
  id?: string;
  inputType?: string;
  contentEditable?: boolean;
  /** Present when the element has an href attribute. The value is never read. */
  hasHref?: boolean;
};

function hasFourDigitRun(value: string): boolean {
  return /\d{4}/.test(value);
}

function isTextField(input: SignatureParts, tag: string): boolean {
  if (input.contentEditable) return true;
  if (tag === 'textarea') return true;
  if (tag === 'input') {
    const type = (input.inputType ?? '').toLowerCase();
    if (type === '' || TEXT_INPUT_TYPES.has(type)) return true;
  }
  return false;
}

function resolveRole(input: SignatureParts, tag: string): string | null {
  const explicit = (input.role ?? '').toLowerCase();
  if (CONTROL_ROLES.has(explicit)) return explicit;
  if (tag === 'button') return 'button';
  if (tag === 'a' && input.hasHref) return 'link';
  if (tag === 'input') {
    const type = (input.inputType ?? '').toLowerCase();
    if (type === 'checkbox') return 'checkbox';
    if (type === 'radio') return 'radio';
    if (SUBMIT_INPUT_TYPES.has(type)) return 'button';
  }
  return null;
}

/** `stepper` excludes the click. A string is the label slug. Null falls through. */
function labelSlug(raw: string | undefined): 'stepper' | string | null {
  if (!raw) return null;
  const collapsed = raw.trim().replace(/\s+/g, ' ');
  if (collapsed === '') return null;
  if (!LABEL_RE.test(collapsed) || hasFourDigitRun(collapsed)) return null;
  if (STEPPER_LABELS.has(collapsed.toLowerCase())) return 'stepper';
  const slug = collapsed
    .toLowerCase()
    .replace(/ /g, '-')
    .replace(/\+/g, '-')
    .replace(/[^a-z0-9._-]/g, '');
  return slug === '' ? null : slug;
}

function usableId(raw: string | undefined): string | null {
  if (!raw) return null;
  if (!ID_RE.test(raw) || hasFourDigitRun(raw)) return null;
  return raw.toLowerCase();
}

function capSignature(signature: string): string {
  return signature.length > SIGNATURE_MAX ? signature.slice(0, SIGNATURE_MAX) : signature;
}

/**
 * Build a click signature from attributes the caller already read.
 * This module does not touch the DOM.
 */
export function signatureFromParts(input: SignatureParts): { signature: string } | null {
  if (!input || input.ignored) return null;

  const tag = (input.tag ?? '').toLowerCase();
  if (isTextField(input, tag)) return null;

  const token = input.hawkeye?.trim();
  if (token && TOKEN_RE.test(token)) {
    return { signature: capSignature(`he:${token}`) };
  }

  const role = resolveRole(input, tag);
  if (!role || tag === '') return null;

  const label = labelSlug(input.ariaLabel);
  if (label === 'stepper') return null;
  if (typeof label === 'string') {
    return { signature: capSignature(`ctrl:${tag}:${role}:${label}`) };
  }

  const id = usableId(input.id);
  if (id) return { signature: capSignature(`ctrl:${tag}:${role}#${id}`) };
  return { signature: capSignature(`ctrl:${tag}:${role}`) };
}
