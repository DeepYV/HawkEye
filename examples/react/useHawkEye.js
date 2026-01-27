import { useEffect } from 'react';
import { initFrustrationObserver, teardown } from '@hawkeye/observer-sdk';

/**
 * Custom React hook for HawkEye integration
 *
 * @param {Object} config - SDK configuration
 * @returns {void}
 */
export function useHawkEye(config) {
  useEffect(() => {
    initFrustrationObserver(config);

    return () => {
      teardown();
    };
  }, [config.apiKey, config.ingestionUrl]);
}

/**
 * Example usage:
 *
 * function App() {
 *   useHawkEye({
 *     apiKey: 'your-api-key',
 *     ingestionUrl: 'https://api.hawkeye.example.com',
 *     enableDebug: true,
 *     environment: 'production',
 *   });
 *
 *   return <div>Your app</div>;
 * }
 */
