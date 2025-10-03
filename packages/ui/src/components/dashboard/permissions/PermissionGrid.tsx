'use client';

import { useState } from 'react';
import { cn } from '../../../lib/utils';
import { Badge } from '../../ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';
import { Checkbox } from '../../ui/checkbox';
import { Label } from '../../ui/label';

/**
 * Permission item structure
 */
export interface Permission {
  id: string;
  name: string;
  description: string;
  category: string;
  resource?: string;
  action?: string;
  enabled: boolean;
  required?: boolean;
}

/**
 * Props for PermissionGrid component
 */
export interface PermissionGridProps {
  permissions: Permission[];
  onChange?: (permissions: Permission[]) => void;
  readOnly?: boolean;
  groupByCategory?: boolean;
  showDescriptions?: boolean;
  className?: string;
}

/**
 * Grid display for managing multiple permissions
 */
export function PermissionGrid({
  permissions,
  onChange,
  readOnly = false,
  groupByCategory = true,
  showDescriptions = true,
  className,
}: PermissionGridProps) {
  const [selectedPermissions, setSelectedPermissions] = useState<Permission[]>(permissions);

  const handlePermissionChange = (permission: Permission, checked: boolean) => {
    const updated = selectedPermissions.map((p) =>
      p.id === permission.id ? { ...p, enabled: checked } : p
    );
    setSelectedPermissions(updated);
    onChange?.(updated);
  };

  const groupedPermissions = groupByCategory
    ? selectedPermissions.reduce<Record<string, Permission[]>>((acc, permission) => {
        const category = permission.category;
        if (!acc[category]) {
          acc[category] = [];
        }
        acc[category]?.push(permission);
        return acc;
      }, {})
    : { 'All Permissions': selectedPermissions };

  return (
    <div className={cn('space-y-4', className)}>
      {Object.entries(groupedPermissions).map(([category, perms]) => (
        <Card key={category}>
          <CardHeader>
            <CardTitle className="text-base">{category}</CardTitle>
            <CardDescription>
              {perms.filter((p) => p.enabled).length} of {perms.length} permissions enabled
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {perms.map((permission) => (
                <div
                  key={permission.id}
                  className={cn(
                    'flex flex-col space-y-2 rounded-lg border p-3',
                    permission.enabled && 'bg-accent/50',
                    permission.required && 'border-primary'
                  )}
                >
                  <div className="flex items-start space-x-2">
                    <Checkbox
                      id={permission.id}
                      checked={permission.enabled}
                      onCheckedChange={(checked) =>
                        handlePermissionChange(permission, checked as boolean)
                      }
                      disabled={readOnly || permission.required}
                      className="mt-0.5"
                    />
                    <div className="flex-1 space-y-1">
                      <Label
                        htmlFor={permission.id}
                        className={cn(
                          'text-sm font-medium cursor-pointer',
                          readOnly && 'cursor-not-allowed opacity-60'
                        )}
                      >
                        {permission.name}
                        {permission.required && (
                          <Badge variant="secondary" className="ml-2 text-xs">
                            Required
                          </Badge>
                        )}
                      </Label>
                      {showDescriptions && (
                        <p className="text-xs text-muted-foreground">{permission.description}</p>
                      )}
                      {(permission.resource || permission.action) && (
                        <div className="flex gap-2 mt-1">
                          {permission.resource && (
                            <Badge variant="outline" className="text-xs">
                              {permission.resource}
                            </Badge>
                          )}
                          {permission.action && (
                            <Badge variant="outline" className="text-xs">
                              {permission.action}
                            </Badge>
                          )}
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
