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
  Checkbox,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@sonr.io/ui';
import { Edit, Info, Plus, Shield, Trash2 } from 'lucide-react';
import { useState } from 'react';

interface Permission {
  id: string;
  name: string;
  description: string;
  scope: string;
  granted: boolean;
}

interface PermissionManagerProps {
  serviceId: string;
  permissions?: Permission[];
}

export function PermissionManager({ permissions = [] }: PermissionManagerProps) {
  const [selectedPermissions, setSelectedPermissions] = useState<string[]>([]);
  const [showAddDialog, setShowAddDialog] = useState(false);

  const permissionScopes = {
    'read:profile': 'Read user profile information',
    'write:profile': 'Update user profile',
    'read:data': 'Access user data',
    'write:data': 'Modify user data',
    'read:credentials': 'View credentials',
    'manage:credentials': 'Create and manage credentials',
    'read:vault': 'Access vault contents',
    'manage:vault': 'Full vault management',
    'execute:transactions': 'Execute blockchain transactions',
    'delegate:permissions': 'Delegate permissions to others',
  };

  const handlePermissionToggle = (permissionId: string) => {
    setSelectedPermissions((prev) =>
      prev.includes(permissionId)
        ? prev.filter((id) => id !== permissionId)
        : [...prev, permissionId]
    );
  };

  const handleSavePermissions = async () => {
    try {
      // TODO: Implement API call to save permissions
      console.log('Saving permissions:', selectedPermissions);
    } catch (error) {
      console.error('Failed to save permissions:', error);
    }
  };

  return (
    <div className="space-y-6">
      {/* Permission Summary */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold">Current Permissions</h3>
          <p className="text-sm text-muted-foreground">
            {permissions.filter((p) => p.granted).length} of {permissions.length} permissions
            granted
          </p>
        </div>

        <Dialog open={showAddDialog} onOpenChange={setShowAddDialog}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Add Permission
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>Add New Permission</DialogTitle>
              <DialogDescription>Select permissions to add to your service</DialogDescription>
            </DialogHeader>

            <div className="grid grid-cols-1 gap-3 mt-4">
              {Object.entries(permissionScopes).map(([scope, description]) => (
                <label
                  key={scope}
                  htmlFor={`permission-${scope}`}
                  className="flex items-start space-x-3 p-3 rounded-lg border cursor-pointer hover:bg-muted/50"
                >
                  <Checkbox
                    id={`permission-${scope}`}
                    checked={selectedPermissions.includes(scope)}
                    onCheckedChange={() => handlePermissionToggle(scope)}
                  />
                  <div className="space-y-1">
                    <div className="font-medium text-sm">{scope}</div>
                    <div className="text-xs text-muted-foreground">{description}</div>
                  </div>
                </label>
              ))}
            </div>

            <div className="flex justify-end gap-2 mt-4">
              <Button variant="outline" onClick={() => setShowAddDialog(false)}>
                Cancel
              </Button>
              <Button onClick={handleSavePermissions}>Add Permissions</Button>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      {/* UCAN Capabilities */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">UCAN Capabilities</CardTitle>
          <CardDescription>User-Controlled Authorization Network permissions</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {permissions.length > 0 ? (
              permissions.map((permission) => (
                <div
                  key={permission.id}
                  className="flex items-center justify-between p-3 rounded-lg border"
                >
                  <div className="flex items-center gap-3">
                    <Shield className="h-4 w-4 text-muted-foreground" />
                    <div>
                      <div className="font-medium text-sm">{permission.name}</div>
                      <div className="text-xs text-muted-foreground">{permission.description}</div>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <Badge variant={permission.granted ? 'success' : 'secondary'}>
                      {permission.granted ? 'Granted' : 'Pending'}
                    </Badge>
                    <Button size="sm" variant="ghost">
                      <Edit className="h-3 w-3" />
                    </Button>
                    <Button size="sm" variant="ghost">
                      <Trash2 className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
              ))
            ) : (
              <div className="text-center py-6 text-muted-foreground">
                No permissions configured yet
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Permission Audit Log */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Permission Audit Log</CardTitle>
          <CardDescription>Recent permission changes and access attempts</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {[
              {
                id: 'log1',
                action: 'Permission granted',
                scope: 'read:profile',
                time: '2 hours ago',
              },
              { id: 'log2', action: 'Permission revoked', scope: 'write:data', time: '1 day ago' },
              { id: 'log3', action: 'Access attempted', scope: 'manage:vault', time: '3 days ago' },
            ].map((log) => (
              <div key={log.id} className="flex items-center justify-between py-2 text-sm">
                <div>
                  <span className="font-medium">{log.action}:</span>{' '}
                  <code className="text-xs bg-muted px-1 py-0.5 rounded">{log.scope}</code>
                </div>
                <span className="text-xs text-muted-foreground">{log.time}</span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Info Alert */}
      <Alert>
        <Info className="h-4 w-4" />
        <AlertDescription>
          Permissions are managed through UCAN tokens. Changes may take up to 5 minutes to
          propagate.
        </AlertDescription>
      </Alert>
    </div>
  );
}
