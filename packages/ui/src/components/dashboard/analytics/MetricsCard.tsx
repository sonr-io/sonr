import { Activity, Minus, TrendingDown, TrendingUp } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';

export interface MetricsCardProps {
  title: string;
  value: string | number;
  description?: string;
  trend?: {
    value: number;
    direction: 'up' | 'down' | 'neutral';
    period?: string;
  };
  icon?: React.ReactNode;
  variant?: 'default' | 'primary' | 'success' | 'warning' | 'danger';
  loading?: boolean;
}

export function MetricsCard({
  title,
  value,
  description,
  trend,
  icon = <Activity className="h-4 w-4" />,
  variant = 'default',
  loading = false,
}: MetricsCardProps) {
  const getTrendIcon = () => {
    if (!trend) return null;

    switch (trend.direction) {
      case 'up':
        return <TrendingUp className="h-4 w-4" />;
      case 'down':
        return <TrendingDown className="h-4 w-4" />;
      default:
        return <Minus className="h-4 w-4" />;
    }
  };

  const getTrendColor = () => {
    if (!trend) return 'text-muted-foreground';

    if (trend.direction === 'up') {
      return trend.value >= 0 ? 'text-green-600' : 'text-red-600';
    }
    if (trend.direction === 'down') {
      return trend.value < 0 ? 'text-green-600' : 'text-red-600';
    }
    return 'text-muted-foreground';
  };

  const getVariantStyles = () => {
    switch (variant) {
      case 'primary':
        return 'border-primary/20 bg-primary/5';
      case 'success':
        return 'border-green-500/20 bg-green-50 dark:bg-green-950/20';
      case 'warning':
        return 'border-yellow-500/20 bg-yellow-50 dark:bg-yellow-950/20';
      case 'danger':
        return 'border-red-500/20 bg-red-50 dark:bg-red-950/20';
      default:
        return '';
    }
  };

  if (loading) {
    return (
      <Card className={getVariantStyles()}>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">
            <div className="h-4 w-24 bg-muted rounded animate-pulse" />
          </CardTitle>
          <div className="h-4 w-4 bg-muted rounded animate-pulse" />
        </CardHeader>
        <CardContent>
          <div className="h-8 w-32 bg-muted rounded animate-pulse mb-2" />
          <div className="h-3 w-20 bg-muted rounded animate-pulse" />
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={getVariantStyles()}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <div className="text-muted-foreground">{icon}</div>
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        {description && <CardDescription className="text-xs mt-1">{description}</CardDescription>}
        {trend && (
          <div className={`flex items-center gap-1 mt-2 text-xs ${getTrendColor()}`}>
            {getTrendIcon()}
            <span className="font-medium">{Math.abs(trend.value)}%</span>
            {trend.period && <span className="text-muted-foreground">vs {trend.period}</span>}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
