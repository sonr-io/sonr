'use client';

import { Monitor, Moon, Sun } from 'lucide-react';
import { cn } from '../../../lib/utils';
import { Button } from '../../ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '../../ui/dropdown-menu';
import { Toggle } from '../../ui/toggle';

/**
 * Props for ThemeToggle component
 */
export interface ThemeToggleProps {
  theme?: 'light' | 'dark' | 'system';
  onThemeChange?: (theme: 'light' | 'dark' | 'system') => void;
  variant?: 'toggle' | 'dropdown';
  className?: string;
}

/**
 * Theme toggle component for dark mode switching
 */
export function ThemeToggle({
  theme = 'system',
  onThemeChange,
  variant = 'dropdown',
  className,
}: ThemeToggleProps) {
  if (variant === 'toggle') {
    return (
      <Toggle
        pressed={theme === 'dark'}
        onPressedChange={(pressed) => onThemeChange?.(pressed ? 'dark' : 'light')}
        aria-label="Toggle theme"
        className={className}
      >
        {theme === 'dark' ? <Moon className="h-4 w-4" /> : <Sun className="h-4 w-4" />}
      </Toggle>
    );
  }

  const getIcon = () => {
    switch (theme) {
      case 'light':
        return <Sun className="h-4 w-4" />;
      case 'dark':
        return <Moon className="h-4 w-4" />;
      default:
        return <Monitor className="h-4 w-4" />;
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className={cn('w-9 px-0', className)}>
          {getIcon()}
          <span className="sr-only">Toggle theme</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem
          onClick={() => onThemeChange?.('light')}
          className={cn(theme === 'light' && 'bg-accent')}
        >
          <Sun className="mr-2 h-4 w-4" />
          <span>Light</span>
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => onThemeChange?.('dark')}
          className={cn(theme === 'dark' && 'bg-accent')}
        >
          <Moon className="mr-2 h-4 w-4" />
          <span>Dark</span>
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => onThemeChange?.('system')}
          className={cn(theme === 'system' && 'bg-accent')}
        >
          <Monitor className="mr-2 h-4 w-4" />
          <span>System</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
