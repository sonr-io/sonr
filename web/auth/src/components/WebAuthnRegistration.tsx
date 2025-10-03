/**
 * WebAuthn Registration Component with Email/Tel Support
 */

import React, { useState } from 'react';
import { useWebAuthn } from '../hooks/useWebAuthn';

interface RegistrationSuccessInfo {
  did: string;
  vaultId: string;
  assertionMethods: string[];
}

export function WebAuthnRegistration() {
  const { registerUser, isLoading, error, clearError } = useWebAuthn();
  
  // Form state
  const [username, setUsername] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [assertionType, setAssertionType] = useState<'email' | 'tel'>('email');
  const [assertionValue, setAssertionValue] = useState('');
  const [createVault, setCreateVault] = useState(true);
  
  // Success state
  const [success, setSuccess] = useState(false);
  const [registrationInfo, setRegistrationInfo] = useState<RegistrationSuccessInfo | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    clearError();
    setSuccess(false);
    setRegistrationInfo(null);

    try {
      const result = await registerUser(
        username,
        displayName || username,
        assertionType === 'email' ? assertionValue : undefined,
        assertionType === 'tel' ? assertionValue : undefined,
        createVault
      );

      if (result.success && result.did && result.vaultId) {
        setSuccess(true);
        setRegistrationInfo({
          did: result.did,
          vaultId: result.vaultId,
          assertionMethods: [
            `did:sonr:${username}`,
            `did:${assertionType}:${assertionValue}`
          ]
        });
      }
    } catch (err) {
      console.error('Registration failed:', err);
    }
  };

  const resetForm = () => {
    setUsername('');
    setDisplayName('');
    setAssertionValue('');
    setSuccess(false);
    setRegistrationInfo(null);
    clearError();
  };

  if (success && registrationInfo) {
    return (
      <div className="w-full max-w-md mx-auto p-6 bg-white rounded-lg shadow-md">
        <div className="mb-6">
          <h2 className="text-2xl font-bold text-green-600 mb-2">
            ✅ Registration Successful!
          </h2>
          <p className="text-gray-600">
            Your decentralized identity has been created successfully.
          </p>
        </div>

        <div className="space-y-4 mb-6">
          <div className="bg-gray-50 p-4 rounded">
            <h3 className="font-semibold text-gray-700 mb-2">DID Document</h3>
            <code className="block text-xs text-gray-600 break-all">
              {registrationInfo.did}
            </code>
          </div>

          <div className="bg-gray-50 p-4 rounded">
            <h3 className="font-semibold text-gray-700 mb-2">Vault ID</h3>
            <code className="block text-xs text-gray-600 break-all">
              {registrationInfo.vaultId}
            </code>
          </div>

          <div className="bg-gray-50 p-4 rounded">
            <h3 className="font-semibold text-gray-700 mb-2">Assertion Methods</h3>
            <ul className="space-y-1">
              {registrationInfo.assertionMethods.map((method, index) => (
                <li key={index} className="text-sm text-gray-600">
                  • {method}
                </li>
              ))}
            </ul>
          </div>

          <div className="bg-blue-50 p-4 rounded">
            <h3 className="font-semibold text-blue-700 mb-2">Authentication Method</h3>
            <p className="text-sm text-blue-600">
              WebAuthn credential successfully registered
            </p>
          </div>

          <div className="bg-purple-50 p-4 rounded">
            <h3 className="font-semibold text-purple-700 mb-2">UCAN Delegation</h3>
            <p className="text-sm text-purple-600">
              Origin token for wallet admin operations created
            </p>
          </div>
        </div>

        <button
          onClick={resetForm}
          className="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
        >
          Register Another Account
        </button>
      </div>
    );
  }

  return (
    <div className="w-full max-w-md mx-auto p-6 bg-white rounded-lg shadow-md">
      <h2 className="text-2xl font-bold mb-6">Create Your Sonr Identity</h2>
      
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-1">
            Username
          </label>
          <input
            type="text"
            id="username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
            className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="alice"
          />
        </div>

        <div>
          <label htmlFor="displayName" className="block text-sm font-medium text-gray-700 mb-1">
            Display Name (optional)
          </label>
          <input
            type="text"
            id="displayName"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Alice Smith"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Assertion Method
          </label>
          <div className="flex space-x-4 mb-2">
            <label className="flex items-center">
              <input
                type="radio"
                value="email"
                checked={assertionType === 'email'}
                onChange={(e) => setAssertionType(e.target.value as 'email')}
                className="mr-2"
              />
              <span>Email</span>
            </label>
            <label className="flex items-center">
              <input
                type="radio"
                value="tel"
                checked={assertionType === 'tel'}
                onChange={(e) => setAssertionType(e.target.value as 'tel')}
                className="mr-2"
              />
              <span>Phone</span>
            </label>
          </div>
          <input
            type={assertionType === 'email' ? 'email' : 'tel'}
            value={assertionValue}
            onChange={(e) => setAssertionValue(e.target.value)}
            required
            className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder={assertionType === 'email' ? 'alice@example.com' : '+1234567890'}
          />
        </div>

        <div className="flex items-center">
          <input
            type="checkbox"
            id="createVault"
            checked={createVault}
            onChange={(e) => setCreateVault(e.target.checked)}
            className="mr-2"
          />
          <label htmlFor="createVault" className="text-sm text-gray-700">
            Create secure vault for encrypted data storage
          </label>
        </div>

        {error && (
          <div className="p-3 bg-red-50 border border-red-200 rounded">
            <p className="text-sm text-red-600">{error}</p>
          </div>
        )}

        <button
          type="submit"
          disabled={isLoading}
          className="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
        >
          {isLoading ? 'Creating Identity...' : 'Create Identity with WebAuthn'}
        </button>
      </form>

      <div className="mt-6 p-4 bg-blue-50 rounded">
        <h3 className="text-sm font-semibold text-blue-900 mb-2">What will be created:</h3>
        <ul className="text-xs text-blue-700 space-y-1">
          <li>• W3C DID with Sonr address as controller</li>
          <li>• Two assertion methods (Sonr account + {assertionType === 'email' ? 'email' : 'phone'})</li>
          <li>• WebAuthn authentication method for passwordless login</li>
          <li>• UCAN delegation chain with validator proof</li>
          {createVault && <li>• Encrypted vault for secure data storage</li>}
        </ul>
      </div>
    </div>
  );
}