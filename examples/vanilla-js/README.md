# Vanilla JavaScript

```javascript
import { init, track, formOutcome, flow, setRoute, teardown } from '@hawkeye/observer-sdk';

init({
  apiKey: 'dev-key',
  ingestionUrl: 'http://localhost:8080',
  appVersion: '1.4.0',
  events: ['signup_submitted'],
  importantFlows: [{ name: 'signup', steps: ['account'], important: true }],
});

const signup = flow('signup');
signup.start();
formOutcome({ formKey: 'signup', result: 'success' });
track('signup_submitted', { plan: 'free' });
signup.complete();
```

Mark the control in HTML:

```html
<form data-hawkeye="signup">
  <button type="submit" data-hawkeye="signup-submit">Create account</button>
</form>

<div data-hawkeye-ignore>
  <button type="button">Carousel next</button>
</div>
```

`init` is enough for a full page load. The SDK flushes on `pagehide`. Call `teardown()` when a single-page app leaves the page that started the SDK. Call `setRoute('/users/:id')` when the visible path changes and the template is not the raw pathname.

Clicks, dead clicks (2500 ms), navigation, JavaScript errors, and failed requests are automatic. The SDK does not record the DOM, input values, scroll position, or mouse movement.

`track` only sends names listed in `events`. A prop named `email`, `password`, `name`, or another denylisted key drops that event. `formOutcome` takes field keys, not field values.

See `index.html` for a page that imports the local build at `sdk/dist/index.js`.
