'use client';

import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  Line,
  LineChart,
  XAxis,
  YAxis,
} from 'recharts';

import { Badge } from '../../ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';
import {
  type ChartConfig,
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from '../../ui/chart';

export interface ActivityData {
  date: string;
  requests?: number;
  errors?: number;
  latency?: number;
  users?: number;
  [key: string]: any;
}

export interface ActivityChartProps {
  title: string;
  description?: string;
  data: ActivityData[];
  type?: 'line' | 'area' | 'bar';
  dataKeys: {
    key: string;
    label: string;
    color?: string;
  }[];
  timeRange?: string;
  loading?: boolean;
  height?: number; // Note: height is managed by ChartContainer
  showGrid?: boolean;
  showLegend?: boolean;
}

export function ActivityChart({
  title,
  description,
  data,
  type = 'line',
  dataKeys,
  timeRange,
  loading = false,
  // height = 350, // Note: height is managed by ChartContainer
  showGrid = true,
  showLegend = true,
}: ActivityChartProps) {
  const chartConfig: ChartConfig = dataKeys.reduce((acc, { key, label, color }) => {
    acc[key] = {
      label,
      color: color || `hsl(var(--chart-${Object.keys(acc).length + 1}))`,
    };
    return acc;
  }, {} as ChartConfig);

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>{title}</CardTitle>
              {description && <CardDescription>{description}</CardDescription>}
            </div>
            {timeRange && <Badge variant="secondary">{timeRange}</Badge>}
          </div>
        </CardHeader>
        <CardContent>
          <div className="h-[350px] w-full bg-muted rounded animate-pulse" />
        </CardContent>
      </Card>
    );
  }

  const renderChart = () => {
    const commonProps = {
      data,
      margin: { top: 10, right: 10, left: 0, bottom: 0 },
    };

    switch (type) {
      case 'area':
        return (
          <AreaChart {...commonProps}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />}
            <XAxis
              dataKey="date"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tickFormatter={(value) => {
                const date = new Date(value);
                return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
              }}
            />
            <YAxis tickLine={false} axisLine={false} tickMargin={8} />
            <ChartTooltip content={<ChartTooltipContent />} />
            {dataKeys.map(({ key, color }) => (
              <Area
                key={key}
                type="monotone"
                dataKey={key}
                stroke={color || chartConfig[key]?.color}
                strokeWidth={2}
                fill={color || chartConfig[key]?.color}
                fillOpacity={0.2}
              />
            ))}
          </AreaChart>
        );

      case 'bar':
        return (
          <BarChart {...commonProps}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />}
            <XAxis
              dataKey="date"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tickFormatter={(value) => {
                const date = new Date(value);
                return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
              }}
            />
            <YAxis tickLine={false} axisLine={false} tickMargin={8} />
            <ChartTooltip content={<ChartTooltipContent />} />
            {dataKeys.map(({ key, color }) => (
              <Bar
                key={key}
                dataKey={key}
                fill={color || chartConfig[key]?.color}
                radius={[4, 4, 0, 0]}
              />
            ))}
          </BarChart>
        );

      default: // line
        return (
          <LineChart {...commonProps}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />}
            <XAxis
              dataKey="date"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tickFormatter={(value) => {
                const date = new Date(value);
                return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
              }}
            />
            <YAxis tickLine={false} axisLine={false} tickMargin={8} />
            <ChartTooltip content={<ChartTooltipContent />} />
            {dataKeys.map(({ key, color }) => (
              <Line
                key={key}
                type="monotone"
                dataKey={key}
                stroke={color || chartConfig[key]?.color}
                strokeWidth={2}
                dot={false}
              />
            ))}
          </LineChart>
        );
    }
  };

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>{title}</CardTitle>
            {description && <CardDescription>{description}</CardDescription>}
          </div>
          {timeRange && <Badge variant="secondary">{timeRange}</Badge>}
        </div>
      </CardHeader>
      <CardContent className="pb-4">
        <ChartContainer config={chartConfig} className="h-[350px] w-full">
          {renderChart()}
        </ChartContainer>
        {showLegend && <ChartLegend content={<ChartLegendContent payload={[]} />} />}
      </CardContent>
    </Card>
  );
}
