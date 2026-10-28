import assert from 'node:assert/strict';
import test from 'node:test';
import { flow, formOutcome, init, setRoute, teardown, track } from '../dist/index.js';

test('public functions are exported', () => {
  for (const fn of [init, track, formOutcome, flow, setRoute, teardown]) {
    assert.equal(typeof fn, 'function');
  }
});

test('flow before init returns a handle that does not throw', () => {
  const handle = flow('signup');
  assert.equal(typeof handle.start, 'function');
  assert.equal(typeof handle.step, 'function');
  assert.equal(typeof handle.complete, 'function');
  assert.equal(typeof handle.fail, 'function');
  assert.equal(typeof handle.abandon, 'function');
  assert.doesNotThrow(() => {
    handle.start();
    handle.step('account');
    handle.complete();
    handle.fail('host_failed');
    handle.abandon('host_abandoned');
  });
  assert.doesNotThrow(() => {
    track('signup_submitted', { plan: 'free' });
    formOutcome({ formKey: 'signup', result: 'success' });
    setRoute('/signup');
  });
});

test('init in Node does not throw and does not touch a DOM', () => {
  assert.equal(typeof window, 'undefined');
  assert.equal(typeof document, 'undefined');
  assert.doesNotThrow(() => {
    init({ apiKey: 'k', ingestionUrl: 'http://localhost:11224' });
  });
});

test('teardown before a browser init resolves', async () => {
  await assert.doesNotReject(() => teardown());
});
