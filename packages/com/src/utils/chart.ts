/**
 * Chart & Analytics Utilities
 * Functions for data visualization and analytics
 */

import type { DataPoint } from '../types';

/**
 * Generate chart colors
 */
export function generateChartColors(count: number): string[] {
  const baseColors = [
    '#3b82f6', // blue
    '#10b981', // emerald
    '#f59e0b', // amber
    '#ef4444', // red
    '#8b5cf6', // violet
    '#ec4899', // pink
    '#14b8a6', // teal
    '#f97316', // orange
  ];

  const colors: string[] = [];
  for (let i = 0; i < count; i++) {
    colors.push(baseColors[i % baseColors.length]);
  }

  return colors;
}

/**
 * Calculate percentage change
 */
export function calculatePercentageChange(oldValue: number, newValue: number): number {
  if (oldValue === 0) return newValue > 0 ? 100 : 0;
  return ((newValue - oldValue) / Math.abs(oldValue)) * 100;
}

/**
 * Aggregate data points by interval
 */
export function aggregateDataPoints(
  points: DataPoint[],
  interval: 'hour' | 'day' | 'week' | 'month'
): DataPoint[] {
  const grouped = new Map<string, DataPoint[]>();

  points.forEach((point) => {
    const date = new Date(point.timestamp);
    let key: string;

    switch (interval) {
      case 'hour':
        key = `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}-${date.getHours()}`;
        break;
      case 'day':
        key = `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
        break;
      case 'week': {
        const weekNumber = Math.floor(date.getDate() / 7);
        key = `${date.getFullYear()}-${date.getMonth()}-W${weekNumber}`;
        break;
      }
      case 'month':
        key = `${date.getFullYear()}-${date.getMonth()}`;
        break;
    }

    if (!grouped.has(key)) {
      grouped.set(key, []);
    }
    grouped.get(key)?.push(point);
  });

  const aggregated: DataPoint[] = [];
  grouped.forEach((group, key) => {
    const sum = group.reduce((acc, point) => acc + point.value, 0);
    const avg = sum / group.length;

    aggregated.push({
      timestamp: group[0].timestamp,
      value: avg,
      label: key,
      metadata: {
        count: group.length,
        sum,
        avg,
        min: Math.min(...group.map((p) => p.value)),
        max: Math.max(...group.map((p) => p.value)),
      },
    });
  });

  return aggregated.sort(
    (a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
  );
}

/**
 * Calculate statistics for data points
 */
export function calculateStatistics(values: number[]): {
  min: number;
  max: number;
  mean: number;
  median: number;
  sum: number;
  count: number;
  stdDev: number;
} {
  const sorted = [...values].sort((a, b) => a - b);
  const sum = values.reduce((acc, val) => acc + val, 0);
  const mean = sum / values.length;

  const median =
    values.length % 2 === 0
      ? (sorted[values.length / 2 - 1] + sorted[values.length / 2]) / 2
      : sorted[Math.floor(values.length / 2)];

  const variance = values.reduce((acc, val) => acc + (val - mean) ** 2, 0) / values.length;
  const stdDev = Math.sqrt(variance);

  return {
    min: Math.min(...values),
    max: Math.max(...values),
    mean,
    median,
    sum,
    count: values.length,
    stdDev,
  };
}

/**
 * Generate trend line data
 */
export function generateTrendLine(points: DataPoint[]): DataPoint[] {
  if (points.length < 2) return points;

  // Simple linear regression
  const n = points.length;
  const sumX = points.reduce((acc, _, i) => acc + i, 0);
  const sumY = points.reduce((acc, p) => acc + p.value, 0);
  const sumXY = points.reduce((acc, p, i) => acc + i * p.value, 0);
  const sumX2 = points.reduce((acc, _, i) => acc + i * i, 0);

  const slope = (n * sumXY - sumX * sumY) / (n * sumX2 - sumX * sumX);
  const intercept = (sumY - slope * sumX) / n;

  return points.map((point, i) => ({
    ...point,
    value: slope * i + intercept,
    metadata: {
      ...point.metadata,
      isTrendLine: true,
    },
  }));
}

/**
 * Calculate moving average
 */
export function calculateMovingAverage(values: number[], window: number): number[] {
  if (window > values.length) return values;

  const result: number[] = [];
  for (let i = 0; i < values.length; i++) {
    const start = Math.max(0, i - window + 1);
    const windowValues = values.slice(start, i + 1);
    const avg = windowValues.reduce((acc, val) => acc + val, 0) / windowValues.length;
    result.push(avg);
  }

  return result;
}
