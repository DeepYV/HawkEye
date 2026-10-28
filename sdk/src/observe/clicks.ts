import { DEAD_CLICK_MS, LINK_BLUR_MS, isDeadElapsed } from '../pure/constants.js';
import { TOKEN_RE } from '../pure/grammar.js';
import { signatureFromParts, type SignatureParts } from '../pure/signature.js';
import type { WireEvent } from '../types/wire.js';

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

export type WatchNotice = 'navigation' | 'form_outcome' | 'flow_step' | 'business';

type EnvelopeFn = () => Omit<WireEvent, 'type'>;

type Watch = {
  key: string;
  clickId: string;
  signature: string;
  formKey: string;
  flowId: string;
  flowName: string;
  pageSeq: number;
  route: string;
  sessionId: string;
  startedMono: number;
  linkBlur: boolean;
  done: boolean;
  observer: MutationObserver | null;
  timer: ReturnType<typeof setTimeout> | null;
  snap: { disabled: boolean; busy: string | null; expanded: string | null };
};

function tagName(el: Element): string {
  return el.tagName.toLowerCase();
}

function isTextField(el: Element): boolean {
  if ('isContentEditable' in el && (el as HTMLElement).isContentEditable) return true;
  const tag = tagName(el);
  if (tag === 'textarea') return true;
  if (tag === 'input') {
    const type = (el.getAttribute('type') || '').toLowerCase();
    if (type === '' || TEXT_INPUT_TYPES.has(type)) return true;
  }
  return false;
}

function isControl(el: Element): boolean {
  const tag = tagName(el);
  if (tag === 'button') return true;
  if (tag === 'a' && el.hasAttribute('href')) return true;
  if (tag === 'input') {
    const type = (el.getAttribute('type') || '').toLowerCase();
    if (SUBMIT_INPUT_TYPES.has(type) || type === 'checkbox' || type === 'radio') return true;
  }
  const role = (el.getAttribute('role') || '').toLowerCase();
  return CONTROL_ROLES.has(role);
}

function ignored(start: Element): boolean {
  let node: Element | null = start;
  while (node) {
    if (node.hasAttribute('data-hawkeye-ignore')) return true;
    if (tagName(node) === 'body') break;
    node = node.parentElement;
  }
  return false;
}

function nearestToken(start: Element): { token: string; el: Element } | null {
  let node: Element | null = start;
  while (node) {
    if (node.hasAttribute('data-hawkeye')) {
      const raw = node.getAttribute('data-hawkeye') || '';
      if (TOKEN_RE.test(raw)) return { token: raw, el: node };
    }
    if (tagName(node) === 'body') break;
    node = node.parentElement;
  }
  return null;
}

function nearestControl(start: Element): Element | null {
  let node: Element | null = start;
  while (node) {
    if (tagName(node) === 'label' && 'control' in node) {
      const control = (node as HTMLLabelElement).control;
      if (control && isControl(control) && !isTextField(control)) return control;
    }
    if (isControl(node) && !isTextField(node)) return node;
    if (tagName(node) === 'body') break;
    node = node.parentElement;
  }
  return null;
}

function partsFrom(el: Element, extra: { hawkeye?: string }): SignatureParts {
  const tag = tagName(el);
  return {
    hawkeye: extra.hawkeye,
    tag,
    role: el.getAttribute('role') || undefined,
    ariaLabel: el.getAttribute('aria-label') || undefined,
    id: el.getAttribute('id') || undefined,
    inputType: tag === 'input' ? el.getAttribute('type') || '' : undefined,
    contentEditable: 'isContentEditable' in el && (el as HTMLElement).isContentEditable === true,
    hasHref: el.hasAttribute('href'),
  };
}

function formKeyFor(nodes: Element[]): string {
  for (const el of nodes) {
    if (!('form' in el)) continue;
    const form = (el as HTMLButtonElement).form;
    if (!(form instanceof HTMLFormElement)) continue;
    const token = form.getAttribute('data-hawkeye');
    if (token && TOKEN_RE.test(token)) return token;
    const name = form.getAttribute('name') || '';
    if (TOKEN_RE.test(name)) return name;
    return '';
  }
  return '';
}

function isDisabled(el: Element): boolean {
  if (el.hasAttribute('disabled')) return true;
  return 'disabled' in el && (el as HTMLButtonElement).disabled === true;
}

function sameTabLink(el: Element): boolean {
  if (tagName(el) !== 'a' || !el.hasAttribute('href')) return false;
  if (!el.hasAttribute('target')) return true;
  const target = (el.getAttribute('target') || '').toLowerCase();
  return target === '' || target === '_self';
}

function linkBlurEligible(el: Element): boolean {
  const tag = tagName(el);
  const role = (el.getAttribute('role') || '').toLowerCase();
  if (tag !== 'a' && role !== 'link') return false;
  if (!el.hasAttribute('target')) return false;
  const target = (el.getAttribute('target') || '').toLowerCase();
  return target !== '' && target !== '_self';
}

function clickTarget(event: MouseEvent): Element | null {
  const target = event.target;
  if (target instanceof Element) return target;
  if (target instanceof Node && target.parentElement) return target.parentElement;
  return null;
}

/**
 * Capture-phase clicks and one dead-click watch per (session, page, signature).
 * The attribute observer is attached to the signature element only.
 */
export function installClicks(deps: {
  envelope: EnvelopeFn;
  emit: (event: WireEvent) => void;
  mono: () => number;
  onSameTabLink: () => void;
}): {
  stop: () => void;
  notify: (kind: WatchNotice, data?: { formKey?: string; flowId?: string }) => void;
  onPageHide: () => void;
} {
  const watches = new Map<string, Watch>();
  const ac = new AbortController();

  const finish = (watch: Watch, result: 'responded' | 'dead', reason: string, elapsed: number): void => {
    if (watch.done) return;
    watch.done = true;
    if (watch.timer !== null) clearTimeout(watch.timer);
    watch.observer?.disconnect();
    watches.delete(watch.key);
    const base = deps.envelope();
    deps.emit({
      ...base,
      type: 'click_outcome',
      page_seq: watch.pageSeq,
      route: watch.route,
      flow_id: watch.flowId,
      flow_name: watch.flowName,
      session_id: watch.sessionId,
      click_id: watch.clickId,
      element_signature: watch.signature,
      result,
      reason,
      elapsed_ms: Math.max(0, Math.round(elapsed)),
    });
  };

  const arm = (watch: Watch): void => {
    const tick = (): void => {
      if (watch.done) return;
      const elapsed = deps.mono() - watch.startedMono;
      if (isDeadElapsed(elapsed)) {
        finish(watch, 'dead', 'timeout', elapsed);
        return;
      }
      watch.timer = setTimeout(tick, Math.max(0, DEAD_CLICK_MS - elapsed));
    };
    watch.timer = setTimeout(tick, DEAD_CLICK_MS);
  };

  const watchAttributes = (watch: Watch, el: Element): void => {
    const snap = watch.snap;
    const observer = new MutationObserver(() => {
      if (watch.done) return;
      const elapsed = deps.mono() - watch.startedMono;
      if (isDeadElapsed(elapsed)) return;
      const disabled = isDisabled(el);
      const busy = el.getAttribute('aria-busy');
      const expanded = el.getAttribute('aria-expanded');
      if (!snap.disabled && disabled) {
        finish(watch, 'responded', 'disabled', elapsed);
        return;
      }
      if (busy === 'true' && snap.busy !== 'true') {
        finish(watch, 'responded', 'aria_busy', elapsed);
        return;
      }
      if (snap.busy === 'true' && busy === 'false') {
        finish(watch, 'responded', 'aria_busy', elapsed);
        return;
      }
      if (expanded !== snap.expanded) {
        finish(watch, 'responded', 'aria_expanded', elapsed);
      }
    });
    observer.observe(el, {
      attributes: true,
      attributeFilter: ['disabled', 'aria-busy', 'aria-expanded'],
      subtree: false,
    });
    watch.observer = observer;
  };

  const openWatch = (
    el: Element,
    signature: string,
    formKey: string,
    clickId: string,
    base: Omit<WireEvent, 'type'>,
    startedMono: number,
  ): void => {
    const key = `${base.session_id}\n${base.page_seq}\n${signature}`;
    if (watches.has(key)) return;
    const watch: Watch = {
      key,
      clickId,
      signature,
      formKey,
      flowId: base.flow_id,
      flowName: base.flow_name,
      pageSeq: base.page_seq,
      route: base.route,
      sessionId: base.session_id,
      startedMono,
      linkBlur: linkBlurEligible(el),
      done: false,
      observer: null,
      timer: null,
      snap: {
        disabled: isDisabled(el),
        busy: el.getAttribute('aria-busy'),
        expanded: el.getAttribute('aria-expanded'),
      },
    };
    watches.set(key, watch);
    watchAttributes(watch, el);
    arm(watch);
  };

  const onClick = (event: MouseEvent): void => {
    if (event.button !== 0) return;
    if (event.ctrlKey || event.metaKey || event.altKey || event.shiftKey) return;
    const target = clickTarget(event);
    if (!target) return;
    if (ignored(target)) return;
    if (isTextField(target)) return;

    const token = nearestToken(target);
    let signatureEl: Element;
    let built: { signature: string } | null;
    if (token) {
      signatureEl = token.el;
      built = signatureFromParts(partsFrom(signatureEl, { hawkeye: token.token }));
    } else {
      const control = nearestControl(target);
      if (!control) return;
      signatureEl = control;
      built = signatureFromParts(partsFrom(control, {}));
    }
    if (!built) return;

    if (sameTabLink(signatureEl) || sameTabLink(target)) deps.onSameTabLink();

    const startedMono = deps.mono();
    const base = deps.envelope();
    const formKey = formKeyFor([signatureEl, target]);
    deps.emit({
      ...base,
      type: 'click',
      element_signature: built.signature,
      form_key: formKey,
    });
    openWatch(signatureEl, built.signature, formKey, base.event_id, base, startedMono);
  };

  const onBlur = (): void => {
    const now = deps.mono();
    for (const watch of [...watches.values()]) {
      if (watch.done || !watch.linkBlur) continue;
      const elapsed = now - watch.startedMono;
      if (elapsed < LINK_BLUR_MS && !isDeadElapsed(elapsed)) {
        finish(watch, 'responded', 'link_blur', elapsed);
      }
    }
  };

  document.addEventListener('click', onClick, { capture: true, signal: ac.signal });
  window.addEventListener('blur', onBlur, { signal: ac.signal });

  return {
    stop: () => {
      ac.abort();
      for (const watch of watches.values()) {
        if (watch.timer !== null) clearTimeout(watch.timer);
        watch.observer?.disconnect();
        watch.done = true;
      }
      watches.clear();
    },
    notify: (kind, data) => {
      const now = deps.mono();
      for (const watch of [...watches.values()]) {
        if (watch.done) continue;
        const elapsed = now - watch.startedMono;
        if (isDeadElapsed(elapsed)) continue;
        if (kind === 'navigation') {
          finish(watch, 'responded', 'navigation', elapsed);
          continue;
        }
        if (kind === 'form_outcome') {
          const formKey = data?.formKey ?? '';
          if (formKey !== '' && formKey === watch.formKey) {
            finish(watch, 'responded', 'form_outcome', elapsed);
          }
          continue;
        }
        const flowId = data?.flowId ?? '';
        if (flowId !== '' && flowId === watch.flowId) {
          finish(watch, 'responded', kind, elapsed);
        }
      }
    },
    onPageHide: () => {
      const now = deps.mono();
      for (const watch of [...watches.values()]) {
        if (watch.done) continue;
        const elapsed = now - watch.startedMono;
        if (!isDeadElapsed(elapsed)) finish(watch, 'responded', 'unload', elapsed);
      }
    },
  };
}
