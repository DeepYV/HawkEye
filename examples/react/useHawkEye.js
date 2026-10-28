import { useEffect } from 'react';
import { init, teardown } from '@hawkeye/observer-sdk';

/**
 * Start HawkEye for this mount and stop it on cleanup.
 * Pass the same object shape as init().
 */
export function useHawkEye(config) {
  useEffect(() => {
    init(config);
    return () => {
      teardown();
    };
  }, [config.apiKey, config.ingestionUrl]);
}
