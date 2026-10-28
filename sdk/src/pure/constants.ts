/** Detection numbers from the HawkEye detection contract. */

export const REPEATED_CLICK_GAP_MS = 1000;
export const DEAD_CLICK_MS = 2500;
export const LINK_BLUR_MS = 1000;
export const VALIDATION_LOOP_GAP_MS = 120000;
export const FLOW_WINDOW_MS = 1800000;
export const FLOW_WINDOW_MIN_MS = 60000;
export const FLOW_WINDOW_MAX_MS = 86400000;
export const ERROR_BUDGET_LIMIT = 10;
export const ERROR_BUDGET_INTERVAL_MS = 10000;
export const MESSAGE_SLUG_MAX = 120;
export const ARIA_LABEL_MAX = 40;
export const SIGNATURE_MAX = 120;
export const BATCH_SIZE = 10;
export const BATCH_INTERVAL_MS = 5000;
export const MAX_PROPS = 8;
export const MAX_PROP_LENGTH = 64;
export const MAX_FIELD_KEYS = 8;
export const SDK_VERSION = '1.0.0';

/** True only when the gap is strictly under the repeated-click window. */
export function continuesBurst(gapMs: number): boolean {
  return gapMs < REPEATED_CLICK_GAP_MS;
}

/** True when the dead-click deadline has been reached. */
export function isDeadElapsed(elapsedMs: number): boolean {
  return elapsedMs >= DEAD_CLICK_MS;
}
