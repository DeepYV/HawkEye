/**
 * HawkEye SDK Wrapper for Next.js
 * Provides client-side only initialization
 */

let isInitialized = false;

export async function initHawkEye(config) {
  // Only initialize on client side
  if (typeof window === 'undefined') {
    return;
  }

  // Prevent double initialization
  if (isInitialized) {
    return;
  }

  const { initFrustrationObserver } = await import('@hawkeye/observer-sdk');

  initFrustrationObserver({
    apiKey: config.apiKey || process.env.NEXT_PUBLIC_HAWKEYE_API_KEY,
    ingestionUrl: config.ingestionUrl || process.env.NEXT_PUBLIC_HAWKEYE_INGESTION_URL,
    enableDebug: config.enableDebug ?? process.env.NODE_ENV === 'development',
    environment: config.environment || process.env.NODE_ENV,
    appId: config.appId,
    batchSize: config.batchSize || 10,
    batchInterval: config.batchInterval || 5000,
  });

  isInitialized = true;
}

export async function trackEvent(eventType, target, metadata) {
  if (typeof window === 'undefined') {
    return;
  }

  const { captureEvent } = await import('@hawkeye/observer-sdk');
  captureEvent(eventType, target, metadata);
}
