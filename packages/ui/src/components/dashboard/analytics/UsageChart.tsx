import { Area, AreaChart, ResponsiveContainer, XAxis, YAxis } from 'recharts';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '../../ui/chart';

type TimeRange = '24h' | '7d' | '30d' | '90d';

interface UsageChartProps {
  timeRange: TimeRange;
  height?: number;
}

const chartConfig = {
  usage: {
    label: 'Usage',
    color: 'hsl(var(--primary))',
  },
};

export function UsageChart({ timeRange, height = 300 }: UsageChartProps) {
  // Generate mock data based on time range
  const generateUsageData = () => {
    const days = timeRange === '24h' ? 24 : timeRange === '7d' ? 7 : timeRange === '30d' ? 30 : 90;
    const points = Math.min(days, 30); // Limit to 30 data points for readability

    return Array.from({ length: points }, (_, i) => {
      const date = new Date();
      date.setDate(date.getDate() - (points - 1 - i));

      return {
        date:
          timeRange === '24h'
            ? date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
            : date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
        usage: Math.floor(Math.random() * 1000) + 200,
        previousUsage: Math.floor(Math.random() * 800) + 150,
      };
    });
  };

  const data = generateUsageData();

  return (
    <Card>
      <CardHeader>
        <CardTitle>Usage Trends</CardTitle>
        <CardDescription>
          Historical usage patterns and growth metrics over the selected time period
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig}>
          <ResponsiveContainer width="100%" height={height}>
            <AreaChart data={data}>
              <defs>
                <linearGradient id="colorUsage" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis
                dataKey="date"
                axisLine={false}
                tickLine={false}
                tick={{ fontSize: 12 }}
                className="text-muted-foreground"
              />
              <YAxis
                axisLine={false}
                tickLine={false}
                tick={{ fontSize: 12 }}
                className="text-muted-foreground"
              />
              <ChartTooltip content={<ChartTooltipContent />} />
              <Area
                type="monotone"
                dataKey="usage"
                stroke="hsl(var(--primary))"
                fillOpacity={1}
                fill="url(#colorUsage)"
                strokeWidth={2}
              />
            </AreaChart>
          </ResponsiveContainer>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}

export type { UsageChartProps };
