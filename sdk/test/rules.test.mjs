import assert from 'node:assert/strict';
import test from 'node:test';
import { takeErrorBudget } from '../dist/pure/budget.js';
import {
  DEAD_CLICK_MS,
  REPEATED_CLICK_GAP_MS,
  continuesBurst,
  isDeadElapsed,
} from '../dist/pure/constants.js';
import { businessPropAllowed } from '../dist/pure/grammar.js';
import { acceptRouteTemplate, normalizePath } from '../dist/pure/route.js';
import { signatureFromParts } from '../dist/pure/signature.js';
import { messageSlug } from '../dist/pure/slug.js';

test('token signup-submit becomes he:signup-submit', () => {
  const result = signatureFromParts({ hawkeye: 'signup-submit', tag: 'button' });
  assert.equal(result?.signature, 'he:signup-submit');
});

test('ignore on an ancestor yields no signature', () => {
  assert.equal(
    signatureFromParts({ ignored: true, hawkeye: 'signup-submit', tag: 'button' }),
    null,
  );
});

test('a plain button is ctrl:button:button', () => {
  assert.equal(signatureFromParts({ tag: 'button' })?.signature, 'ctrl:button:button');
});

test('aria-label Save draft becomes a label slug', () => {
  assert.equal(
    signatureFromParts({ tag: 'button', ariaLabel: 'Save draft' })?.signature,
    'ctrl:button:button:save-draft',
  );
});

test('aria-label next is excluded', () => {
  assert.equal(signatureFromParts({ tag: 'button', ariaLabel: 'next' }), null);
});

test('input type text is excluded', () => {
  assert.equal(signatureFromParts({ tag: 'input', inputType: 'text' }), null);
});

test('id save is ctrl:button:button#save', () => {
  assert.equal(signatureFromParts({ tag: 'button', id: 'save' })?.signature, 'ctrl:button:button#save');
});

test('uuid segment becomes :id', () => {
  assert.equal(
    normalizePath('/users/550e8400-e29b-41d4-a716-446655440000/settings'),
    '/users/:id/settings',
  );
});

test('digit segment becomes :id', () => {
  assert.equal(normalizePath('/posts/123'), '/posts/:id');
});

test('short non-id segment stays', () => {
  assert.equal(normalizePath('/posts/abc'), '/posts/abc');
});

test('a template containing ? is rejected', () => {
  assert.equal(acceptRouteTemplate('/search?q=1'), false);
  assert.equal(acceptRouteTemplate('/users/:id'), true);
});

test('a digit run in a JS message becomes #', () => {
  assert.equal(messageSlug('request failed 404'), 'request failed #');
});

test('a business prop named email is rejected', () => {
  assert.equal(businessPropAllowed('email'), false);
  assert.equal(businessPropAllowed('plan'), true);
});

test('error budget truncates the 10th, drops the 11th, keeps one after 10000 ms', () => {
  const state = {};
  let result;
  for (let i = 0; i < 9; i += 1) {
    result = takeErrorBudget(state, 'sig', 0);
    assert.equal(result.send, true);
    assert.equal(result.truncated, false);
  }
  result = takeErrorBudget(state, 'sig', 0);
  assert.equal(result.send, true);
  assert.equal(result.truncated, true);

  result = takeErrorBudget(state, 'sig', 9999);
  assert.equal(result.send, false);

  result = takeErrorBudget(state, 'sig', 10000);
  assert.equal(result.send, true);
  assert.equal(result.truncated, true);
});

test('gap under 1000 continues and 1000 does not', () => {
  assert.equal(REPEATED_CLICK_GAP_MS, 1000);
  assert.equal(continuesBurst(0), true);
  assert.equal(continuesBurst(999), true);
  assert.equal(continuesBurst(1000), false);
});

test('dead elapsed is >= 2500', () => {
  assert.equal(DEAD_CLICK_MS, 2500);
  assert.equal(isDeadElapsed(2499), false);
  assert.equal(isDeadElapsed(2500), true);
  assert.equal(isDeadElapsed(2501), true);
});
