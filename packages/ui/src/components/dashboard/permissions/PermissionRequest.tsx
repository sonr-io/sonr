'use client';

import { AlertCircle, Clock, FileText, Info, Key, Shield, User } from 'lucide-react';
import { useState } from 'react';
import { Alert, AlertDescription } from '../../ui/alert';
import { Badge } from '../../ui/badge';
import { Button } from '../../ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '../../ui/card';
import { Checkbox } from '../../ui/checkbox';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../../ui/dialog';
import { Input } from '../../ui/input';
import { Label } from '../../ui/label';
import { Textarea } from '../../ui/textarea';
// import { cn } from "../../../lib/utils" // Uncomment if needed

/**
 * Permission request data structure
 */
export interface PermissionRequestData {
  requester: {
    id: string;
    name: string;
    type: 'user' | 'service';
  };
  permissions: Array<{
    resource: string;
    action: string;
    reason?: string;
  }>;
  duration?: {
    value: number;
    unit: 'hours' | 'days' | 'weeks' | 'months';
  };
  justification: string;
  metadata?: Record<string, any>;
}

/**
 * Props for PermissionRequest component
 */
export interface PermissionRequestProps {
  onSubmit?: (request: PermissionRequestData) => Promise<void>;
  availablePermissions?: Array<{
    resource: string;
    actions: string[];
  }>;
  requester?: PermissionRequestData['requester'];
  maxDuration?: number;
  showJustification?: boolean;
  className?: string;
}

/**
 * Dialog workflow for requesting permissions
 */
export function PermissionRequest({
  onSubmit,
  availablePermissions = [],
  requester,
  maxDuration = 30,
  showJustification = true,
  className,
}: PermissionRequestProps) {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>();
  const [formData, setFormData] = useState<Partial<PermissionRequestData>>({
    requester: requester || { id: '', name: '', type: 'user' },
    permissions: [],
    justification: '',
  });
  const [selectedPermissions, setSelectedPermissions] = useState<Set<string>>(new Set());

  const handlePermissionToggle = (resource: string, action: string) => {
    const key = `${resource}:${action}`;
    const newSelected = new Set(selectedPermissions);

    if (newSelected.has(key)) {
      newSelected.delete(key);
    } else {
      newSelected.add(key);
    }

    setSelectedPermissions(newSelected);

    // Update form data
    const permissions = Array.from(newSelected).map((k) => {
      const [res, act] = k.split(':');
      return { resource: res || '', action: act || '' };
    });
    setFormData((prev) => ({ ...prev, permissions }));
  };

  const handleSubmit = async () => {
    if (!formData.permissions || formData.permissions.length === 0) {
      setError('Please select at least one permission');
      return;
    }

    if (showJustification && !formData.justification) {
      setError('Please provide a justification');
      return;
    }

    setLoading(true);
    setError(undefined);

    try {
      await onSubmit?.(formData as PermissionRequestData);
      setOpen(false);
      // Reset form
      setFormData({
        requester: requester || { id: '', name: '', type: 'user' },
        permissions: [],
        justification: '',
      });
      setSelectedPermissions(new Set());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit request');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button className={className}>
          <Shield className="h-4 w-4 mr-2" />
          Request Permissions
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Request Permissions</DialogTitle>
          <DialogDescription>
            Select the permissions you need and provide justification for your request
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Requester Information */}
          {requester && (
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm">Requester</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-3">
                  <User className="h-4 w-4 text-muted-foreground" />
                  <div>
                    <p className="text-sm font-medium">{requester.name}</p>
                    <p className="text-xs text-muted-foreground">
                      {requester.type === 'service' ? 'Service Account' : 'User Account'}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Permission Selection */}
          <div className="space-y-3">
            <Label>Select Permissions</Label>
            {availablePermissions.length === 0 ? (
              <Alert>
                <Info className="h-4 w-4" />
                <AlertDescription>No permissions available to request</AlertDescription>
              </Alert>
            ) : (
              <div className="space-y-3">
                {availablePermissions.map((perm) => (
                  <Card key={perm.resource}>
                    <CardHeader className="pb-3">
                      <div className="flex items-center gap-2">
                        <Key className="h-4 w-4 text-muted-foreground" />
                        <CardTitle className="text-sm">{perm.resource}</CardTitle>
                      </div>
                    </CardHeader>
                    <CardContent>
                      <div className="grid grid-cols-2 gap-3">
                        {perm.actions.map((action) => {
                          const key = `${perm.resource}:${action}`;
                          return (
                            <div key={action} className="flex items-center space-x-2">
                              <Checkbox
                                id={key}
                                checked={selectedPermissions.has(key)}
                                onCheckedChange={() =>
                                  handlePermissionToggle(perm.resource, action)
                                }
                              />
                              <Label htmlFor={key} className="text-sm cursor-pointer">
                                {action}
                              </Label>
                            </div>
                          );
                        })}
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </div>

          {/* Duration Selection */}
          <div className="space-y-3">
            <Label htmlFor="duration">Duration (Optional)</Label>
            <div className="flex gap-2">
              <Input
                id="duration"
                type="number"
                min="1"
                max={maxDuration}
                placeholder="Duration"
                value={formData.duration?.value || ''}
                onChange={(e) => {
                  const value = Number.parseInt(e.target.value);
                  if (value > 0) {
                    setFormData((prev) => ({
                      ...prev,
                      duration: {
                        value,
                        unit: prev.duration?.unit || 'days',
                      },
                    }));
                  }
                }}
                className="w-24"
              />
              <select
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={formData.duration?.unit || 'days'}
                onChange={(e) => {
                  setFormData((prev) => ({
                    ...prev,
                    duration: {
                      value: prev.duration?.value || 7,
                      unit: e.target.value as any,
                    },
                  }));
                }}
              >
                <option value="hours">Hours</option>
                <option value="days">Days</option>
                <option value="weeks">Weeks</option>
                <option value="months">Months</option>
              </select>
            </div>
            <p className="text-xs text-muted-foreground">
              Leave empty for permanent permissions (subject to approval)
            </p>
          </div>

          {/* Justification */}
          {showJustification && (
            <div className="space-y-3">
              <Label htmlFor="justification">
                Justification <span className="text-red-500">*</span>
              </Label>
              <Textarea
                id="justification"
                placeholder="Explain why you need these permissions..."
                value={formData.justification}
                onChange={(e) =>
                  setFormData((prev) => ({ ...prev, justification: e.target.value }))
                }
                rows={4}
              />
              <p className="text-xs text-muted-foreground">
                Provide a clear business justification for this request
              </p>
            </div>
          )}

          {/* Error Display */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Selected Permissions Summary */}
          {selectedPermissions.size > 0 && (
            <Card className="bg-muted/50">
              <CardHeader className="pb-3">
                <CardTitle className="text-sm flex items-center gap-2">
                  <FileText className="h-4 w-4" />
                  Request Summary
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Clock className="h-3 w-3 text-muted-foreground" />
                    <span className="text-xs text-muted-foreground">
                      {formData.duration
                        ? `${formData.duration.value} ${formData.duration.unit}`
                        : 'Permanent (subject to approval)'}
                    </span>
                  </div>
                  <div className="flex flex-wrap gap-1">
                    {Array.from(selectedPermissions).map((key) => (
                      <Badge key={key} variant="secondary" className="text-xs">
                        {key}
                      </Badge>
                    ))}
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)} disabled={loading}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={loading || selectedPermissions.size === 0}>
            {loading ? 'Submitting...' : 'Submit Request'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
