import React, { useEffect } from 'react';
import { init, teardown, track, formOutcome, flow } from '@hawkeye/observer-sdk';

function App() {
  useEffect(() => {
    init({
      apiKey: 'your-api-key-here',
      ingestionUrl: 'http://localhost:11224',
      appVersion: '1.4.0',
      environment: process.env.NODE_ENV,
      events: ['signup_submitted'],
      importantFlows: [{ name: 'signup', steps: ['account'], important: true }],
    });

    return () => {
      teardown();
    };
  }, []);

  const handleSignup = () => {
    const signup = flow('signup');
    signup.start();
    formOutcome({ formKey: 'signup', result: 'success' });
    track('signup_submitted', { plan: 'free' });
    signup.complete();
  };

  return (
    <div className="App">
      <h1>HawkEye React example</h1>
      <form data-hawkeye="signup" onSubmit={(event) => event.preventDefault()}>
        <button type="submit" data-hawkeye="signup-submit" onClick={handleSignup}>
          Create account
        </button>
      </form>
    </div>
  );
}

export default App;
