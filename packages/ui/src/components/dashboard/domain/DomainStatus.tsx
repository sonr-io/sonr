'use client';

import { AnimatePresence, motion } from 'framer-motion';
import { AlertTriangle, CheckCircle, Clock, Globe, XCircle } from 'lucide-react';
import type React from 'react';
import { useEffect, useState } from 'react';
import { Badge } from '../../ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';

export type DomainStatusType = 'verified' | 'pending' | 'failed' | 'expired';

// Glass filter for glassmorphism effect
const GlassFilter: React.FC = () => (
  <svg style={{ display: 'none' }}>
    <filter
      id="glass-distortion"
      x="0%"
      y="0%"
      width="100%"
      height="100%"
      filterUnits="objectBoundingBox"
    >
      <feTurbulence
        type="fractalNoise"
        baseFrequency="0.001 0.005"
        numOctaves="1"
        seed="17"
        result="turbulence"
      />
      <feGaussianBlur in="turbulence" stdDeviation="3" result="softMap" />
      <feDisplacementMap
        in="SourceGraphic"
        in2="softMap"
        scale="20"
        xChannelSelector="R"
        yChannelSelector="G"
      />
    </filter>
  </svg>
);

// Animated pulse effect for status indicator
const PulseIndicator: React.FC<{ color: string }> = ({ color }) => {
  return (
    <span className="relative flex h-3 w-3 mr-2">
      <span
        className={`animate-ping absolute inline-flex h-full w-full rounded-full ${color} opacity-75`}
      />
      <span className={`relative inline-flex rounded-full h-3 w-3 ${color}`} />
    </span>
  );
};

interface DomainStatusProps {
  domain: string;
  status: DomainStatusType;
  verifiedAt?: Date;
  expiresAt?: Date;
  lastCheckAt?: Date;
  errorMessage?: string;
}

export function DomainStatus({
  domain,
  status,
  verifiedAt,
  expiresAt,
  lastCheckAt,
  errorMessage,
}: DomainStatusProps) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const statusConfig = {
    verified: {
      icon: CheckCircle,
      color: 'text-green-500',
      pulseColor: 'bg-green-500',
      bgColor: 'bg-green-50 dark:bg-green-950/30',
      borderGradient: 'from-green-300 via-green-500 to-emerald-500',
      label: 'Verified',
      badgeVariant: 'default' as const,
    },
    pending: {
      icon: Clock,
      color: 'text-yellow-500',
      pulseColor: 'bg-yellow-500',
      bgColor: 'bg-yellow-50 dark:bg-yellow-950/30',
      borderGradient: 'from-yellow-300 via-yellow-500 to-amber-500',
      label: 'Pending Verification',
      badgeVariant: 'secondary' as const,
    },
    failed: {
      icon: XCircle,
      color: 'text-red-500',
      pulseColor: 'bg-red-500',
      bgColor: 'bg-red-50 dark:bg-red-950/30',
      borderGradient: 'from-red-300 via-red-500 to-rose-500',
      label: 'Verification Failed',
      badgeVariant: 'destructive' as const,
    },
    expired: {
      icon: AlertTriangle,
      color: 'text-orange-500',
      pulseColor: 'bg-orange-500',
      bgColor: 'bg-orange-50 dark:bg-orange-950/30',
      borderGradient: 'from-orange-300 via-orange-500 to-amber-500',
      label: 'Expired',
      badgeVariant: 'outline' as const,
    },
  };

  const config = statusConfig[status];
  const Icon = config.icon;

  if (!mounted) return null;

  return (
    <>
      <GlassFilter />
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <Card className="relative overflow-hidden border-0 shadow-lg bg-background/80 backdrop-blur-md">
          {/* Gradient border effect */}
          <div
            className={`absolute inset-0 p-[2px] rounded-lg bg-gradient-to-r ${config.borderGradient} opacity-70`}
            style={{ filter: 'blur(0.5px)' }}
          />

          {/* Glass effect background */}
          <div
            className="absolute inset-[2px] rounded-lg bg-background/90 backdrop-blur-sm"
            style={{ filter: 'url(#glass-distortion)' }}
          />

          <div className="relative z-10">
            <CardHeader>
              <div className="flex items-start justify-between">
                <div className="flex items-start space-x-3">
                  <Globe className="h-5 w-5 text-muted-foreground mt-0.5" />
                  <div>
                    <CardTitle className="text-lg font-medium">
                      <span className="bg-clip-text text-transparent bg-gradient-to-r from-foreground to-foreground/70">
                        {domain}
                      </span>
                    </CardTitle>
                    <CardDescription>Domain verification status</CardDescription>
                  </div>
                </div>
                <Badge
                  variant={config.badgeVariant}
                  className="transition-all duration-300 hover:scale-105"
                >
                  {config.label}
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div
                className={`${config.bgColor} rounded-lg p-4 space-y-3 backdrop-blur-sm transition-all duration-300 hover:shadow-md`}
              >
                <div className="flex items-center space-x-2">
                  <PulseIndicator color={config.pulseColor} />
                  <Icon className={`h-5 w-5 ${config.color}`} />
                  <span className={`font-medium ${config.color}`}>{config.label}</span>
                </div>

                <AnimatePresence>
                  {errorMessage && status === 'failed' && (
                    <motion.div
                      initial={{ opacity: 0, height: 0 }}
                      animate={{ opacity: 1, height: 'auto' }}
                      exit={{ opacity: 0, height: 0 }}
                      className="text-sm text-red-600 bg-red-50 dark:bg-red-950/30 p-3 rounded-md"
                    >
                      <p className="font-medium">Error:</p>
                      <p>{errorMessage}</p>
                    </motion.div>
                  )}
                </AnimatePresence>

                <div className="space-y-2 text-sm text-muted-foreground">
                  {verifiedAt && (
                    <div className="flex justify-between">
                      <span>Verified at:</span>
                      <span>{verifiedAt.toLocaleDateString()}</span>
                    </div>
                  )}
                  {expiresAt && (
                    <div className="flex justify-between">
                      <span>Expires at:</span>
                      <span className={status === 'expired' ? `${config.color} font-medium` : ''}>
                        {expiresAt.toLocaleDateString()}
                      </span>
                    </div>
                  )}
                  {lastCheckAt && (
                    <div className="flex justify-between">
                      <span>Last checked:</span>
                      <span>{lastCheckAt.toLocaleString()}</span>
                    </div>
                  )}
                </div>
              </div>

              <AnimatePresence>
                {status === 'pending' && (
                  <motion.div
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -10 }}
                    className="mt-4 p-3 bg-yellow-50 dark:bg-yellow-950/30 border border-yellow-200 dark:border-yellow-800/50 rounded-lg"
                  >
                    <p className="text-sm text-yellow-800 dark:text-yellow-300">
                      DNS propagation can take up to 48 hours. We'll check periodically and notify
                      you once verification is complete.
                    </p>
                  </motion.div>
                )}

                {status === 'expired' && (
                  <motion.div
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -10 }}
                    className="mt-4 p-3 bg-orange-50 dark:bg-orange-950/30 border border-orange-200 dark:border-orange-800/50 rounded-lg"
                  >
                    <p className="text-sm text-orange-800 dark:text-orange-300">
                      Your domain verification has expired. Please re-verify to continue using this
                      domain.
                    </p>
                  </motion.div>
                )}
              </AnimatePresence>
            </CardContent>
          </div>
        </Card>
      </motion.div>
    </>
  );
}
