import { z } from 'zod';

/**
 * Time range presets
 */
export enum TimeRangePreset {
  LAST_HOUR = 'last_hour',
  LAST_24_HOURS = 'last_24_hours',
  LAST_7_DAYS = 'last_7_days',
  LAST_30_DAYS = 'last_30_days',
  LAST_90_DAYS = 'last_90_days',
  CUSTOM = 'custom',
}

/**
 * Metric aggregation type
 */
export enum AggregationType {
  SUM = 'sum',
  AVG = 'avg',
  MIN = 'min',
  MAX = 'max',
  COUNT = 'count',
  P50 = 'p50',
  P95 = 'p95',
  P99 = 'p99',
}

/**
 * Chart type for visualization
 */
export enum ChartType {
  LINE = 'line',
  BAR = 'bar',
  AREA = 'area',
  PIE = 'pie',
  DONUT = 'donut',
  SCATTER = 'scatter',
  HEATMAP = 'heatmap',
  METRIC = 'metric',
}

/**
 * Time range configuration
 */
export interface TimeRange {
  preset?: TimeRangePreset;
  start: string;
  end: string;
  timezone?: string;
  granularity?: 'minute' | 'hour' | 'day' | 'week' | 'month';
}

/**
 * Metric definition
 */
export interface Metric {
  id: string;
  name: string;
  value: number;
  unit?: string;
  change?: {
    value: number;
    percentage: number;
    direction: 'up' | 'down' | 'stable';
  };
  sparkline?: number[];
  timestamp: string;
}

/**
 * Time series data point
 */
export interface TimeSeriesDataPoint {
  timestamp: string;
  value: number;
  label?: string;
  metadata?: Record<string, any>;
}

/**
 * Time series dataset
 */
export interface TimeSeriesDataset {
  id: string;
  name: string;
  data: TimeSeriesDataPoint[];
  color?: string;
  aggregation?: AggregationType;
  visible?: boolean;
}

/**
 * Analytics query configuration
 */
export interface AnalyticsQuery {
  metrics: string[];
  dimensions?: string[];
  filters?: Array<{
    field: string;
    operator: 'eq' | 'neq' | 'gt' | 'lt' | 'gte' | 'lte' | 'in' | 'nin';
    value: any;
  }>;
  groupBy?: string[];
  orderBy?: Array<{
    field: string;
    direction: 'asc' | 'desc';
  }>;
  timeRange: TimeRange;
  limit?: number;
  offset?: number;
}

/**
 * Service analytics metrics
 */
export interface ServiceAnalytics {
  serviceId: string;
  timeRange: TimeRange;
  summary: {
    totalRequests: number;
    uniqueUsers: number;
    averageLatency: number;
    errorRate: number;
    successRate: number;
    bandwidth: number;
  };
  timeSeries: {
    requests: TimeSeriesDataset;
    latency: TimeSeriesDataset;
    errors: TimeSeriesDataset;
    users: TimeSeriesDataset;
  };
  breakdown: {
    byEndpoint: Array<{ endpoint: string; count: number; percentage: number }>;
    byStatus: Array<{ status: string; count: number; percentage: number }>;
    byUser: Array<{ user: string; count: number; percentage: number }>;
    byCountry: Array<{ country: string; code: string; count: number }>;
  };
}

/**
 * API usage analytics
 */
export interface ApiUsageAnalytics {
  apiKeyId?: string;
  timeRange: TimeRange;
  usage: {
    totalCalls: number;
    successfulCalls: number;
    failedCalls: number;
    quotaUsed: number;
    quotaLimit: number;
    averageResponseTime: number;
  };
  endpoints: Array<{
    path: string;
    method: string;
    calls: number;
    avgLatency: number;
    errorRate: number;
  }>;
  errors: Array<{
    code: string;
    message: string;
    count: number;
    lastOccurred: string;
  }>;
  rateLimits: {
    current: number;
    limit: number;
    resetsAt: string;
  };
}

/**
 * Performance metrics
 */
export interface PerformanceMetrics {
  serviceId: string;
  timestamp: string;
  cpu: {
    usage: number;
    cores: number;
    loadAverage: [number, number, number];
  };
  memory: {
    used: number;
    total: number;
    percentage: number;
  };
  latency: {
    p50: number;
    p95: number;
    p99: number;
    max: number;
  };
  throughput: {
    requestsPerSecond: number;
    bytesPerSecond: number;
  };
  availability: {
    uptime: number;
    downtime: number;
    percentage: number;
  };
}

/**
 * Alert configuration
 */
export interface AlertConfig {
  id: string;
  name: string;
  description?: string;
  metric: string;
  condition: {
    operator: 'gt' | 'lt' | 'gte' | 'lte' | 'eq' | 'neq';
    threshold: number;
    duration?: string;
  };
  actions: Array<{
    type: 'email' | 'webhook' | 'slack' | 'discord';
    config: Record<string, any>;
  }>;
  enabled: boolean;
  createdAt: string;
  lastTriggered?: string;
}

/**
 * Dashboard configuration
 */
export interface DashboardConfig {
  id: string;
  name: string;
  description?: string;
  widgets: Array<{
    id: string;
    type: 'chart' | 'metric' | 'table' | 'list';
    title: string;
    config: {
      chartType?: ChartType;
      metrics?: string[];
      query?: AnalyticsQuery;
      display?: Record<string, any>;
    };
    position: {
      x: number;
      y: number;
      w: number;
      h: number;
    };
  }>;
  filters?: Record<string, any>;
  refreshInterval?: number;
  createdAt: string;
  updatedAt: string;
}

/**
 * Validation schemas
 */
export const TimeRangeSchema = z.object({
  preset: z.nativeEnum(TimeRangePreset).optional(),
  start: z.string(),
  end: z.string(),
  timezone: z.string().optional(),
  granularity: z.enum(['minute', 'hour', 'day', 'week', 'month']).optional(),
});

export const MetricSchema = z.object({
  id: z.string(),
  name: z.string(),
  value: z.number(),
  unit: z.string().optional(),
  change: z
    .object({
      value: z.number(),
      percentage: z.number(),
      direction: z.enum(['up', 'down', 'stable']),
    })
    .optional(),
  sparkline: z.array(z.number()).optional(),
  timestamp: z.string(),
});

export const AnalyticsQuerySchema = z.object({
  metrics: z.array(z.string()).min(1),
  dimensions: z.array(z.string()).optional(),
  filters: z
    .array(
      z.object({
        field: z.string(),
        operator: z.enum(['eq', 'neq', 'gt', 'lt', 'gte', 'lte', 'in', 'nin']),
        value: z.any(),
      })
    )
    .optional(),
  groupBy: z.array(z.string()).optional(),
  orderBy: z
    .array(
      z.object({
        field: z.string(),
        direction: z.enum(['asc', 'desc']),
      })
    )
    .optional(),
  timeRange: TimeRangeSchema,
  limit: z.number().min(1).max(1000).optional(),
  offset: z.number().min(0).optional(),
});

export const AlertConfigSchema = z.object({
  id: z.string(),
  name: z.string().min(1).max(100),
  description: z.string().max(500).optional(),
  metric: z.string(),
  condition: z.object({
    operator: z.enum(['gt', 'lt', 'gte', 'lte', 'eq', 'neq']),
    threshold: z.number(),
    duration: z.string().optional(),
  }),
  actions: z.array(
    z.object({
      type: z.enum(['email', 'webhook', 'slack', 'discord']),
      config: z.record(z.any()),
    })
  ),
  enabled: z.boolean(),
  createdAt: z.string(),
  lastTriggered: z.string().optional(),
});

/**
 * Utility functions
 */
export function formatMetricValue(value: number, unit?: string): string {
  if (unit === 'bytes') {
    return formatBytes(value);
  }
  if (unit === 'percentage') {
    return `${value.toFixed(1)}%`;
  }
  if (unit === 'ms') {
    return `${value.toFixed(0)}ms`;
  }
  if (value >= 1000000) {
    return `${(value / 1000000).toFixed(1)}M`;
  }
  if (value >= 1000) {
    return `${(value / 1000).toFixed(1)}K`;
  }
  return value.toFixed(0);
}

export function formatBytes(bytes: number, decimals = 1): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / k ** i).toFixed(decimals)} ${sizes[i]}`;
}

export function getTimeRangeFromPreset(preset: TimeRangePreset): TimeRange {
  const end = new Date();
  const start = new Date();

  switch (preset) {
    case TimeRangePreset.LAST_HOUR:
      start.setHours(start.getHours() - 1);
      break;
    case TimeRangePreset.LAST_24_HOURS:
      start.setDate(start.getDate() - 1);
      break;
    case TimeRangePreset.LAST_7_DAYS:
      start.setDate(start.getDate() - 7);
      break;
    case TimeRangePreset.LAST_30_DAYS:
      start.setDate(start.getDate() - 30);
      break;
    case TimeRangePreset.LAST_90_DAYS:
      start.setDate(start.getDate() - 90);
      break;
    default:
      start.setDate(start.getDate() - 7);
  }

  return {
    preset,
    start: start.toISOString(),
    end: end.toISOString(),
  };
}

export function getMetricChangeDirection(
  current: number,
  previous: number
): 'up' | 'down' | 'stable' {
  if (current > previous) return 'up';
  if (current < previous) return 'down';
  return 'stable';
}

export function calculatePercentageChange(current: number, previous: number): number {
  if (previous === 0) return current > 0 ? 100 : 0;
  return ((current - previous) / previous) * 100;
}

export function aggregateTimeSeries(
  data: TimeSeriesDataPoint[],
  aggregation: AggregationType
): number {
  if (data.length === 0) return 0;

  const values = data.map((d) => d.value);

  switch (aggregation) {
    case AggregationType.SUM:
      return values.reduce((sum, val) => sum + val, 0);
    case AggregationType.AVG:
      return values.reduce((sum, val) => sum + val, 0) / values.length;
    case AggregationType.MIN:
      return Math.min(...values);
    case AggregationType.MAX:
      return Math.max(...values);
    case AggregationType.COUNT:
      return values.length;
    case AggregationType.P50:
      return percentile(values, 0.5);
    case AggregationType.P95:
      return percentile(values, 0.95);
    case AggregationType.P99:
      return percentile(values, 0.99);
    default:
      return 0;
  }
}

function percentile(values: number[], p: number): number {
  const sorted = [...values].sort((a, b) => a - b);
  const index = Math.ceil(sorted.length * p) - 1;
  return sorted[Math.max(0, index)] || 0;
}
