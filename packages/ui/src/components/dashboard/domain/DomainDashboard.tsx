'use client';

import { AnimatePresence, motion } from 'framer-motion';
import {
  Activity,
  AlertTriangle,
  BarChart3,
  CheckCircle,
  ChevronRight,
  Clock,
  Filter,
  Globe,
  Lock,
  Plus,
  RefreshCw,
  Search,
  Server,
  Shield,
  Sparkles,
  XCircle,
  Zap,
} from 'lucide-react';
import { useState } from 'react';
import { cn } from '../../../lib/utils';
import { Badge } from '../../ui/badge';
import { Button } from '../../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../ui/card';
import { Input } from '../../ui/input';
import { Progress } from '../../ui/progress';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../../ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../../ui/tabs';

// Types
interface Domain {
  id: string;
  name: string;
  status: 'active' | 'pending' | 'expired' | 'failed';
  verifiedAt?: Date;
  expiresAt?: Date;
  dnsRecords: DNSRecord[];
  ssl: {
    enabled: boolean;
    issuer?: string;
    expiresAt?: Date;
  };
  analytics: {
    requests: number;
    bandwidth: number;
    uptime: number;
  };
}

interface DNSRecord {
  type: 'A' | 'AAAA' | 'CNAME' | 'MX' | 'TXT' | 'NS';
  name: string;
  value: string;
  ttl: number;
  priority?: number;
}

// Mock data
const mockDomains: Domain[] = [
  {
    id: '1',
    name: 'app.sonr.io',
    status: 'active',
    verifiedAt: new Date('2024-01-15'),
    expiresAt: new Date('2025-01-15'),
    dnsRecords: [
      { type: 'A', name: '@', value: '192.168.1.1', ttl: 3600 },
      { type: 'CNAME', name: 'www', value: 'app.sonr.io', ttl: 3600 },
      { type: 'TXT', name: '_verification', value: 'sonr-verify-abc123', ttl: 300 },
    ],
    ssl: {
      enabled: true,
      issuer: "Let's Encrypt",
      expiresAt: new Date('2024-12-31'),
    },
    analytics: {
      requests: 15234,
      bandwidth: 2.4,
      uptime: 99.9,
    },
  },
  {
    id: '2',
    name: 'api.sonr.io',
    status: 'active',
    verifiedAt: new Date('2024-02-01'),
    expiresAt: new Date('2025-02-01'),
    dnsRecords: [{ type: 'A', name: '@', value: '192.168.1.2', ttl: 3600 }],
    ssl: {
      enabled: true,
      issuer: "Let's Encrypt",
      expiresAt: new Date('2024-12-31'),
    },
    analytics: {
      requests: 45678,
      bandwidth: 8.7,
      uptime: 99.99,
    },
  },
  {
    id: '3',
    name: 'docs.sonr.io',
    status: 'pending',
    dnsRecords: [{ type: 'TXT', name: '_verification', value: 'sonr-verify-xyz789', ttl: 300 }],
    ssl: {
      enabled: false,
    },
    analytics: {
      requests: 0,
      bandwidth: 0,
      uptime: 0,
    },
  },
];

export function DomainDashboard() {
  const [domains] = useState<Domain[]>(mockDomains);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTab, setSelectedTab] = useState('all');

  const filteredDomains = domains.filter((domain) => {
    if (selectedTab !== 'all' && domain.status !== selectedTab) return false;
    if (searchQuery && !domain.name.includes(searchQuery)) return false;
    return true;
  });

  const stats = {
    total: domains.length,
    active: domains.filter((d) => d.status === 'active').length,
    pending: domains.filter((d) => d.status === 'pending').length,
    expired: domains.filter((d) => d.status === 'expired').length,
  };

  const getStatusIcon = (status: Domain['status']) => {
    switch (status) {
      case 'active':
        return <CheckCircle className="h-4 w-4 text-emerald-500" />;
      case 'pending':
        return <Clock className="h-4 w-4 text-amber-500" />;
      case 'expired':
        return <AlertTriangle className="h-4 w-4 text-red-500" />;
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-600" />;
    }
  };

  const getStatusColor = (status: Domain['status']) => {
    switch (status) {
      case 'active':
        return 'bg-emerald-500/10 text-emerald-600 border-emerald-500/20';
      case 'pending':
        return 'bg-amber-500/10 text-amber-600 border-amber-500/20';
      case 'expired':
        return 'bg-red-500/10 text-red-600 border-red-500/20';
      case 'failed':
        return 'bg-red-600/10 text-red-700 border-red-600/20';
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-white to-slate-50 dark:from-slate-950 dark:via-slate-900 dark:to-slate-950 p-6 space-y-8">
      {/* Header with Glassmorphism Effect */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 p-[2px]">
        <div className="relative backdrop-blur-xl bg-white/90 dark:bg-slate-900/90 rounded-2xl p-8">
          <div className="absolute inset-0 bg-grid-slate-100/50 dark:bg-grid-slate-800/50 [mask-image:radial-gradient(ellipse_at_center,transparent_20%,black)]" />
          <div className="relative">
            <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-6">
              <div>
                <h1 className="text-4xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 dark:from-indigo-400 dark:to-purple-400 bg-clip-text text-transparent">
                  Domain Management
                </h1>
                <p className="text-slate-600 dark:text-slate-400 mt-2">
                  Manage your domains, DNS records, and SSL certificates
                </p>
              </div>
              <Button className="bg-gradient-to-r from-indigo-500 to-purple-500 hover:from-indigo-600 hover:to-purple-600 text-white shadow-lg shadow-purple-500/25">
                <Plus className="mr-2 h-4 w-4" />
                Add Domain
              </Button>
            </div>

            {/* Stats Cards */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mt-8">
              {[
                {
                  label: 'Total Domains',
                  value: stats.total,
                  icon: Globe,
                  color: 'from-blue-500 to-indigo-500',
                },
                {
                  label: 'Active',
                  value: stats.active,
                  icon: CheckCircle,
                  color: 'from-emerald-500 to-green-500',
                },
                {
                  label: 'Pending',
                  value: stats.pending,
                  icon: Clock,
                  color: 'from-amber-500 to-orange-500',
                },
                {
                  label: 'Expired',
                  value: stats.expired,
                  icon: AlertTriangle,
                  color: 'from-red-500 to-pink-500',
                },
              ].map((stat, index) => (
                <motion.div
                  key={stat.label}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: index * 0.1 }}
                >
                  <Card className="backdrop-blur-sm bg-white/50 dark:bg-slate-800/50 border-white/20 dark:border-slate-700/50">
                    <CardContent className="p-6">
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="text-sm text-slate-600 dark:text-slate-400">{stat.label}</p>
                          <p
                            className={`text-3xl font-bold mt-2 bg-gradient-to-r ${stat.color} bg-clip-text text-transparent`}
                          >
                            {stat.value}
                          </p>
                        </div>
                        <div className={`rounded-full p-3 bg-gradient-to-r ${stat.color}`}>
                          <stat.icon className="h-6 w-6 text-white" />
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </motion.div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Search and Filters */}
      <div className="flex flex-col md:flex-row gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
          <Input
            placeholder="Search domains..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10 bg-white/70 dark:bg-slate-800/70 backdrop-blur-sm border-slate-200/50 dark:border-slate-700/50"
          />
        </div>
        <Select defaultValue="all">
          <SelectTrigger className="w-[180px] bg-white/70 dark:bg-slate-800/70 backdrop-blur-sm border-slate-200/50 dark:border-slate-700/50">
            <SelectValue placeholder="Filter by status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Domains</SelectItem>
            <SelectItem value="active">Active</SelectItem>
            <SelectItem value="pending">Pending</SelectItem>
            <SelectItem value="expired">Expired</SelectItem>
          </SelectContent>
        </Select>
        <Button
          variant="outline"
          className="bg-white/70 dark:bg-slate-800/70 backdrop-blur-sm border-slate-200/50 dark:border-slate-700/50"
        >
          <Filter className="mr-2 h-4 w-4" />
          More Filters
        </Button>
      </div>

      {/* Domains Grid with Tabs */}
      <Tabs value={selectedTab} onValueChange={setSelectedTab} className="space-y-6">
        <TabsList className="bg-white/70 dark:bg-slate-800/70 backdrop-blur-sm border-slate-200/50 dark:border-slate-700/50">
          <TabsTrigger value="all">All Domains</TabsTrigger>
          <TabsTrigger value="active">Active</TabsTrigger>
          <TabsTrigger value="pending">Pending</TabsTrigger>
          <TabsTrigger value="expired">Expired</TabsTrigger>
        </TabsList>

        <TabsContent value={selectedTab} className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <AnimatePresence mode="popLayout">
              {filteredDomains.map((domain, index) => (
                <motion.div
                  key={domain.id}
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.95 }}
                  transition={{ delay: index * 0.05 }}
                  className="group"
                >
                  <Card className="relative overflow-hidden backdrop-blur-sm bg-white/70 dark:bg-slate-800/70 border-slate-200/50 dark:border-slate-700/50 hover:shadow-xl hover:shadow-purple-500/10 transition-all duration-300 cursor-pointer">
                    {/* Animated Background Gradient */}
                    <div className="absolute inset-0 bg-gradient-to-br from-indigo-500/5 via-purple-500/5 to-pink-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-500" />

                    <CardHeader>
                      <div className="flex items-start justify-between">
                        <div className="space-y-1">
                          <div className="flex items-center gap-2">
                            <Globe className="h-5 w-5 text-indigo-500" />
                            <CardTitle className="text-lg font-semibold">{domain.name}</CardTitle>
                          </div>
                          <CardDescription>
                            {domain.verifiedAt
                              ? `Verified ${domain.verifiedAt.toLocaleDateString()}`
                              : 'Pending verification'}
                          </CardDescription>
                        </div>
                        <Badge className={cn('border', getStatusColor(domain.status))}>
                          <span className="flex items-center gap-1">
                            {getStatusIcon(domain.status)}
                            {domain.status}
                          </span>
                        </Badge>
                      </div>
                    </CardHeader>

                    <CardContent className="space-y-4">
                      {/* SSL Status */}
                      <div className="flex items-center justify-between p-3 rounded-lg bg-slate-50 dark:bg-slate-900/50">
                        <div className="flex items-center gap-2">
                          {domain.ssl.enabled ? (
                            <>
                              <Lock className="h-4 w-4 text-emerald-500" />
                              <span className="text-sm text-emerald-600 dark:text-emerald-400">
                                SSL Active
                              </span>
                            </>
                          ) : (
                            <>
                              <Lock className="h-4 w-4 text-slate-400" />
                              <span className="text-sm text-slate-500">SSL Inactive</span>
                            </>
                          )}
                        </div>
                        {domain.ssl.issuer && (
                          <span className="text-xs text-slate-500">{domain.ssl.issuer}</span>
                        )}
                      </div>

                      {/* DNS Records Count */}
                      <div className="flex items-center justify-between p-3 rounded-lg bg-slate-50 dark:bg-slate-900/50">
                        <div className="flex items-center gap-2">
                          <Server className="h-4 w-4 text-blue-500" />
                          <span className="text-sm text-slate-600 dark:text-slate-400">
                            DNS Records
                          </span>
                        </div>
                        <span className="text-sm font-medium">{domain.dnsRecords.length}</span>
                      </div>

                      {/* Analytics Preview */}
                      {domain.status === 'active' && (
                        <div className="space-y-2">
                          <div className="flex items-center justify-between text-sm">
                            <span className="text-slate-500">Uptime</span>
                            <span className="font-medium text-emerald-600">
                              {domain.analytics.uptime}%
                            </span>
                          </div>
                          <Progress value={domain.analytics.uptime} className="h-1.5" />

                          <div className="grid grid-cols-2 gap-2 mt-3">
                            <div className="flex items-center gap-1">
                              <Activity className="h-3 w-3 text-indigo-500" />
                              <span className="text-xs text-slate-500">
                                {domain.analytics.requests.toLocaleString()} reqs
                              </span>
                            </div>
                            <div className="flex items-center gap-1">
                              <Zap className="h-3 w-3 text-purple-500" />
                              <span className="text-xs text-slate-500">
                                {domain.analytics.bandwidth} GB
                              </span>
                            </div>
                          </div>
                        </div>
                      )}

                      {/* Pending Status Message */}
                      {domain.status === 'pending' && (
                        <div className="flex items-center gap-2 p-3 rounded-lg bg-amber-50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-800">
                          <RefreshCw className="h-4 w-4 text-amber-600 animate-spin" />
                          <span className="text-xs text-amber-700 dark:text-amber-400">
                            DNS propagation in progress...
                          </span>
                        </div>
                      )}

                      {/* Action Button */}
                      <Button variant="ghost" className="w-full group/button">
                        <span>Manage Domain</span>
                        <ChevronRight className="ml-2 h-4 w-4 transition-transform group-hover/button:translate-x-1" />
                      </Button>
                    </CardContent>

                    {/* Animated Border Gradient */}
                    <div className="absolute inset-x-0 bottom-0 h-[2px] bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 transform scale-x-0 group-hover:scale-x-100 transition-transform duration-500" />
                  </Card>
                </motion.div>
              ))}
            </AnimatePresence>
          </div>

          {/* Empty State */}
          {filteredDomains.length === 0 && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="text-center py-12"
            >
              <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gradient-to-br from-indigo-500/20 to-purple-500/20 mb-4">
                <Globe className="h-8 w-8 text-indigo-500" />
              </div>
              <h3 className="text-lg font-semibold text-slate-900 dark:text-slate-100 mb-2">
                No domains found
              </h3>
              <p className="text-sm text-slate-500 dark:text-slate-400 mb-6">
                {searchQuery
                  ? 'Try adjusting your search criteria'
                  : 'Get started by adding your first domain'}
              </p>
              <Button className="bg-gradient-to-r from-indigo-500 to-purple-500 hover:from-indigo-600 hover:to-purple-600 text-white">
                <Plus className="mr-2 h-4 w-4" />
                Add Your First Domain
              </Button>
            </motion.div>
          )}
        </TabsContent>
      </Tabs>

      {/* Quick Actions Card */}
      <Card className="backdrop-blur-sm bg-gradient-to-br from-indigo-500/5 via-purple-500/5 to-pink-500/5 border-purple-200/20 dark:border-purple-800/20">
        <CardHeader>
          <div className="flex items-center gap-2">
            <Sparkles className="h-5 w-5 text-purple-500" />
            <CardTitle>Quick Actions</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Button
              variant="outline"
              className="justify-start bg-white/50 dark:bg-slate-800/50 backdrop-blur-sm"
            >
              <Shield className="mr-2 h-4 w-4 text-blue-500" />
              Verify Domain Ownership
            </Button>
            <Button
              variant="outline"
              className="justify-start bg-white/50 dark:bg-slate-800/50 backdrop-blur-sm"
            >
              <Server className="mr-2 h-4 w-4 text-purple-500" />
              Configure DNS Records
            </Button>
            <Button
              variant="outline"
              className="justify-start bg-white/50 dark:bg-slate-800/50 backdrop-blur-sm"
            >
              <BarChart3 className="mr-2 h-4 w-4 text-indigo-500" />
              View Analytics Dashboard
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
