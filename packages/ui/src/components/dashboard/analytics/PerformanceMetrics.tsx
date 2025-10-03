import {
  Activity,
  AlertTriangle,
  CheckCircle,
  Clock,
  Cpu,
  HardDrive,
  Wifi,
  XCircle,
  Zap,
} from 'lucide-react';
import { Alert, AlertDescription } from '../../ui/alert';
import { Badge } from '../../ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';
import { Progress } from '../../ui/progress';

export interface PerformanceMetric {
  name: string;
  value: number;
  unit: string;
  threshold?: {
    warning: number;
    critical: number;
  };
  status?: 'healthy' | 'warning' | 'critical';
  description?: string;
}

export interface PerformanceMetricsProps {
  title?: string;
  description?: string;
  metrics: PerformanceMetric[];
  showAlerts?: boolean;
  loading?: boolean;
}

export function PerformanceMetrics({
  title = 'Performance Metrics',
  description,
  metrics,
  showAlerts = true,
  loading = false,
}: PerformanceMetricsProps) {
  const getMetricIcon = (name: string) => {
    const iconMap: Record<string, React.ReactNode> = {
      latency: <Clock className="h-4 w-4" />,
      throughput: <Zap className="h-4 w-4" />,
      uptime: <Activity className="h-4 w-4" />,
      cpu: <Cpu className="h-4 w-4" />,
      memory: <HardDrive className="h-4 w-4" />,
      network: <Wifi className="h-4 w-4" />,
    };

    const key = name.toLowerCase();
    for (const [k, icon] of Object.entries(iconMap)) {
      if (key.includes(k)) return icon;
    }
    return <Activity className="h-4 w-4" />;
  };

  const getStatusIcon = (status?: string) => {
    switch (status) {
      case 'healthy':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'warning':
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
      case 'critical':
        return <XCircle className="h-4 w-4 text-red-500" />;
      default:
        return null;
    }
  };

  const getStatusColor = (status?: string) => {
    switch (status) {
      case 'healthy':
        return 'text-green-600';
      case 'warning':
        return 'text-yellow-600';
      case 'critical':
        return 'text-red-600';
      default:
        return 'text-muted-foreground';
    }
  };

  const getProgressColor = (status?: string) => {
    switch (status) {
      case 'healthy':
        return 'bg-green-500';
      case 'warning':
        return 'bg-yellow-500';
      case 'critical':
        return 'bg-red-500';
      default:
        return '';
    }
  };

  const calculateProgress = (metric: PerformanceMetric) => {
    if (!metric.threshold) return metric.value;

    const max = metric.threshold.critical * 1.2;
    return Math.min((metric.value / max) * 100, 100);
  };

  const criticalMetrics = metrics.filter((m) => m.status === 'critical');
  const warningMetrics = metrics.filter((m) => m.status === 'warning');

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
          {description && <CardDescription>{description}</CardDescription>}
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="space-y-2">
                <div className="h-4 w-32 bg-muted rounded animate-pulse" />
                <div className="h-2 w-full bg-muted rounded animate-pulse" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>{title}</CardTitle>
            {description && <CardDescription>{description}</CardDescription>}
          </div>
          <div className="flex gap-2">
            {criticalMetrics.length > 0 && (
              <Badge variant="destructive">{criticalMetrics.length} Critical</Badge>
            )}
            {warningMetrics.length > 0 && (
              <Badge variant="secondary" className="bg-yellow-100 text-yellow-800">
                {warningMetrics.length} Warning
              </Badge>
            )}
            {criticalMetrics.length === 0 && warningMetrics.length === 0 && (
              <Badge variant="default" className="bg-green-100 text-green-800">
                All Healthy
              </Badge>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {showAlerts && criticalMetrics.length > 0 && (
            <Alert variant="destructive">
              <AlertTriangle className="h-4 w-4" />
              <AlertDescription>
                {criticalMetrics.length} metric{criticalMetrics.length > 1 ? 's' : ''} exceeded
                critical threshold
              </AlertDescription>
            </Alert>
          )}

          <div className="space-y-4">
            {metrics.map((metric, index) => (
              <div key={index} className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    {getMetricIcon(metric.name)}
                    <span className="text-sm font-medium">{metric.name}</span>
                    {getStatusIcon(metric.status)}
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`text-sm font-bold ${getStatusColor(metric.status)}`}>
                      {metric.value} {metric.unit}
                    </span>
                    {metric.threshold && (
                      <div className="text-xs text-muted-foreground">
                        <span>W: {metric.threshold.warning}</span>
                        <span className="mx-1">|</span>
                        <span>C: {metric.threshold.critical}</span>
                      </div>
                    )}
                  </div>
                </div>

                <Progress
                  value={calculateProgress(metric)}
                  className={`h-2 ${getProgressColor(metric.status)}`}
                />

                {metric.description && (
                  <p className="text-xs text-muted-foreground">{metric.description}</p>
                )}
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
