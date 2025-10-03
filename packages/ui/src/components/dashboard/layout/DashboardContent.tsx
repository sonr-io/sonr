import type React from 'react';

export interface DashboardContentProps {
  children: React.ReactNode;
  className?: string;
}

/**
 * Dashboard Content Container
 * Provides consistent styling and layout for dashboard page content
 */
export function DashboardContent({ children, className }: DashboardContentProps) {
  return <main className={`flex-1 space-y-4 p-8 pt-6 ${className || ''}`}>{children}</main>;
}
