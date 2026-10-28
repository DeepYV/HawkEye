/**
 * HawkEye vanilla JS example.
 * Clicks on data-hawkeye controls are automatic. This file only reports the host outcome.
 */

import { init, track, formOutcome, flow } from '@hawkeye/observer-sdk';

const config = {
  apiKey: 'dev-api-key',
  ingestionUrl: 'http://localhost:11224',
  appVersion: '1.4.0',
  environment: 'development',
  events: ['signup_submitted', 'demo_clicked'],
  importantFlows: [{ name: 'signup', steps: ['account'], important: true }],
};

document.addEventListener('DOMContentLoaded', () => {
  init(config);

  const signup = flow('signup');
  const form = document.getElementById('demo-form');
  form?.addEventListener('submit', (event) => {
    event.preventDefault();
    signup.start();
    formOutcome({ formKey: 'signup', result: 'success' });
    track('signup_submitted', { plan: 'free' });
    signup.complete();
  });

  document.getElementById('custom-event')?.addEventListener('click', () => {
    track('demo_clicked', { source: 'vanilla' });
  });
});
