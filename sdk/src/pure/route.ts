const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const HEX_RE = /^[0-9a-f]{8,}$/;
const DIGITS_RE = /^\d+$/;

/**
 * Lowercase a pathname, drop search and hash, and replace id-like segments with `:id`.
 * An empty path becomes `/`.
 */
export function normalizePath(pathname: string): string {
  let path = typeof pathname === 'string' ? pathname : '';
  const hash = path.indexOf('#');
  if (hash >= 0) path = path.slice(0, hash);
  const query = path.indexOf('?');
  if (query >= 0) path = path.slice(0, query);
  path = path.toLowerCase();
  if (path === '') return '/';
  if (!path.startsWith('/')) path = `/${path}`;

  const segments = path.split('/');
  const mapped = segments.map((segment, index) => {
    if (index === 0 || segment === '') return segment;
    if (DIGITS_RE.test(segment) || UUID_RE.test(segment) || HEX_RE.test(segment)) return ':id';
    return segment;
  });
  const out = mapped.join('/');
  return out === '' ? '/' : out;
}

/** False when a host template contains `?`, `#`, `@`, or `%`. */
export function acceptRouteTemplate(template: string): boolean {
  return typeof template === 'string' && !/[?#@%]/.test(template);
}
