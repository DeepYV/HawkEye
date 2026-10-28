# React

Call `init` once from a `useEffect` in the root component. Return `teardown` from that effect. Do not initialize in every component.

```jsx
import { useEffect } from 'react';
import { init, teardown, track, formOutcome, flow } from '@hawkeye/observer-sdk';

function App() {
  useEffect(() => {
    init({
      apiKey: 'your-api-key-here',
      ingestionUrl: 'http://localhost:11224',
      appVersion: '1.4.0',
      events: ['signup_submitted'],
      importantFlows: [{ name: 'signup', steps: ['account'], important: true }],
    });
    return () => {
      teardown();
    };
  }, []);

  const submit = () => {
    const signup = flow('signup');
    signup.start();
    formOutcome({ formKey: 'signup', result: 'success' });
    track('signup_submitted', { plan: 'free' });
    signup.complete();
  };

  return (
    <form data-hawkeye="signup">
      <button type="submit" data-hawkeye="signup-submit" onClick={submit}>
        Create account
      </button>
    </form>
  );
}
```

`useHawkEye` is the same `init` / `teardown` pair:

```jsx
import { useHawkEye } from './useHawkEye';

function App() {
  useHawkEye({
    apiKey: 'your-api-key-here',
    ingestionUrl: 'http://localhost:11224',
    events: ['signup_submitted'],
    importantFlows: [{ name: 'signup', steps: ['account'], important: true }],
  });
  return <div>Your app</div>;
}
```

For more than one step, call `flow(name).step` in order, then `complete()` once the final step is the furthest step.

Clicks, dead clicks, navigation, and errors are automatic. Mark the control with `data-hawkeye`. Put `data-hawkeye-ignore` on chrome that should not count, such as a carousel control. The SDK does not record the DOM, input values, or scroll position.

Props named `email`, `password`, `name`, and the rest of the denylist drop the `track` event. Do not send field values.
