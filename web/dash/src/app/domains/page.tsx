'use client';

import type { Domain, DomainVerificationStatus } from '@sonr.io/com/types/domain';
import {
  DNSInstructions,
  DomainList,
  DomainSelector,
  VerificationStatus,
  VerificationWizard,
} from '@sonr.io/ui';
import { DashboardContent, DashboardHeader } from '@sonr.io/ui';
import { Alert, AlertDescription } from '@sonr.io/ui/components/ui/alert';
import { Badge } from '@sonr.io/ui/components/ui/badge';
import { Button } from '@sonr.io/ui/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@sonr.io/ui/components/ui/card';
import { Dialog, DialogContent, DialogTrigger } from '@sonr.io/ui/components/ui/dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@sonr.io/ui/components/ui/tabs';
import { Globe, Plus, RefreshCw, Shield } from 'lucide-react';
import React, { useEffect, useState } from 'react';

/**
 * Domain Management Page
 * Handles domain verification flow, status tracking, and DNS instructions
 */
export default function DomainsPage() {
  const [domains, setDomains] = useState<Domain[]>([]);
  const [selectedDomain, setSelectedDomain] = useState<string | null>(null);
  const [_isLoading, setIsLoading] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [showVerificationWizard, setShowVerificationWizard] = useState(false);
  const [verificationStatus, setVerificationStatus] = useState<DomainVerificationStatus | null>(
    null
  );

  // Fetch domains on mount
  useEffect(() => {
    fetchDomains();
  }, []);

  const fetchDomains = async () => {
    setIsLoading(true);
    try {
      // TODO: Replace with actual API call
      const mockDomains: Domain[] = [
        {
          id: '1',
          name: 'example.com',
          status: 'verified',
          verifiedAt: new Date('2024-01-15'),
          txtRecord: 'sonr-verify=abc123def456',
          serviceCount: 3,
        },
        {
          id: '2',
          name: 'api.example.com',
          status: 'pending',
          verifiedAt: null,
          txtRecord: 'sonr-verify=xyz789ghi012',
          serviceCount: 0,
        },
      ];
      setDomains(mockDomains);
    } catch (error) {
      console.error('Failed to fetch domains:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const refreshVerificationStatus = async () => {
    setIsRefreshing(true);
    try {
      // TODO: Replace with actual API call to check DNS records
      await new Promise((resolve) => setTimeout(resolve, 2000));

      // Mock status update
      const updatedStatus: DomainVerificationStatus = {
        isVerified: false,
        lastChecked: new Date(),
        dnsRecordsFound: true,
        expectedRecord: 'sonr-verify=abc123def456',
      };
      setVerificationStatus(updatedStatus);

      // Refresh domain list
      await fetchDomains();
    } catch (error) {
      console.error('Failed to refresh verification status:', error);
    } finally {
      setIsRefreshing(false);
    }
  };

  const handleDomainVerification = async (domain: string) => {
    try {
      // TODO: Replace with actual API call
      console.log('Starting verification for:', domain);
      setShowVerificationWizard(false);
      await fetchDomains();
    } catch (error) {
      console.error('Failed to verify domain:', error);
    }
  };

  const handleDomainSelection = (domainId: string) => {
    setSelectedDomain(domainId);
    const domain = domains.find((d) => d.id === domainId);
    if (domain && domain.status === 'pending') {
      refreshVerificationStatus();
    }
  };

  return (
    <div className="flex flex-col gap-6">
      <DashboardHeader
        title="Domain Management"
        description="Verify and manage your domains for service registration"
      >
        <Dialog open={showVerificationWizard} onOpenChange={setShowVerificationWizard}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Add Domain
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-3xl">
            <VerificationWizard
              onComplete={handleDomainVerification}
              onCancel={() => setShowVerificationWizard(false)}
            />
          </DialogContent>
        </Dialog>
      </DashboardHeader>

      <DashboardContent>
        <div className="grid gap-6">
          {/* Domain Overview Cards */}
          <div className="grid gap-4 md:grid-cols-3">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Domains</CardTitle>
                <Globe className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{domains.length}</div>
                <p className="text-xs text-muted-foreground">
                  {domains.filter((d) => d.status === 'verified').length} verified
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Pending Verification</CardTitle>
                <RefreshCw className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {domains.filter((d) => d.status === 'pending').length}
                </div>
                <p className="text-xs text-muted-foreground">Awaiting DNS verification</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Active Services</CardTitle>
                <Shield className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {domains.reduce((sum, d) => sum + (d.serviceCount || 0), 0)}
                </div>
                <p className="text-xs text-muted-foreground">Across all domains</p>
              </CardContent>
            </Card>
          </div>

          {/* Domain Management Tabs */}
          <Card>
            <CardHeader>
              <CardTitle>Your Domains</CardTitle>
              <CardDescription>
                Manage your verified domains and track verification status
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Tabs defaultValue="all" className="space-y-4">
                <TabsList>
                  <TabsTrigger value="all">All Domains</TabsTrigger>
                  <TabsTrigger value="verified">Verified</TabsTrigger>
                  <TabsTrigger value="pending">Pending</TabsTrigger>
                </TabsList>

                <TabsContent value="all" className="space-y-4">
                  <DomainList
                    domains={domains}
                    onDomainSelect={handleDomainSelection}
                    onRefresh={refreshVerificationStatus}
                    isRefreshing={isRefreshing}
                    selectedDomainId={selectedDomain}
                  />
                </TabsContent>

                <TabsContent value="verified" className="space-y-4">
                  <DomainList
                    domains={domains.filter((d) => d.status === 'verified')}
                    onDomainSelect={handleDomainSelection}
                    onRefresh={refreshVerificationStatus}
                    isRefreshing={isRefreshing}
                    selectedDomainId={selectedDomain}
                  />
                </TabsContent>

                <TabsContent value="pending" className="space-y-4">
                  <DomainList
                    domains={domains.filter((d) => d.status === 'pending')}
                    onDomainSelect={handleDomainSelection}
                    onRefresh={refreshVerificationStatus}
                    isRefreshing={isRefreshing}
                    selectedDomainId={selectedDomain}
                  />
                </TabsContent>
              </Tabs>
            </CardContent>
          </Card>

          {/* DNS Instructions for Selected Domain */}
          {selectedDomain &&
            domains.find((d) => d.id === selectedDomain && d.status === 'pending') && (
              <Card>
                <CardHeader>
                  <CardTitle>DNS Verification Instructions</CardTitle>
                  <CardDescription>
                    Add the following TXT record to your domain's DNS settings
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <DNSInstructions
                    domain={domains.find((d) => d.id === selectedDomain)?.name || ''}
                    txtRecord={domains.find((d) => d.id === selectedDomain)?.txtRecord || ''}
                  />

                  {verificationStatus && (
                    <div className="mt-4">
                      <VerificationStatus
                        status={verificationStatus}
                        domain={domains.find((d) => d.id === selectedDomain)?.name || ''}
                      />
                    </div>
                  )}

                  <div className="mt-4">
                    <Button
                      onClick={refreshVerificationStatus}
                      disabled={isRefreshing}
                      variant="outline"
                    >
                      {isRefreshing ? (
                        <>
                          <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                          Checking DNS Records...
                        </>
                      ) : (
                        <>
                          <RefreshCw className="mr-2 h-4 w-4" />
                          Check Verification Status
                        </>
                      )}
                    </Button>
                  </div>
                </CardContent>
              </Card>
            )}

          {/* Domain Selector for Service Registration */}
          <Card>
            <CardHeader>
              <CardTitle>Quick Actions</CardTitle>
              <CardDescription>Select a verified domain to register new services</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-4">
                <DomainSelector
                  domains={domains.filter((d) => d.status === 'verified')}
                  onSelect={(domain) => {
                    // Navigate to service registration with selected domain
                    window.location.href = `/services/new?domain=${domain}`;
                  }}
                  placeholder="Select a verified domain"
                />
                <Alert>
                  <AlertDescription>
                    Only verified domains can be used to register new services. Complete domain
                    verification first.
                  </AlertDescription>
                </Alert>
              </div>
            </CardContent>
          </Card>
        </div>
      </DashboardContent>
    </div>
  );
}
