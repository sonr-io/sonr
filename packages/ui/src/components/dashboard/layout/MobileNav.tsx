'use client';

import { Menu } from 'lucide-react';
import { useState } from 'react';
import { Button } from '../../ui/button';
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '../../ui/sheet';
import type { NavItem } from './DashboardSidebar';
import { DashboardSidebar } from './DashboardSidebar';

/**
 * Props for MobileNav component
 */
export interface MobileNavProps {
  items?: NavItem[];
  currentPath?: string;
  onNavigate?: (href: string) => void;
  user?: {
    name: string;
    email: string;
    avatar?: string;
  };
  onLogout?: () => void;
  trigger?: React.ReactNode;
  className?: string;
}

/**
 * Mobile navigation drawer using Sheet component
 */
export function MobileNav({
  items,
  currentPath,
  onNavigate,
  user,
  onLogout,
  trigger,
  className,
}: MobileNavProps) {
  const [open, setOpen] = useState(false);

  const handleNavigate = (href: string) => {
    onNavigate?.(href);
    setOpen(false); // Close sheet after navigation
  };

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        {trigger || (
          <Button variant="ghost" size="sm" className={className}>
            <Menu className="h-5 w-5" />
            <span className="sr-only">Toggle navigation menu</span>
          </Button>
        )}
      </SheetTrigger>
      <SheetContent side="left" className="w-[280px] p-0">
        <SheetHeader className="sr-only">
          <SheetTitle>Navigation Menu</SheetTitle>
        </SheetHeader>
        <DashboardSidebar
          items={items}
          currentPath={currentPath}
          onNavigate={handleNavigate}
          user={user}
          onLogout={() => {
            onLogout?.();
            setOpen(false);
          }}
          className="h-full border-r-0"
        />
      </SheetContent>
    </Sheet>
  );
}
