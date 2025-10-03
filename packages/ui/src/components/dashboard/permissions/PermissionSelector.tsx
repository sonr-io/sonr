'use client';

import { Info } from 'lucide-react';
import { useState } from 'react';
import { cn } from '../../../lib/utils';
import { Badge } from '../../ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from '../../ui/select';

/**
 * Permission template structure
 */
export interface PermissionTemplate {
  id: string;
  name: string;
  description: string;
  permissions: string[];
  category: 'basic' | 'standard' | 'advanced' | 'custom';
  icon?: React.ReactNode;
}

/**
 * Props for PermissionSelector component
 */
export interface PermissionSelectorProps {
  templates: PermissionTemplate[];
  value?: string;
  onChange?: (templateId: string, template: PermissionTemplate) => void;
  showDetails?: boolean;
  disabled?: boolean;
  className?: string;
}

/**
 * Dropdown selector for permission templates with descriptions
 */
export function PermissionSelector({
  templates,
  value,
  onChange,
  showDetails = true,
  disabled = false,
  className,
}: PermissionSelectorProps) {
  const [selectedTemplate, setSelectedTemplate] = useState<PermissionTemplate | undefined>(
    templates.find((t) => t.id === value)
  );

  const handleChange = (templateId: string) => {
    const template = templates.find((t) => t.id === templateId);
    if (template) {
      setSelectedTemplate(template);
      onChange?.(templateId, template);
    }
  };

  const getCategoryColor = (category: PermissionTemplate['category']) => {
    switch (category) {
      case 'basic':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-400';
      case 'standard':
        return 'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400';
      case 'advanced':
        return 'bg-orange-100 text-orange-800 dark:bg-orange-900/20 dark:text-orange-400';
      case 'custom':
        return 'bg-purple-100 text-purple-800 dark:bg-purple-900/20 dark:text-purple-400';
      default:
        return '';
    }
  };

  const groupedTemplates = templates.reduce<Record<string, PermissionTemplate[]>>(
    (acc, template) => {
      const category = template.category;
      if (!acc[category]) {
        acc[category] = [];
      }
      acc[category]?.push(template);
      return acc;
    },
    {}
  );

  return (
    <div className={cn('space-y-4', className)}>
      <Select value={selectedTemplate?.id || ''} onValueChange={handleChange} disabled={disabled}>
        <SelectTrigger className="w-full">
          <SelectValue placeholder="Select a permission template" />
        </SelectTrigger>
        <SelectContent>
          {Object.entries(groupedTemplates).map(([category, temps]) => (
            <SelectGroup key={category}>
              <SelectLabel className="flex items-center gap-2">
                <Badge
                  variant="secondary"
                  className={cn(
                    'text-xs',
                    getCategoryColor(category as PermissionTemplate['category'])
                  )}
                >
                  {category}
                </Badge>
              </SelectLabel>
              {temps.map((template) => (
                <SelectItem key={template.id} value={template.id} className="cursor-pointer">
                  <div className="flex items-center gap-2">
                    {template.icon}
                    <div>
                      <div className="font-medium">{template.name}</div>
                      <div className="text-xs text-muted-foreground">{template.description}</div>
                    </div>
                  </div>
                </SelectItem>
              ))}
            </SelectGroup>
          ))}
        </SelectContent>
      </Select>

      {showDetails && selectedTemplate && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-base">{selectedTemplate.name}</CardTitle>
              <Badge variant="secondary" className={getCategoryColor(selectedTemplate.category)}>
                {selectedTemplate.category}
              </Badge>
            </div>
            <CardDescription>{selectedTemplate.description}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Info className="h-4 w-4" />
                <span>
                  This template includes {selectedTemplate.permissions.length} permissions:
                </span>
              </div>
              <div className="flex flex-wrap gap-2">
                {selectedTemplate.permissions.map((permission) => (
                  <Badge key={permission} variant="outline" className="text-xs">
                    {permission}
                  </Badge>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
