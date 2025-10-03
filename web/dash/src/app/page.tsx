'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@sonr.io/ui/components/ui/card';
import { Activity, Clock, Server, Users } from 'lucide-react';
import { useEffect, useState } from 'react';

export default function DashboardHome() {
  const [isLoading, setIsLoading] = useState(true);
  const [metrics, setMetrics] = useState({
    totalServices: 0,
    activeServices: 0,
    totalRequests: 0,
    avgResponseTime: 0,
  });

  useEffect(() => {
    // Simulate loading and fetching metrics
    const fetchMetrics = async () => {
      try {
        setIsLoading(true);
        await new Promise((resolve) => setTimeout(resolve, 1000));

        setMetrics({
          totalServices: 12,
          activeServices: 8,
          totalRequests: 24658,
          avgResponseTime: 142,
        });
      } catch (_err) {
        // Handle error
      } finally {
        setIsLoading(false);
      }
    };

    fetchMetrics();
  }, []);

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Sonr Developer Dashboard</h1>
          <p className="text-gray-600 mt-2">Manage your services, domains, and analytics</p>
        </div>

        {/* Metrics Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <MetricCard
            title="Total Services"
            value={isLoading ? '...' : metrics.totalServices.toString()}
            icon={<Server className="h-6 w-6" />}
            trend="+12%"
            isLoading={isLoading}
          />
          <MetricCard
            title="Active Services"
            value={isLoading ? '...' : metrics.activeServices.toString()}
            icon={<Activity className="h-6 w-6" />}
            trend="+8%"
            isLoading={isLoading}
          />
          <MetricCard
            title="Total Requests"
            value={isLoading ? '...' : metrics.totalRequests.toLocaleString()}
            icon={<Users className="h-6 w-6" />}
            trend="+23%"
            isLoading={isLoading}
          />
          <MetricCard
            title="Avg Response Time"
            value={isLoading ? '...' : `${metrics.avgResponseTime}ms`}
            icon={<Clock className="h-6 w-6" />}
            trend="-5%"
            isLoading={isLoading}
          />
        </div>

        {/* Quick Actions */}
        <Card className="mb-8">
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <ActionButton
                title="Register Service"
                description="Add a new service to your dashboard"
                href="/services?action=register"
              />
              <ActionButton
                title="Verify Domain"
                description="Verify domain ownership for your services"
                href="/domains?action=verify"
              />
              <ActionButton
                title="View Analytics"
                description="Monitor your service performance and usage"
                href="/analytics"
              />
            </div>
          </CardContent>
        </Card>

        {/* Recent Activity */}
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <ActivityItem
                title="Service registered"
                description="API Gateway service was successfully registered"
                time="2 hours ago"
                type="success"
              />
              <ActivityItem
                title="Domain verified"
                description="Domain api.example.com verification completed"
                time="5 hours ago"
                type="success"
              />
              <ActivityItem
                title="API key generated"
                description="New API key created for Data Service"
                time="1 day ago"
                type="info"
              />
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

interface MetricCardProps {
  title: string;
  value: string;
  icon: React.ReactNode;
  trend: string;
  isLoading: boolean;
}

function MetricCard({ title, value, icon, trend, isLoading }: MetricCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <div className="text-muted-foreground">{icon}</div>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-between">
          {isLoading ? (
            <div className="h-8 bg-muted rounded animate-pulse flex-1" />
          ) : (
            <div className="text-2xl font-bold">{value}</div>
          )}
          <div className="text-xs text-green-600 ml-2">{trend}</div>
        </div>
      </CardContent>
    </Card>
  );
}

interface ActionButtonProps {
  title: string;
  description: string;
  href: string;
}

function ActionButton({ title, description, href }: ActionButtonProps) {
  return (
    <a
      href={href}
      className="block p-4 border border-gray-200 rounded-lg hover:border-blue-300 hover:bg-blue-50 transition-colors"
    >
      <h3 className="font-medium text-gray-900">{title}</h3>
      <p className="text-sm text-gray-600 mt-1">{description}</p>
    </a>
  );
}

interface ActivityItemProps {
  title: string;
  description: string;
  time: string;
  type: 'success' | 'info' | 'warning';
}

function ActivityItem({ title, description, time, type }: ActivityItemProps) {
  const colors = {
    success: 'bg-green-100 text-green-800',
    info: 'bg-blue-100 text-blue-800',
    warning: 'bg-yellow-100 text-yellow-800',
  };

  return (
    <div className="flex items-start space-x-3">
      <div className={`px-2 py-1 rounded-full text-xs ${colors[type]}`}>
        {type === 'success' ? '✓' : type === 'info' ? 'i' : '!'}
      </div>
      <div className="flex-1">
        <p className="text-sm font-medium text-gray-900">{title}</p>
        <p className="text-sm text-gray-600">{description}</p>
        <p className="text-xs text-gray-400 mt-1">{time}</p>
      </div>
    </div>
  );
}
