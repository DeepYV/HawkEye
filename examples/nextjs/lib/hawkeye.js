/**
 * Client-side HawkEye helpers for Next.js.
 * Call these from useEffect or event handlers, not from a server component.
 */

export async function initHawkEye(config) {
  if (typeof window === 'undefined') return;

  const { init } = await import('@hawkeye/observer-sdk');
  init({
    apiKey: config.apiKey || process.env.NEXT_PUBLIC_HAWKEYE_API_KEY,
    ingestionUrl: config.ingestionUrl || process.env.NEXT_PUBLIC_HAWKEYE_INGESTION_URL,
    appVersion: config.appVersion,
    appBuild: config.appBuild,
    appNamespace: config.appNamespace,
    releaseId: config.releaseId,
    environment: config.environment || process.env.NODE_ENV,
    events: config.events,
    importantFlows: config.importantFlows,
  });
}

export async function trackEvent(name, props) {
  if (typeof window === 'undefined') return;
  const { track } = await import('@hawkeye/observer-sdk');
  track(name, props);
}

export async function reportForm(input) {
  if (typeof window === 'undefined') return;
  const { formOutcome } = await import('@hawkeye/observer-sdk');
  formOutcome(input);
}

/** Next.js `[id]` and `[...slug]` segments become HawkEye `:id` templates. */
export function hawkeyeTemplate(pathname) {
  return String(pathname || '/')
    .replace(/\[\.\.\.[^\]]+\]/g, ':id')
    .replace(/\[[^\]]+\]/g, ':id');
}
