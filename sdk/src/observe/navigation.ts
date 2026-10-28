import { acceptRouteTemplate, normalizePath } from '../pure/route.js';
import type { NavKind } from '../types/wire.js';

export type RouteChange = {
  from: string;
  to: string;
  kind: NavKind;
};

type HistoryFns = {
  pushState: History['pushState'];
  replaceState: History['replaceState'];
};

/**
 * History and setRoute. The host template wins until the next history change.
 * Query-only and hash-only updates emit nothing.
 */
export function installNavigation(onChange: (change: RouteChange) => void): {
  stop: () => void;
  setRoute: (template: string) => void;
  current: () => string;
  noteSameTabLink: () => void;
} {
  let route = normalizePath(window.location.pathname || '/');
  let rawPath = (window.location.pathname || '/').toLowerCase();
  let linkIntent = false;
  let linkTimer: ReturnType<typeof setTimeout> | null = null;
  const original: HistoryFns = {
    pushState: history.pushState.bind(history),
    replaceState: history.replaceState.bind(history),
  };

  const consumeLink = (): boolean => {
    if (!linkIntent) return false;
    linkIntent = false;
    if (linkTimer !== null) {
      clearTimeout(linkTimer);
      linkTimer = null;
    }
    return true;
  };

  const emitIfChanged = (kind: NavKind, nextRoute: string, nextRaw: string): void => {
    const rawChanged = nextRaw !== rawPath;
    const routeChanged = nextRoute !== route;
    if (!rawChanged && !routeChanged) return;
    const from = route;
    route = nextRoute;
    rawPath = nextRaw;
    onChange({ from, to: nextRoute, kind });
  };

  const fromLocation = (kind: Exclude<NavKind, 'load'>): void => {
    const nextRaw = (window.location.pathname || '/').toLowerCase();
    const nextRoute = normalizePath(nextRaw);
    const actual = consumeLink() ? 'link' : kind;
    emitIfChanged(actual, nextRoute, nextRaw);
  };

  const onPop = (): void => {
    fromLocation('pop');
  };

  history.pushState = function pushState(this: History, ...args: Parameters<History['pushState']>) {
    const result = original.pushState.apply(this, args);
    fromLocation('push');
    return result;
  };

  history.replaceState = function replaceState(
    this: History,
    ...args: Parameters<History['replaceState']>
  ) {
    const result = original.replaceState.apply(this, args);
    fromLocation('replace');
    return result;
  };

  window.addEventListener('popstate', onPop);

  // First load has no previous template in this document.
  onChange({ from: '', to: route, kind: 'load' });

  return {
    current: () => route,
    noteSameTabLink: () => {
      linkIntent = true;
      if (linkTimer !== null) clearTimeout(linkTimer);
      linkTimer = setTimeout(() => {
        linkIntent = false;
        linkTimer = null;
      }, 0);
    },
    setRoute: (template: string) => {
      if (typeof template !== 'string') return;
      const trimmed = template.trim();
      if (trimmed === '' || !acceptRouteTemplate(trimmed)) return;
      const next = normalizePath(trimmed);
      const kind: NavKind = consumeLink() ? 'link' : 'push';
      if (next === route) return;
      const from = route;
      route = next;
      onChange({ from, to: next, kind });
    },
    stop: () => {
      history.pushState = original.pushState;
      history.replaceState = original.replaceState;
      window.removeEventListener('popstate', onPop);
      if (linkTimer !== null) clearTimeout(linkTimer);
    },
  };
}
