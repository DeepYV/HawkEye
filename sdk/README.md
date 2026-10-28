# @hawkeye/observer-sdk

Browser SDK for dead clicks, failed actions, and incomplete flows. It does not record the DOM, session replay, scroll position, input values, or mouse movement.

## Install

```bash
npm install @hawkeye/observer-sdk
```

## Quick start

Call `init` from the browser, after the page exists. In React and Next.js, call it from `useEffect` and return `teardown` from that effect. Do not call `init` at module scope in a server component.

```javascript
import { init, track, formOutcome, flow } from '@hawkeye/observer-sdk';

init({
  apiKey: 'dev-key',
  ingestionUrl: 'http://localhost:8080',
  appVersion: '1.4.0',
  events: ['signup_submitted'],
  importantFlows: [{ name: 'signup', steps: ['account'], important: true }],
});

// <form data-hawkeye="signup"><button type="submit" data-hawkeye="signup-submit">
const signup = flow('signup');
signup.start();
formOutcome({ formKey: 'signup', result: 'success' });
track('signup_submitted');
signup.complete();
```

`init` returns immediately when `window` is missing, so a Next.js module can be imported during SSR. In the browser, a missing `apiKey` or `ingestionUrl` throws. A second `init` keeps the first configuration.

## Marking a control

```html
<form data-hawkeye="signup">
  <button type="submit" data-hawkeye="signup-submit">Create account</button>
</form>

<div data-hawkeye-ignore>
  <button type="button">Carousel next</button>
</div>
```

`data-hawkeye` is a token: lowercase letters, digits, and hyphens, starting with a letter or digit, at most 64 characters. The nearest token becomes the click signature `he:<token>`. `formOutcome({ formKey })` must use that same token.

`data-hawkeye-ignore` on the target or an ancestor up to `body` drops the click. Navigation, errors, and host calls still emit.

Without a token, a button, link, or other control can still be reported as `ctrl:<tag>:<role>`, optionally with an `aria-label` slug or a short `id`. Text fields, modifier clicks, and non-primary buttons are not clicks. Stepper labels such as `next` and `previous` are not clicks.

When an action starts, set `disabled`, `aria-busy`, or `aria-expanded` on the marked element. The SDK watches only that element. A spinner that is a child is not a response.

## What the SDK sends

After `init` in the browser, the SDK emits `click`, `click_outcome`, `navigation`, and `error` on its own. The host adds `form_outcome`, `business` (`track`), and `flow_step` (`flow`).

A click that gets no navigation, matching host outcome, attribute change, unload, or link blur within 2500 ms is reported as `click_outcome` with `result: "dead"`. Leaving the page before that deadline is a response (`unload`), not a dead click.

Routes come from `setRoute(template)` when the host supplies one. Otherwise the SDK uses `location.pathname`, drops the query and hash, and replaces digit, long-hex, and UUID segments with `:id`. A template that contains `?`, `#`, `@`, or `%` is ignored.

JavaScript errors come from `window` `error` and `unhandledrejection` only. HTTP failures are status `0` or `>= 400` from `fetch` and `XMLHttpRequest`. Response bodies are not read. The HawkEye ingest URL is not reported.

Events are posted to `{ingestionUrl}/v1/events` with header `X-API-Key`, in batches of 10 or every 5 seconds. `pagehide` and a hidden document flush with `keepalive`.

## API

```ts
init(options: InitOptions): void
track(name: string, props?: Record<string, string | number | boolean>): void
formOutcome(input: FormOutcomeInput): void
flow(name: string): FlowHandle
setRoute(template: string): void
teardown(): Promise<void>
```

`InitOptions`: `apiKey`, `ingestionUrl`, and optional `appVersion`, `appBuild`, `appNamespace`, `releaseId`, `environment`, `importantFlows`, `events`.

`importantFlows` is the whole registry. A name that is not in it is ignored. `events` is the allowlist for `track`. If `events` is omitted, every `track` is dropped. At most 8 props. A string longer than 64 characters drops the event. These prop names drop the event: `password`, `email`, `token`, `secret`, `name`, `query`, `html`, `value`, `text`, `dom`, `href`, `selector`, `message`, `input`, `content`.

```ts
const signup = flow('signup');
signup.start();
signup.step('verify'); // only the next step, in order
signup.complete();      // when the final step is already the furthest step
signup.fail('timeout'); // or signup.abandon('closed')
```

`formOutcome` results are `success`, `validation_error`, and `submit_error`. `fieldKeys` are names only. Field values are not accepted.

`teardown` removes listeners, restores `fetch` and XHR, and flushes the queue. It does not clear `localStorage`.

`flow`, `track`, and `formOutcome` before `init` do nothing.

## What is not collected

The SDK does not record input values, text, selection, scroll offsets, class lists, CSS selectors, element chains, coordinates, href values, network bodies, or console arguments. It does not start a document mutation observer, a scroll listener, an input listener, or a mouse-move listener.

## Framework notes

- React: [examples/react](../examples/react/README.md)
- Next.js: [examples/nextjs](../examples/nextjs/README.md). Pass the path template with `setRoute`. Initialize in `useEffect`, not in a server component.
- Vanilla: [examples/vanilla-js](../examples/vanilla-js/README.md)
