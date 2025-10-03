import * as React from 'react';
import { cn } from '../../lib/utils';
import { Alert, AlertDescription } from './alert';

export interface ErrorAlertProps {
  message: string;
  className?: string;
  onDismiss?: () => void;
}

const ErrorAlert = React.forwardRef<HTMLDivElement, ErrorAlertProps>(
  ({ message, className, onDismiss, ...props }, ref) => {
    return (
      <Alert ref={ref} className={cn('border-destructive', className)} {...props}>
        <AlertDescription className="flex items-center justify-between">
          <span>{message}</span>
          {onDismiss && (
            <button
              onClick={onDismiss}
              className="ml-2 text-destructive hover:text-destructive/80 focus:outline-none"
              aria-label="Dismiss error"
            >
              ✕
            </button>
          )}
        </AlertDescription>
      </Alert>
    );
  }
);

ErrorAlert.displayName = 'ErrorAlert';

export { ErrorAlert };
