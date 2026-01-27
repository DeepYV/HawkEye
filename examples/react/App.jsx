import React, { useEffect } from 'react';
import { initFrustrationObserver, captureEvent, teardown } from '@hawkeye/observer-sdk';

function App() {
  useEffect(() => {
    // Initialize HawkEye on component mount
    initFrustrationObserver({
      apiKey: 'your-api-key-here',
      ingestionUrl: 'https://api.hawkeye.example.com',
      enableDebug: process.env.NODE_ENV === 'development',
      environment: process.env.NODE_ENV,
      appId: 'my-react-app',
      batchSize: 10,
      batchInterval: 5000,
    });

    // Cleanup on unmount
    return () => {
      teardown();
    };
  }, []);

  const handleCustomEvent = () => {
    // Track custom events
    captureEvent(
      'custom_action',
      { type: 'button', id: 'custom-button' },
      { feature: 'premium', action: 'upgrade_clicked' }
    );
  };

  return (
    <div className="App">
      <h1>HawkEye React Example</h1>
      <p>User interactions are automatically tracked</p>
      <button onClick={handleCustomEvent}>
        Track Custom Event
      </button>
    </div>
  );
}

export default App;
