import { ERROR_BUDGET_INTERVAL_MS, ERROR_BUDGET_LIMIT } from './constants.js';

export type ErrorBudgetSlot = {
  sent: number;
  lastSentMs: number;
  sawDrop: boolean;
};

/** Mutable per-signature counters. One object per page. */
export type ErrorBudgetState = Record<string, ErrorBudgetSlot>;

/**
 * First 10 errors of a signature are sent. The 10th sets `truncated`.
 * Further events inside 10000 ms are dropped. The next event sent after a drop
 * sets `truncated` again.
 */
export function takeErrorBudget(
  state: ErrorBudgetState,
  signature: string,
  nowMs: number,
): { send: boolean; truncated: boolean } {
  let slot = state[signature];
  if (!slot) {
    slot = { sent: 0, lastSentMs: Number.NEGATIVE_INFINITY, sawDrop: false };
    state[signature] = slot;
  }

  if (slot.sent >= ERROR_BUDGET_LIMIT && nowMs - slot.lastSentMs < ERROR_BUDGET_INTERVAL_MS) {
    slot.sawDrop = true;
    return { send: false, truncated: false };
  }

  const truncated = slot.sent === ERROR_BUDGET_LIMIT - 1 || slot.sawDrop;
  slot.sent += 1;
  slot.lastSentMs = nowMs;
  slot.sawDrop = false;
  return { send: true, truncated };
}
