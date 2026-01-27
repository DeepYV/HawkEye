import { useEffect } from 'react';
import { initFrustrationObserver, teardown } from '@hawkeye/observer-sdk';

function MyApp({ Component, pageProps }) {
  useEffect(() => {
    // Initialize HawkEye on client side only
    if (typeof window !== 'undefined') {
      initFrustrationObserver({
        apiKey: process.env.NEXT_PUBLIC_HAWKEYE_API_KEY,
        ingestionUrl: process.env.NEXT_PUBLIC_HAWKEYE_INGESTION_URL,
        enableDebug: process.env.NODE_ENV === 'development',
        environment: process.env.NODE_ENV,
        appId: 'my-nextjs-app',
      });
    }

    return () => {
      teardown();
    };
  }, []);

  return <Component {...pageProps} />;
}

export default MyApp;
