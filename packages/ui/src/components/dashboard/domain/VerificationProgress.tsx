import { Activity, CheckCircle, Clock, XCircle } from 'lucide-react';
import { Badge } from '../../ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';
import { Progress } from '../../ui/progress';

export interface VerificationCheck {
  id: string;
  name: string;
  status: 'pending' | 'checking' | 'passed' | 'failed';
  message?: string;
  timestamp?: Date;
}

interface VerificationProgressProps {
  domain: string;
  checks: VerificationCheck[];
  overallProgress: number;
  estimatedTime?: string;
  onRetryCheck?: (checkId: string) => void;
}

export function VerificationProgress({
  domain,
  checks,
  overallProgress,
  estimatedTime,
  onRetryCheck,
}: VerificationProgressProps) {
  const passedChecks = checks.filter((c) => c.status === 'passed').length;
  const totalChecks = checks.length;

  const getStatusIcon = (status: VerificationCheck['status']) => {
    switch (status) {
      case 'passed':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-500" />;
      case 'checking':
        return <Activity className="h-4 w-4 text-blue-500 animate-pulse" />;
      default:
        return <Clock className="h-4 w-4 text-gray-400" />;
    }
  };

  const getStatusColor = (status: VerificationCheck['status']) => {
    switch (status) {
      case 'passed':
        return 'text-green-600';
      case 'failed':
        return 'text-red-600';
      case 'checking':
        return 'text-blue-600';
      default:
        return 'text-gray-500';
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Verification Progress</CardTitle>
        <CardDescription>Checking domain ownership for {domain}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {/* Overall Progress */}
          <div>
            <div className="flex justify-between items-center mb-2">
              <span className="text-sm font-medium">
                {passedChecks} of {totalChecks} checks passed
              </span>
              <span className="text-sm text-muted-foreground">{Math.round(overallProgress)}%</span>
            </div>
            <Progress value={overallProgress} className="h-3" />
            {estimatedTime && (
              <p className="text-xs text-muted-foreground mt-2">
                Estimated time remaining: {estimatedTime}
              </p>
            )}
          </div>

          {/* Individual Checks */}
          <div className="space-y-3">
            {checks.map((check) => (
              <div
                key={check.id}
                className="flex items-start justify-between p-3 border rounded-lg"
              >
                <div className="flex items-start space-x-3">
                  {getStatusIcon(check.status)}
                  <div>
                    <p className={`text-sm font-medium ${getStatusColor(check.status)}`}>
                      {check.name}
                    </p>
                    {check.message && (
                      <p className="text-xs text-muted-foreground mt-1">{check.message}</p>
                    )}
                    {check.timestamp && (
                      <p className="text-xs text-muted-foreground mt-1">
                        {check.timestamp.toLocaleTimeString()}
                      </p>
                    )}
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  {check.status === 'checking' && (
                    <Badge variant="secondary" className="animate-pulse">
                      Checking...
                    </Badge>
                  )}
                  {check.status === 'failed' && onRetryCheck && (
                    <button
                      onClick={() => onRetryCheck(check.id)}
                      className="text-xs text-blue-600 hover:text-blue-800 underline"
                    >
                      Retry
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>

          {/* Status Summary */}
          {overallProgress === 100 && (
            <div className="bg-green-50 border border-green-200 rounded-lg p-4">
              <div className="flex items-center space-x-2">
                <CheckCircle className="h-5 w-5 text-green-600" />
                <div>
                  <p className="text-sm font-medium text-green-900">Verification Complete</p>
                  <p className="text-xs text-green-700 mt-1">
                    Your domain has been successfully verified and is ready to use.
                  </p>
                </div>
              </div>
            </div>
          )}

          {checks.some((c) => c.status === 'failed') && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4">
              <div className="flex items-center space-x-2">
                <XCircle className="h-5 w-5 text-red-600" />
                <div>
                  <p className="text-sm font-medium text-red-900">Verification Issues</p>
                  <p className="text-xs text-red-700 mt-1">
                    Some checks have failed. Please review the DNS records and try again.
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
