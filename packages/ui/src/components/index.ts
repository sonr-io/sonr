// OAuth Components
export { SignInWithSonr } from './SignInWithSonr';
export { SignInWithSonrModal } from './SignInWithSonrModal';

// OAuth Hooks
export { useSignInWithSonr } from '../hooks/useSignInWithSonr';
export type {
  UseSignInWithSonrOptions,
  UseSignInWithSonrReturn,
  UseSignInWithSonrState,
} from '../hooks/useSignInWithSonr';

// OAuth Utilities
export * from '../lib/oauth';

export {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from './ui/card';

export {
  Table,
  TableHeader,
  TableBody,
  TableFooter,
  TableHead,
  TableRow,
  TableCell,
  TableCaption,
} from './ui/table';

export { Badge, badgeVariants } from './ui/badge';

export { Button, buttonVariants } from './ui/button';

export { Input } from './ui/input';

export {
  Alert,
  AlertDescription,
  AlertTitle,
} from './ui/alert';

export { ErrorAlert } from './ui/error-alert';

export {
  Sheet,
  SheetTrigger,
  SheetClose,
  SheetContent,
  SheetHeader,
  SheetFooter,
  SheetTitle,
  SheetDescription,
} from './ui/sheet';

export {
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
} from './ui/tabs';

export {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogOverlay,
  DialogPortal,
  DialogTitle,
  DialogTrigger,
} from './ui/dialog';

export { Skeleton } from './ui/skeleton';

export { Checkbox } from './ui/checkbox';

export { Label } from './ui/label';

export {
  Command,
  CommandDialog,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
  CommandShortcut,
  CommandSeparator,
} from './ui/command';

export {
  useFormField,
  Form,
  FormItem,
  FormLabel,
  FormControl,
  FormDescription,
  FormMessage,
  FormField,
} from './ui/form';

// Export Progress component
export { Progress } from './ui/progress';

// Export Calendar component
export { Calendar } from './ui/calendar';

// Export Chart components
export {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  ChartLegend,
  ChartLegendContent,
  ChartStyle,
  type ChartConfig,
} from './ui/chart';

// Export Popover components
export {
  Popover,
  PopoverTrigger,
  PopoverContent,
} from './ui/popover';

// Export Select components
export {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectScrollDownButton,
  SelectScrollUpButton,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from './ui/select';

// Export DropdownMenu components
export {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuCheckboxItem,
  DropdownMenuRadioItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuGroup,
  DropdownMenuPortal,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuRadioGroup,
} from './ui/dropdown-menu';

// Export dashboard service components
// TODO: Uncomment when services components are implemented
// export {
//   ServiceList,
//   ServiceCard,
//   ServiceForm,
//   ServiceDetails,
//   ServiceSearch,
//   ServiceMetrics,
//   type Service,
// } from './dashboard/services';

// Export dashboard domain verification components
export {
  DomainVerificationFlow,
  DomainStatus,
  DNSRecordDisplay,
  VerificationProgress,
  DomainDashboard,
  DomainList,
  DomainSelector,
  DNSInstructions,
  VerificationStatus,
  VerificationWizard,
  type Domain,
  type VerificationStep,
  type DomainStatusType,
  type DNSRecord,
  type VerificationCheck,
} from './dashboard/domain';

// Export dashboard analytics components
export {
  MetricsCard,
  ActivityChart,
  RequestPatternChart,
  PerformanceMetrics,
  TimeRangeSelector,
  type MetricsCardProps,
  type ActivityData,
  type ActivityChartProps,
  type RequestPatternData,
  type RequestPatternChartProps,
  type PerformanceMetric,
  type PerformanceMetricsProps,
  type TimeRangeSelectorProps,
} from './dashboard/analytics';

// Export dashboard permissions components
export {
  PermissionGrid,
  PermissionSelector,
  UCANViewer,
  PermissionAuditLog,
  PermissionRequest,
  type Permission,
  type PermissionGridProps,
  type PermissionSelectorProps,
  type UCANCapability,
  type UCANToken,
  type UCANViewerProps,
  type AuditLogEntry,
  type PermissionAuditLogProps,
  type PermissionRequestProps,
} from './dashboard/permissions';

// Export dashboard layout components
export {
  DashboardSidebar,
  DashboardHeader,
  DashboardContent,
  MobileNav,
  BreadcrumbNav,
  ThemeToggle,
  type NavItem,
  type DashboardSidebarProps,
  type DashboardHeaderProps,
  type DashboardContentProps,
  type MobileNavProps,
  type BreadcrumbNavItem,
  type BreadcrumbNavProps,
  type ThemeToggleProps,
} from './dashboard/layout';

// Export additional UI components
export { Textarea } from './ui/textarea';
export { Toggle, toggleVariants } from './ui/toggle';
export {
  Breadcrumb,
  BreadcrumbList,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbPage,
  BreadcrumbSeparator,
  BreadcrumbEllipsis,
} from './ui/breadcrumb';

// Export Sidebar components
export {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupAction,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInput,
  SidebarInset,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuBadge,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSkeleton,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarProvider,
  SidebarRail,
  SidebarSeparator,
  SidebarTrigger,
  useSidebar,
} from './ui/sidebar';

// Export Collapsible components
export {
  Collapsible,
  CollapsibleTrigger,
  CollapsibleContent,
} from './ui/collapsible';

// Export Separator component
export { Separator } from './ui/separator';

// Export Tooltip components
export {
  Tooltip,
  TooltipTrigger,
  TooltipContent,
  TooltipProvider,
} from './ui/tooltip';

// Export utility functions
export { cn } from '../lib/utils';
