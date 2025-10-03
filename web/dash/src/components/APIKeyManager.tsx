'use client';

import {
  Alert,
  AlertDescription,
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  Input,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@sonr.io/ui';
import { CheckCircle, Copy, Eye, EyeOff, Key, Plus, RefreshCw, Trash2 } from 'lucide-react';
import { useState } from 'react';

interface APIKey {
  id: string;
  name: string;
  key: string;
  createdAt: string;
  lastUsed?: string;
  expiresAt?: string;
  status: 'active' | 'expired' | 'revoked';
}

export function APIKeyManager() {
  const [apiKeys, setApiKeys] = useState<APIKey[]>([
    {
      id: '1',
      name: 'Production API Key',
      key: 'sk_live_abc123...',
      createdAt: '2024-01-15',
      lastUsed: '2024-01-20',
      status: 'active',
    },
    {
      id: '2',
      name: 'Development API Key',
      key: 'sk_test_xyz789...',
      createdAt: '2024-01-10',
      lastUsed: '2024-01-19',
      status: 'active',
    },
  ]);

  const [showKey, setShowKey] = useState<string | null>(null);
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [newKeyName, setNewKeyName] = useState('');
  const [newKey, setNewKey] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);

  const handleCreateKey = async () => {
    setIsCreating(true);
    try {
      // TODO: Implement actual API key creation
      await new Promise((resolve) => setTimeout(resolve, 1500));
      const generatedKey = `sk_${Math.random().toString(36).substring(2, 15)}`;
      setNewKey(generatedKey);

      const newApiKey: APIKey = {
        id: Date.now().toString(),
        name: newKeyName,
        key: generatedKey,
        createdAt: new Date().toISOString().split('T')[0],
        status: 'active',
      };

      setApiKeys((prev) => [...prev, newApiKey]);
      setNewKeyName('');
    } catch (error) {
      console.error('Failed to create API key:', error);
    } finally {
      setIsCreating(false);
    }
  };

  const handleCopyKey = (key: string) => {
    navigator.clipboard.writeText(key);
    setCopiedKey(key);
    setTimeout(() => setCopiedKey(null), 2000);
  };

  const handleRevokeKey = async (keyId: string) => {
    try {
      // TODO: Implement actual API key revocation
      setApiKeys((prev) =>
        prev.map((key) => (key.id === keyId ? { ...key, status: 'revoked' as const } : key))
      );
    } catch (error) {
      console.error('Failed to revoke API key:', error);
    }
  };

  const handleRegenerateKey = async (keyId: string) => {
    try {
      // TODO: Implement actual API key regeneration
      const newKey = `sk_${Math.random().toString(36).substring(2, 15)}`;
      setApiKeys((prev) => prev.map((key) => (key.id === keyId ? { ...key, key: newKey } : key)));
    } catch (error) {
      console.error('Failed to regenerate API key:', error);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold">API Keys</h3>
          <p className="text-sm text-muted-foreground">
            Manage API keys for authenticating requests
          </p>
        </div>

        <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Create API Key
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New API Key</DialogTitle>
              <DialogDescription>Generate a new API key for your service</DialogDescription>
            </DialogHeader>

            {!newKey ? (
              <div className="space-y-4">
                <div className="space-y-2">
                  <label htmlFor="keyName" className="text-sm font-medium">
                    Key Name
                  </label>
                  <Input
                    id="keyName"
                    value={newKeyName}
                    onChange={(e) => setNewKeyName(e.target.value)}
                    placeholder="e.g., Production API Key"
                  />
                </div>

                <Button
                  onClick={handleCreateKey}
                  disabled={!newKeyName || isCreating}
                  className="w-full"
                >
                  {isCreating ? (
                    <>
                      <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                      Creating...
                    </>
                  ) : (
                    'Create Key'
                  )}
                </Button>
              </div>
            ) : (
              <div className="space-y-4">
                <Alert>
                  <CheckCircle className="h-4 w-4" />
                  <AlertDescription>
                    API key created successfully! Copy it now as it won't be shown again.
                  </AlertDescription>
                </Alert>

                <div className="space-y-2">
                  <label htmlFor="generatedKey" className="text-sm font-medium">
                    Your API Key
                  </label>
                  <div className="flex items-center gap-2">
                    <Code id="generatedKey" className="flex-1 p-2 text-xs font-mono">
                      {newKey}
                    </Code>
                    <Button size="sm" variant="outline" onClick={() => handleCopyKey(newKey)}>
                      {copiedKey === newKey ? (
                        <CheckCircle className="h-4 w-4" />
                      ) : (
                        <Copy className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>

                <Button
                  onClick={() => {
                    setShowCreateDialog(false);
                    setNewKey(null);
                  }}
                  className="w-full"
                >
                  Done
                </Button>
              </div>
            )}
          </DialogContent>
        </Dialog>
      </div>

      {/* API Keys Table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Active API Keys</CardTitle>
          <CardDescription>View and manage your service API keys</CardDescription>
        </CardHeader>
        <CardContent>
          {apiKeys.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Key</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Last Used</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {apiKeys.map((apiKey) => (
                  <TableRow key={apiKey.id}>
                    <TableCell className="font-medium">{apiKey.name}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <code className="text-xs">
                          {showKey === apiKey.id ? apiKey.key : `${apiKey.key.substring(0, 10)}...`}
                        </code>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => setShowKey(showKey === apiKey.id ? null : apiKey.id)}
                        >
                          {showKey === apiKey.id ? (
                            <EyeOff className="h-3 w-3" />
                          ) : (
                            <Eye className="h-3 w-3" />
                          )}
                        </Button>
                        <Button size="sm" variant="ghost" onClick={() => handleCopyKey(apiKey.key)}>
                          {copiedKey === apiKey.key ? (
                            <CheckCircle className="h-3 w-3" />
                          ) : (
                            <Copy className="h-3 w-3" />
                          )}
                        </Button>
                      </div>
                    </TableCell>
                    <TableCell>{apiKey.createdAt}</TableCell>
                    <TableCell>{apiKey.lastUsed || '-'}</TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          apiKey.status === 'active'
                            ? 'success'
                            : apiKey.status === 'expired'
                              ? 'warning'
                              : 'destructive'
                        }
                      >
                        {apiKey.status}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => handleRegenerateKey(apiKey.id)}
                          disabled={apiKey.status === 'revoked'}
                        >
                          <RefreshCw className="h-3 w-3" />
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => handleRevokeKey(apiKey.id)}
                          disabled={apiKey.status === 'revoked'}
                        >
                          <Trash2 className="h-3 w-3" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="text-center py-8">
              <Key className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-muted-foreground mb-4">No API keys created yet</p>
              <Button onClick={() => setShowCreateDialog(true)}>Create Your First API Key</Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Usage Instructions */}
      <Alert>
        <AlertDescription>
          Use your API key in the <code>Authorization</code> header:{' '}
          <code>Bearer YOUR_API_KEY</code>
        </AlertDescription>
      </Alert>
    </div>
  );
}

function Code({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={`bg-muted rounded px-2 py-1 font-mono ${className || ''}`}>{children}</div>
  );
}
