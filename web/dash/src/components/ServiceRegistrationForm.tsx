'use client';

import { zodResolver } from '@hookform/resolvers/zod';
import {
  Alert,
  AlertDescription,
  Button,
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  Input,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
  Stepper,
  StepperItem,
  Textarea,
} from '@sonr.io/ui';
import { Loader2 } from 'lucide-react';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import * as z from 'zod';

const serviceSchema = z.object({
  name: z.string().min(3, 'Service name must be at least 3 characters'),
  description: z.string().min(10, 'Description must be at least 10 characters'),
  domain: z.string().url('Must be a valid domain'),
  category: z.string().min(1, 'Please select a category'),
  permissions: z.array(z.string()).min(1, 'Select at least one permission'),
});

type ServiceFormData = z.infer<typeof serviceSchema>;

interface ServiceRegistrationFormProps {
  onSuccess?: () => void;
}

export function ServiceRegistrationForm({ onSuccess }: ServiceRegistrationFormProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [currentStep, setCurrentStep] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const form = useForm<ServiceFormData>({
    resolver: zodResolver(serviceSchema),
    defaultValues: {
      name: '',
      description: '',
      domain: '',
      category: '',
      permissions: [],
    },
  });

  const onSubmit = async (data: ServiceFormData) => {
    setIsSubmitting(true);
    setError(null);

    try {
      // TODO: Implement actual service registration API call
      const response = await fetch('/api/services', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        throw new Error('Failed to register service');
      }

      onSuccess?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setIsSubmitting(false);
    }
  };

  const steps = [
    { title: 'Basic Info', description: 'Service name and description' },
    { title: 'Domain', description: 'Configure domain verification' },
    { title: 'Permissions', description: 'Set required permissions' },
  ];

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <div className="space-y-4">
          {/* Step 1: Basic Info */}
          <div className={currentStep === 0 ? 'block' : 'hidden'}>
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Service Name</FormLabel>
                  <FormControl>
                    <Input placeholder="My Awesome Service" {...field} />
                  </FormControl>
                  <FormDescription>A unique name for your service</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem className="mt-4">
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="Describe what your service does..."
                      {...field}
                      rows={4}
                    />
                  </FormControl>
                  <FormDescription>Help users understand your service's purpose</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="category"
              render={({ field }) => (
                <FormItem className="mt-4">
                  <FormLabel>Category</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select a category" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="api">API Service</SelectItem>
                      <SelectItem value="webapp">Web Application</SelectItem>
                      <SelectItem value="mobile">Mobile App</SelectItem>
                      <SelectItem value="iot">IoT Device</SelectItem>
                      <SelectItem value="other">Other</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormDescription>Choose the category that best fits your service</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          {/* Step 2: Domain */}
          <div className={currentStep === 1 ? 'block' : 'hidden'}>
            <FormField
              control={form.control}
              name="domain"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Domain</FormLabel>
                  <FormControl>
                    <Input placeholder="https://example.com" {...field} />
                  </FormControl>
                  <FormDescription>The domain where your service is hosted</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Alert className="mt-4">
              <AlertDescription>
                After registration, you'll need to verify domain ownership by adding a TXT record to
                your DNS.
              </AlertDescription>
            </Alert>
          </div>

          {/* Step 3: Permissions */}
          <div className={currentStep === 2 ? 'block' : 'hidden'}>
            <FormField
              control={form.control}
              name="permissions"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Required Permissions</FormLabel>
                  <div className="space-y-2 mt-2">
                    {['read:profile', 'write:data', 'read:credentials', 'manage:vault'].map(
                      (permission) => (
                        <label
                          key={permission}
                          className="flex items-center space-x-2 cursor-pointer"
                        >
                          <input
                            type="checkbox"
                            value={permission}
                            checked={field.value?.includes(permission)}
                            onChange={(e) => {
                              const value = e.target.value;
                              const current = field.value || [];
                              if (e.target.checked) {
                                field.onChange([...current, value]);
                              } else {
                                field.onChange(current.filter((v) => v !== value));
                              }
                            }}
                            className="rounded border-gray-300"
                          />
                          <span className="text-sm">{permission}</span>
                        </label>
                      )
                    )}
                  </div>
                  <FormDescription className="mt-2">
                    Select the permissions your service requires
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </div>

        {/* Navigation */}
        <div className="flex justify-between pt-4">
          <Button
            type="button"
            variant="outline"
            onClick={() => setCurrentStep((prev) => Math.max(0, prev - 1))}
            disabled={currentStep === 0}
          >
            Previous
          </Button>

          {currentStep < steps.length - 1 ? (
            <Button
              type="button"
              onClick={() => setCurrentStep((prev) => Math.min(steps.length - 1, prev + 1))}
            >
              Next
            </Button>
          ) : (
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Registering...
                </>
              ) : (
                'Register Service'
              )}
            </Button>
          )}
        </div>
      </form>
    </Form>
  );
}
