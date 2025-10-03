'use client';

import { useEffect } from 'react';

export default function TestEcho() {
  useEffect(() => {
    console.log('Testing browser-echo from dashboard');
    console.info('This is an info message');
    console.warn('This is a warning message');
    console.error('This is an error message');
    console.debug('This is a debug message');
  }, []);

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold mb-4">Browser Echo Test</h1>
      <p>Check the terminal to see if console logs are being streamed!</p>
      <button
        type="button"
        onClick={() => {
          console.log('Button clicked at:', new Date().toISOString());
        }}
        className="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
      >
        Click to test console.log
      </button>
    </div>
  );
}
