'use client';

import { ChevronDownIcon, ChevronRightIcon, LinkIcon } from '@heroicons/react/24/outline';
import { useState } from 'react';

interface DelegationNode {
  id: string;
  issuer: string;
  audience: string;
  capabilities: UCANCapability[];
  expiresAt?: string;
  issuedAt: string;
  signature: string;
  proofs: string[];
  children?: DelegationNode[];
}

interface UCANCapability {
  action: string;
  resource: string;
  caveats?: Record<string, any>;
}

interface DelegationChainViewerProps {
  rootDelegation: DelegationNode;
  className?: string;
}

export function DelegationChainViewer({
  rootDelegation,
  className = '',
}: DelegationChainViewerProps) {
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set([rootDelegation.id]));

  const toggleNode = (nodeId: string) => {
    setExpandedNodes((prev) => {
      const next = new Set(prev);
      if (next.has(nodeId)) {
        next.delete(nodeId);
      } else {
        next.add(nodeId);
      }
      return next;
    });
  };

  const formatDID = (did: string) => {
    if (did.length > 20) {
      return `${did.slice(0, 8)}...${did.slice(-8)}`;
    }
    return did;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const renderNode = (node: DelegationNode, depth = 0): JSX.Element => {
    const isExpanded = expandedNodes.has(node.id);
    const hasChildren = node.children && node.children.length > 0;

    return (
      <div key={node.id} className="relative">
        {/* Connection line for non-root nodes */}
        {depth > 0 && (
          <div
            className="absolute left-6 top-0 w-px bg-gray-300"
            style={{ height: '24px', transform: 'translateY(-24px)' }}
          />
        )}

        <div className={`relative ${depth > 0 ? 'ml-8' : ''}`}>
          {/* Node container */}
          <div className="group relative bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow">
            {/* Expand/collapse button */}
            {hasChildren && (
              <button
                onClick={() => toggleNode(node.id)}
                className="absolute -left-3 top-6 w-6 h-6 bg-white border border-gray-300 rounded-full flex items-center justify-center hover:bg-gray-50"
              >
                {isExpanded ? (
                  <ChevronDownIcon className="w-4 h-4 text-gray-600" />
                ) : (
                  <ChevronRightIcon className="w-4 h-4 text-gray-600" />
                )}
              </button>
            )}

            {/* Node content */}
            <div className="space-y-3">
              {/* Header */}
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <LinkIcon className="w-4 h-4 text-gray-400" />
                    <span className="text-sm font-medium text-gray-900">
                      Delegation {depth > 0 ? `(Level ${depth})` : '(Root)'}
                    </span>
                  </div>
                  <p className="mt-1 text-xs text-gray-500">ID: {node.id.slice(0, 16)}...</p>
                </div>
                <span className="text-xs text-gray-500">{formatDate(node.issuedAt)}</span>
              </div>

              {/* Issuer and Audience */}
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div>
                  <span className="text-xs text-gray-500">Issuer:</span>
                  <p className="font-mono text-xs mt-1">{formatDID(node.issuer)}</p>
                </div>
                <div>
                  <span className="text-xs text-gray-500">Audience:</span>
                  <p className="font-mono text-xs mt-1">{formatDID(node.audience)}</p>
                </div>
              </div>

              {/* Capabilities */}
              <div>
                <span className="text-xs text-gray-500">Capabilities:</span>
                <div className="mt-1 flex flex-wrap gap-1">
                  {node.capabilities.map((cap, idx) => (
                    <span
                      key={idx}
                      className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800"
                    >
                      {cap.action}:{cap.resource}
                    </span>
                  ))}
                </div>
              </div>

              {/* Expiration */}
              {node.expiresAt && (
                <div className="text-xs">
                  <span className="text-gray-500">Expires:</span>
                  <span
                    className={`ml-2 ${new Date(node.expiresAt) < new Date() ? 'text-red-600' : 'text-gray-900'}`}
                  >
                    {formatDate(node.expiresAt)}
                  </span>
                </div>
              )}

              {/* Signature preview */}
              <div className="text-xs">
                <span className="text-gray-500">Signature:</span>
                <span className="ml-2 font-mono text-gray-600">
                  {node.signature.slice(0, 20)}...
                </span>
              </div>
            </div>
          </div>

          {/* Children nodes */}
          {hasChildren && isExpanded && (
            <div className="relative mt-4 space-y-4">
              {/* Vertical connection line */}
              {node.children!.map(
                (_, index) =>
                  index < node.children!.length - 1 && (
                    <div
                      key={`line-${index}`}
                      className="absolute left-6 top-6 w-px bg-gray-300"
                      style={{
                        height: `calc(100% - 24px)`,
                        transform: 'translateY(24px)',
                      }}
                    />
                  )
              )}

              {node.children!.map((child) => renderNode(child, depth + 1))}
            </div>
          )}
        </div>
      </div>
    );
  };

  return <div className={`delegation-chain-viewer ${className}`}>{renderNode(rootDelegation)}</div>;
}
