'use client';

import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  Pie,
  PieChart,
  PolarAngleAxis,
  PolarGrid,
  PolarRadiusAxis,
  Radar,
  RadarChart,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from '../../ui/chart';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../../ui/select';

export interface RequestPatternData {
  endpoint?: string;
  method?: string;
  status?: string;
  count: number;
  percentage?: number;
  avgLatency?: number;
  errorRate?: number;
  [key: string]: any;
}

export interface RequestPatternChartProps {
  title: string;
  description?: string;
  data: RequestPatternData[];
  type?: 'pie' | 'radar' | 'bar';
  groupBy?: 'endpoint' | 'method' | 'status';
  loading?: boolean;
  height?: number;
  showLegend?: boolean;
  onGroupByChange?: (value: string) => void;
}

const COLORS = [
  'hsl(var(--chart-1))',
  'hsl(var(--chart-2))',
  'hsl(var(--chart-3))',
  'hsl(var(--chart-4))',
  'hsl(var(--chart-5))',
];

export function RequestPatternChart({
  title,
  description,
  data,
  type = 'pie',
  groupBy = 'endpoint',
  loading = false,
  // height = 350, // Note: height is managed by ChartContainer
  showLegend = true,
  onGroupByChange,
}: RequestPatternChartProps) {
  if (loading) {
    return (
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>{title}</CardTitle>
              {description && <CardDescription>{description}</CardDescription>}
            </div>
            {onGroupByChange && (
              <Select value={groupBy} onValueChange={onGroupByChange}>
                <SelectTrigger className="w-32">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="endpoint">By Endpoint</SelectItem>
                  <SelectItem value="method">By Method</SelectItem>
                  <SelectItem value="status">By Status</SelectItem>
                </SelectContent>
              </Select>
            )}
          </div>
        </CardHeader>
        <CardContent>
          <div className="h-[350px] w-full bg-muted rounded animate-pulse" />
        </CardContent>
      </Card>
    );
  }

  const chartConfig: ChartConfig = data.reduce((acc, item, index) => {
    const key = item[groupBy] || `item-${index}`;
    acc[key] = {
      label: key,
      color: COLORS[index % COLORS.length],
    };
    return acc;
  }, {} as ChartConfig);

  const renderChart = () => {
    switch (type) {
      case 'radar':
        return (
          <RadarChart data={data}>
            <PolarGrid className="stroke-muted" />
            <PolarAngleAxis dataKey={groupBy} />
            <PolarRadiusAxis angle={90} domain={[0, 'dataMax']} />
            <Radar
              name="Requests"
              dataKey="count"
              stroke="hsl(var(--primary))"
              fill="hsl(var(--primary))"
              fillOpacity={0.6}
            />
            {data[0]?.avgLatency && (
              <Radar
                name="Avg Latency"
                dataKey="avgLatency"
                stroke="hsl(var(--chart-2))"
                fill="hsl(var(--chart-2))"
                fillOpacity={0.6}
              />
            )}
            <Tooltip />
            <Legend />
          </RadarChart>
        );

      case 'bar':
        return (
          <BarChart data={data}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
            <XAxis
              dataKey={groupBy}
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              angle={-45}
              textAnchor="end"
              height={80}
            />
            <YAxis tickLine={false} axisLine={false} tickMargin={8} />
            <ChartTooltip content={<ChartTooltipContent />} />
            <Bar dataKey="count" fill="hsl(var(--primary))" radius={[4, 4, 0, 0]}>
              {data.map((_entry, index) => (
                <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
              ))}
            </Bar>
            {data[0]?.errorRate !== undefined && (
              <Bar dataKey="errorRate" fill="hsl(var(--destructive))" radius={[4, 4, 0, 0]} />
            )}
          </BarChart>
        );

      default: // pie
        return (
          <PieChart>
            <Pie
              data={data}
              cx="50%"
              cy="50%"
              labelLine={false}
              label={({ percentage }) => `${(percentage * 100).toFixed(0)}%`}
              outerRadius={120}
              fill="#8884d8"
              dataKey="count"
              nameKey={groupBy}
            >
              {data.map((_entry, index) => (
                <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
              ))}
            </Pie>
            <Tooltip
              formatter={(value: number) => value.toLocaleString()}
              labelFormatter={(label) => `${groupBy}: ${label}`}
            />
            {showLegend && <Legend />}
          </PieChart>
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
          {onGroupByChange && (
            <Select value={groupBy} onValueChange={onGroupByChange}>
              <SelectTrigger className="w-32">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="endpoint">By Endpoint</SelectItem>
                <SelectItem value="method">By Method</SelectItem>
                <SelectItem value="status">By Status</SelectItem>
              </SelectContent>
            </Select>
          )}
        </div>
      </CardHeader>
      <CardContent className="pb-4">
        <ChartContainer config={chartConfig} className="h-[350px] w-full">
          {renderChart()}
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
