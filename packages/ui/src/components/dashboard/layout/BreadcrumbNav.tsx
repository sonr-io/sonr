'use client';

import { Home } from 'lucide-react';
import { cn } from '../../../lib/utils';
import {
  Breadcrumb,
  BreadcrumbEllipsis,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '../../ui/breadcrumb';

/**
 * Breadcrumb item structure
 */
export interface BreadcrumbNavItem {
  label: string;
  href?: string;
  current?: boolean;
}

/**
 * Props for BreadcrumbNav component
 */
export interface BreadcrumbNavProps {
  items: BreadcrumbNavItem[];
  onNavigate?: (href: string) => void;
  showHome?: boolean;
  homeHref?: string;
  homeLabel?: string;
  maxItems?: number;
  className?: string;
}

/**
 * Breadcrumb navigation component for dashboard
 */
export function BreadcrumbNav({
  items,
  onNavigate,
  showHome = true,
  homeHref = '/dashboard',
  homeLabel = 'Dashboard',
  maxItems = 4,
  className,
}: BreadcrumbNavProps) {
  // Handle ellipsis for long breadcrumb trails
  const displayItems =
    items.length > maxItems
      ? [
          items[0],
          { label: '...', href: undefined, current: false },
          ...items.slice(-(maxItems - 2)),
        ]
      : items;

  return (
    <Breadcrumb className={cn('mb-4', className)}>
      <BreadcrumbList>
        {showHome && (
          <>
            <BreadcrumbItem>
              {items.length > 0 ? (
                <BreadcrumbLink
                  asChild
                  className="cursor-pointer"
                  onClick={() => onNavigate?.(homeHref)}
                >
                  <a>
                    <Home className="h-4 w-4" />
                    <span className="ml-2">{homeLabel}</span>
                  </a>
                </BreadcrumbLink>
              ) : (
                <BreadcrumbPage>
                  <Home className="h-4 w-4" />
                  <span className="ml-2">{homeLabel}</span>
                </BreadcrumbPage>
              )}
            </BreadcrumbItem>
            {items.length > 0 && <BreadcrumbSeparator />}
          </>
        )}

        {displayItems.map((item, index) => {
          if (!item) return null;
          const isLast = index === displayItems.length - 1;
          const isEllipsis = item.label === '...';

          return (
            <BreadcrumbItem key={index}>
              {isEllipsis ? (
                <BreadcrumbEllipsis />
              ) : isLast || item.current ? (
                <BreadcrumbPage>{item.label}</BreadcrumbPage>
              ) : (
                <BreadcrumbLink
                  asChild
                  className="cursor-pointer"
                  onClick={() => item.href && onNavigate?.(item.href)}
                >
                  <a>{item.label}</a>
                </BreadcrumbLink>
              )}
              {!isLast && !isEllipsis && <BreadcrumbSeparator />}
            </BreadcrumbItem>
          );
        })}
      </BreadcrumbList>
    </Breadcrumb>
  );
}
