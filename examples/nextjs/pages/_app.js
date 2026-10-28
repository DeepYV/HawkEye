import { useEffect } from 'react';
import { useRouter } from 'next/router';
import { init, teardown, setRoute } from '@hawkeye/observer-sdk';
import { hawkeyeTemplate } from '../lib/hawkeye';

function MyApp({ Component, pageProps }) {
  const router = useRouter();

  useEffect(() => {
    init({
      apiKey: process.env.NEXT_PUBLIC_HAWKEYE_API_KEY,
      ingestionUrl: process.env.NEXT_PUBLIC_HAWKEYE_INGESTION_URL,
      appVersion: process.env.NEXT_PUBLIC_APP_VERSION,
      environment: process.env.NODE_ENV,
      events: ['signup_submitted'],
      importantFlows: [{ name: 'signup', steps: ['account'], important: true }],
    });

    return () => {
      teardown();
    };
  }, []);

  useEffect(() => {
    if (!router.pathname) return;
    setRoute(hawkeyeTemplate(router.pathname));
  }, [router.pathname]);

  return <Component {...pageProps} />;
}

export default MyApp;
