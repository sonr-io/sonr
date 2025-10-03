'use client';

import {
  ArrowLeftIcon,
  ArrowTopRightOnSquareIcon,
  ShieldCheckIcon,
  TrashIcon,
} from '@heroicons/react/24/outline';
import Link from 'next/link';
import { useEffect, useState } from 'react';

interface OAuthGrant {
  id: string;
  clientId: string;
  clientName: string;
  clientLogo?: string;
  scopes: string[];
  capabilities: UCANCapability[];
  grantedAt: string;
  lastUsed: string;
  expiresAt?: string;
  status: 'active' | 'expired' | 'revoked';
}

interface UCANCapability {
  action: string;
  resource: string;
  caveats?: Record<string, any>;
}

export default function CapabilitiesPage() {
  const [grants, setGrants] = useState<OAuthGrant[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedGrant, setSelectedGrant] = useState<OAuthGrant | null>(null);
  const [revoking, setRevoking] = useState<string | null>(null);

  useEffect(() => {
    fetchGrants();
  }, []);

  const fetchGrants = async () => {
    try {
      // TODO: Replace with actual API call
      const mockGrants: OAuthGrant[] = [
        {
          id: 'grant_1',
          clientId: 'example-app',
          clientName: 'Example Application',
          scopes: ['vault:read', 'vault:write', 'profile'],
          capabilities: [
            { action: 'read', resource: 'vault:*' },
            { action: 'write', resource: 'vault:*' },
            { action: 'read', resource: 'profile:*' },
          ],
          grantedAt: '2024-01-15T10:00:00Z',
          lastUsed: '2024-01-20T15:30:00Z',
          status: 'active',
        },
        {
          id: 'grant_2',
          clientId: 'defi-wallet',
          clientName: 'DeFi Wallet',
          scopes: ['dwn:read', 'svc:register'],
          capabilities: [
            { action: 'read', resource: 'dwn:*' },
            { action: 'register', resource: 'svc:*' },
          ],
          grantedAt: '2024-01-10T08:00:00Z',
          lastUsed: '2024-01-18T12:00:00Z',
          expiresAt: '2024-02-10T08:00:00Z',
          status: 'active',
        },
      ];

      setGrants(mockGrants);
    } catch (error) {
      console.error('Failed to fetch grants:', error);
    } finally {
      setLoading(false);
    }
  };

  const revokeGrant = async (grantId: string) => {
    setRevoking(grantId);
    try {
      // TODO: Implement actual revocation API call
      await new Promise((resolve) => setTimeout(resolve, 1000));

      setGrants((prev) =>
        prev.map((grant) =>
          grant.id === grantId ? { ...grant, status: 'revoked' as const } : grant
        )
      );

      if (selectedGrant?.id === grantId) {
        setSelectedGrant(null);
      }
    } catch (error) {
      console.error('Failed to revoke grant:', error);
    } finally {
      setRevoking(null);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusColor = (status: OAuthGrant['status']) => {
    switch (status) {
      case 'active':
        return 'bg-green-100 text-green-800';
      case 'expired':
        return 'bg-yellow-100 text-yellow-800';
      case 'revoked':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <Link
            href="/"
            className="inline-flex items-center text-sm text-gray-500 hover:text-gray-700 mb-4"
          >
            <ArrowLeftIcon className="w-4 h-4 mr-1" />
            Back to Home
          </Link>

          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">OAuth Capabilities</h1>
              <p className="mt-2 text-gray-600">
                Manage applications and services that have access to your account
              </p>
            </div>
            <ShieldCheckIcon className="w-10 h-10 text-blue-600" />
          </div>
        </div>

        {loading ? (
          <div className="flex justify-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
          </div>
        ) : grants.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-8 text-center">
            <ShieldCheckIcon className="w-16 h-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No Active Grants</h3>
            <p className="text-gray-600">
              You haven't granted any applications access to your account yet.
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Grants List */}
            <div className="lg:col-span-2 space-y-4">
              {grants.map((grant) => (
                <div
                  key={grant.id}
                  className={`bg-white rounded-lg shadow p-6 cursor-pointer transition-all hover:shadow-lg ${
                    selectedGrant?.id === grant.id ? 'ring-2 ring-blue-500' : ''
                  }`}
                  onClick={() => setSelectedGrant(grant)}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center">
                        {grant.clientLogo ? (
                          <img
                            src={grant.clientLogo}
                            alt={grant.clientName}
                            className="w-10 h-10 rounded-lg mr-3"
                          />
                        ) : (
                          <div className="w-10 h-10 bg-gray-200 rounded-lg mr-3 flex items-center justify-center">
                            <span className="text-gray-600 font-semibold">
                              {grant.clientName[0]}
                            </span>
                          </div>
                        )}
                        <div>
                          <h3 className="text-lg font-semibold text-gray-900">
                            {grant.clientName}
                          </h3>
                          <p className="text-sm text-gray-500">Client ID: {grant.clientId}</p>
                        </div>
                      </div>

                      <div className="mt-4 flex flex-wrap gap-2">
                        {grant.scopes.map((scope) => (
                          <span
                            key={scope}
                            className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800"
                          >
                            {scope}
                          </span>
                        ))}
                      </div>

                      <div className="mt-4 text-sm text-gray-600">
                        <p>Granted: {formatDate(grant.grantedAt)}</p>
                        <p>Last used: {formatDate(grant.lastUsed)}</p>
                        {grant.expiresAt && <p>Expires: {formatDate(grant.expiresAt)}</p>}
                      </div>
                    </div>

                    <div className="flex flex-col items-end ml-4">
                      <span
                        className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(grant.status)}`}
                      >
                        {grant.status}
                      </span>

                      {grant.status === 'active' && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            revokeGrant(grant.id);
                          }}
                          disabled={revoking === grant.id}
                          className="mt-4 inline-flex items-center px-3 py-1 border border-red-300 text-sm font-medium rounded-md text-red-700 bg-white hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50"
                        >
                          {revoking === grant.id ? (
                            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-red-700" />
                          ) : (
                            <>
                              <TrashIcon className="w-4 h-4 mr-1" />
                              Revoke
                            </>
                          )}
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* Grant Details */}
            {selectedGrant && (
              <div className="bg-white rounded-lg shadow p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-4">Capability Details</h3>

                <div className="space-y-4">
                  <div>
                    <h4 className="text-sm font-medium text-gray-700 mb-2">UCAN Capabilities</h4>
                    <div className="space-y-2">
                      {selectedGrant.capabilities.map((cap, index) => (
                        <div key={index} className="p-3 bg-gray-50 rounded-lg text-sm">
                          <div className="flex items-center justify-between">
                            <span className="font-medium text-gray-900">{cap.action}</span>
                            <span className="text-gray-600">{cap.resource}</span>
                          </div>
                          {cap.caveats && Object.keys(cap.caveats).length > 0 && (
                            <div className="mt-2 text-xs text-gray-500">
                              Caveats: {JSON.stringify(cap.caveats)}
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>

                  <div>
                    <h4 className="text-sm font-medium text-gray-700 mb-2">Delegation Chain</h4>
                    <button className="inline-flex items-center text-sm text-blue-600 hover:text-blue-500">
                      View full delegation chain
                      <ArrowTopRightOnSquareIcon className="w-3 h-3 ml-1" />
                    </button>
                  </div>

                  <div className="pt-4 border-t">
                    <h4 className="text-sm font-medium text-gray-700 mb-2">Audit Log</h4>
                    <p className="text-sm text-gray-600">View all activities for this grant</p>
                    <button className="mt-2 inline-flex items-center text-sm text-blue-600 hover:text-blue-500">
                      View audit log
                      <ArrowTopRightOnSquareIcon className="w-3 h-3 ml-1" />
                    </button>
                  </div>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
