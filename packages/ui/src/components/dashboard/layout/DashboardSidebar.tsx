'use client';

import {
  Activity,
  BarChart3,
  ChevronRight,
  Database,
  FileText,
  Globe,
  HelpCircle,
  Home,
  LogOut,
  Server,
  Settings,
  Shield,
  User,
  Users,
} from 'lucide-react';
import { useState } from 'react';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '../../ui/collapsible';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuBadge,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarRail,
  SidebarSeparator,
} from '../../ui/sidebar';

/**
 * Navigation item structure
 */
export interface NavItem {
  title: string;
  href: string;
  icon?: React.ReactNode;
  badge?: string | number;
  disabled?: boolean;
  external?: boolean;
  children?: NavItem[];
}

/**
 * Props for DashboardSidebar component
 */
export interface DashboardSidebarProps {
  items?: NavItem[];
  currentPath?: string;
  onNavigate?: (href: string) => void;
  onLogout?: () => void;
  user?: {
    name: string;
    email: string;
    avatar?: string;
  };
  className?: string;
}

// Default navigation items
const defaultItems: NavItem[] = [
  {
    title: 'Dashboard',
    href: '/dashboard',
    icon: <Home className="h-4 w-4" />,
  },
  {
    title: 'Services',
    href: '/services',
    icon: <Server className="h-4 w-4" />,
    children: [
      { title: 'All Services', href: '/services' },
      { title: 'Create Service', href: '/services/create' },
      { title: 'API Keys', href: '/services/keys' },
    ],
  },
  {
    title: 'Domains',
    href: '/domains',
    icon: <Globe className="h-4 w-4" />,
    badge: 'New',
  },
  {
    title: 'Permissions',
    href: '/permissions',
    icon: <Shield className="h-4 w-4" />,
    children: [
      { title: 'Overview', href: '/permissions' },
      { title: 'UCAN Tokens', href: '/permissions/ucan' },
      { title: 'Audit Log', href: '/permissions/audit' },
    ],
  },
  {
    title: 'Analytics',
    href: '/analytics',
    icon: <BarChart3 className="h-4 w-4" />,
  },
  {
    title: 'Storage',
    href: '/storage',
    icon: <Database className="h-4 w-4" />,
  },
  {
    title: 'Users',
    href: '/users',
    icon: <Users className="h-4 w-4" />,
  },
  {
    title: 'Activity',
    href: '/activity',
    icon: <Activity className="h-4 w-4" />,
  },
  {
    title: 'Documentation',
    href: '/docs',
    icon: <FileText className="h-4 w-4" />,
    external: true,
  },
  {
    title: 'Settings',
    href: '/settings',
    icon: <Settings className="h-4 w-4" />,
  },
];

/**
 * Sidebar navigation for dashboard layout
 */
export function DashboardSidebar({
  items = defaultItems,
  currentPath = '/dashboard',
  onNavigate,
  onLogout,
  user,
  className,
}: DashboardSidebarProps) {
  const [expandedItems, setExpandedItems] = useState<Set<string>>(new Set());

  const toggleExpanded = (href: string) => {
    const newExpanded = new Set(expandedItems);
    if (newExpanded.has(href)) {
      newExpanded.delete(href);
    } else {
      newExpanded.add(href);
    }
    setExpandedItems(newExpanded);
  };

  const handleNavigate = (href: string, external?: boolean) => {
    if (external) {
      window.open(href, '_blank');
    } else {
      onNavigate?.(href);
    }
  };

  const isActive = (href: string) => {
    return currentPath === href || currentPath.startsWith(`${href}/`);
  };

  const renderNavItem = (item: NavItem) => {
    const hasChildren = item.children && item.children.length > 0;
    const expanded = expandedItems.has(item.href);
    const active = isActive(item.href);

    if (hasChildren) {
      return (
        <Collapsible key={item.href} open={expanded} onOpenChange={() => toggleExpanded(item.href)}>
          <SidebarMenuItem>
            <CollapsibleTrigger asChild>
              <SidebarMenuButton isActive={active} disabled={item.disabled}>
                {item.icon}
                <span>{item.title}</span>
                {item.badge && <SidebarMenuBadge>{item.badge}</SidebarMenuBadge>}
                <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
              </SidebarMenuButton>
            </CollapsibleTrigger>
            <CollapsibleContent>
              <SidebarMenuSub>
                {item.children?.map((child) => (
                  <SidebarMenuSubItem key={child.href}>
                    <SidebarMenuSubButton
                      asChild
                      isActive={isActive(child.href)}
                      disabled={child.disabled}
                    >
                      <button
                        onClick={() => handleNavigate(child.href, child.external)}
                        className="w-full"
                      >
                        {child.icon}
                        <span>{child.title}</span>
                        {child.badge && <SidebarMenuBadge>{child.badge}</SidebarMenuBadge>}
                      </button>
                    </SidebarMenuSubButton>
                  </SidebarMenuSubItem>
                ))}
              </SidebarMenuSub>
            </CollapsibleContent>
          </SidebarMenuItem>
        </Collapsible>
      );
    }

    return (
      <SidebarMenuItem key={item.href}>
        <SidebarMenuButton asChild isActive={active} disabled={item.disabled}>
          <button onClick={() => handleNavigate(item.href, item.external)} className="w-full">
            {item.icon}
            <span>{item.title}</span>
            {item.badge && <SidebarMenuBadge>{item.badge}</SidebarMenuBadge>}
          </button>
        </SidebarMenuButton>
      </SidebarMenuItem>
    );
  };

  return (
    <Sidebar className={className} collapsible="icon">
      <SidebarHeader>
        <div className="flex h-12 items-center px-4">
          <h2 className="text-lg font-semibold">Sonr Services</h2>
        </div>
      </SidebarHeader>

      {user && (
        <>
          <SidebarSeparator />
          <SidebarGroup>
            <div className="flex items-center gap-3 px-4 py-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-primary-foreground">
                {user.avatar ? (
                  <img src={user.avatar} alt={user.name} className="h-8 w-8 rounded-full" />
                ) : (
                  <User className="h-4 w-4" />
                )}
              </div>
              <div className="flex-1 overflow-hidden group-data-[collapsible=icon]:hidden">
                <p className="text-sm font-medium truncate">{user.name}</p>
                <p className="text-xs text-muted-foreground truncate">{user.email}</p>
              </div>
            </div>
          </SidebarGroup>
        </>
      )}

      <SidebarSeparator />

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Navigation</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>{items.map((item) => renderNavItem(item))}</SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton>
              <HelpCircle className="h-4 w-4" />
              <span>Help & Support</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          {onLogout && (
            <SidebarMenuItem>
              <SidebarMenuButton
                onClick={onLogout}
                className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950"
              >
                <LogOut className="h-4 w-4" />
                <span>Logout</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          )}
        </SidebarMenu>
      </SidebarFooter>

      <SidebarRail />
    </Sidebar>
  );
}
