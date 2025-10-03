/**
 * Analytics & Metrics Types
 * Types for charts, metrics, and data visualization
 */

import type { Timestamp } from './common';

/**
 * Time range for analytics
 */
export interface TimeRange {
  start: Date;
  end: Date;
  preset?:
    | 'today'
    | 'yesterday'
    | 'last7days'
    | 'last30days'
    | 'thisMonth'
    | 'lastMonth'
    | 'custom';
}

/**
 * Analytics data point
 */
export interface DataPoint {
  timestamp: Timestamp;
  value: number;
  label?: string;
  metadata?: Record<string, any>;
}

/**
 * Chart configuration
 */
export interface ChartConfig {
  type: 'line' | 'area' | 'bar' | 'pie' | 'donut' | 'radar';
  colors?: string[];
  showLegend?: boolean;
  showTooltip?: boolean;
  showGrid?: boolean;
  animated?: boolean;
  stacked?: boolean;
}

/**
 * Metric with trend
 */
export interface Metric {
  label: string;
  value: number | string;
  change?: number;
  changeType?: 'increase' | 'decrease' | 'neutral';
  unit?: string;
  icon?: string;
}

/**
 * Performance metric
 */
export interface PerformanceMetric {
  name: string;
  value: number;
  unit: 'ms' | 's' | 'min' | '%' | 'rpm' | 'rps';
  status: 'good' | 'warning' | 'critical';
  threshold?: {
    warning: number;
    critical: number;
  };
}

/**
 * Activity data for charts
 */
export interface ActivityData {
  date: string;
  value: number;
  category?: string;
}

/**
 * Request pattern data
 */
export interface RequestPatternData {
  endpoint: string;
  method: string;
  count: number;
  avgDuration: number;
  errorRate: number;
}
