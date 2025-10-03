/**
 * W3C Payment Handler API client for Motor Payment Gateway
 * 
 * @packageDocumentation
 */

/// <reference path="./payment-handler.d.ts" />

import type {
  PaymentInstrument,
  PaymentMethod,
  PaymentRequestEvent,
  PaymentHandlerResponse,
  CanMakePaymentRequest,
  CanMakePaymentResponse,
  ProcessPaymentRequest,
  ProcessPaymentResponse,
  ValidatePaymentMethodRequest,
  ValidatePaymentMethodResponse,
  PaymentStatus,
  RefundPaymentRequest,
  RefundPaymentResponse,
  MotorServiceWorkerConfig,
} from './types';

/**
 * Payment Gateway client for Motor service worker
 */
export class PaymentGatewayClient {
  private readonly baseUrl: string;
  private readonly config: MotorServiceWorkerConfig;

  constructor(config: MotorServiceWorkerConfig = {}) {
    this.config = {
      worker_url: '/api',
      timeout: 30000,
      max_retries: 3,
      debug: false,
      ...config,
    };
    this.baseUrl = this.config.worker_url || '/api';
  }

  /**
   * Make an API request to the payment gateway
   */
  private async request<T>(
    method: string,
    endpoint: string,
    data?: unknown
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.config.timeout || 30000);

    try {
      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: data ? JSON.stringify(data) : undefined,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        const error = await response.text();
        throw new Error(`Payment API error: ${response.status} - ${error}`);
      }

      return await response.json();
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof Error && error.name === 'AbortError') {
        throw new Error('Payment request timed out');
      }
      throw error;
    }
  }

  /**
   * Get available payment instruments
   */
  async getPaymentInstruments(): Promise<PaymentInstrument[]> {
    return this.request<PaymentInstrument[]>('GET', '/payment/instruments');
  }

  /**
   * Check if payment can be made with given methods
   */
  async canMakePayment(request: CanMakePaymentRequest): Promise<CanMakePaymentResponse> {
    return this.request<CanMakePaymentResponse>('POST', '/payment/canmakepayment', request);
  }

  /**
   * Handle W3C Payment Request Event
   */
  async handlePaymentRequest(event: PaymentRequestEvent): Promise<PaymentHandlerResponse> {
    return this.request<PaymentHandlerResponse>('POST', '/payment/paymentrequest', event);
  }

  /**
   * Process a payment transaction
   */
  async processPayment(request: ProcessPaymentRequest): Promise<ProcessPaymentResponse> {
    return this.request<ProcessPaymentResponse>('POST', '/payment/process', request);
  }

  /**
   * Validate a payment method
   */
  async validatePaymentMethod(request: ValidatePaymentMethodRequest): Promise<ValidatePaymentMethodResponse> {
    return this.request<ValidatePaymentMethodResponse>('POST', '/payment/validate', request);
  }

  /**
   * Get payment status by ID
   */
  async getPaymentStatus(paymentId: string): Promise<PaymentStatus> {
    return this.request<PaymentStatus>('GET', `/payment/status/${paymentId}`);
  }

  /**
   * Process a refund
   */
  async refundPayment(request: RefundPaymentRequest): Promise<RefundPaymentResponse> {
    return this.request<RefundPaymentResponse>('POST', '/payment/refund', request);
  }
}

/**
 * W3C Payment Handler implementation for Motor
 */
export class MotorPaymentHandler {
  private client: PaymentGatewayClient;
  private registered = false;

  constructor(config?: MotorServiceWorkerConfig) {
    this.client = new PaymentGatewayClient(config);
  }

  /**
   * Register the payment handler with the browser
   */
  async register(): Promise<void> {
    if (!('PaymentManager' in window)) {
      throw new Error('Payment Handler API not supported in this browser');
    }

    const registration = await navigator.serviceWorker.ready;
    
    if (!registration.paymentManager) {
      throw new Error('Payment Manager not available in service worker');
    }

    // Get available instruments from the server
    const instruments = await this.client.getPaymentInstruments();

    // Register each instrument with the browser
    for (const instrument of instruments) {
      if (instrument.enabled) {
        await registration.paymentManager.instruments.set(instrument.id, {
          name: instrument.name,
          icons: instrument.iconUrl ? [{ src: instrument.iconUrl, sizes: '32x32', type: 'image/png' }] : undefined,
          method: instrument.type,
        });
      }
    }

    this.registered = true;
  }

  /**
   * Unregister all payment instruments
   */
  async unregister(): Promise<void> {
    if (!('PaymentManager' in window)) {
      return;
    }

    const registration = await navigator.serviceWorker.ready;
    
    if (!registration.paymentManager) {
      return;
    }

    await registration.paymentManager.instruments.clear();
    this.registered = false;
  }

  /**
   * Check if payment handler is registered
   */
  isRegistered(): boolean {
    return this.registered;
  }

  /**
   * Get the payment gateway client
   */
  getClient(): PaymentGatewayClient {
    return this.client;
  }
}

/**
 * Payment Request builder for Motor
 */
export class MotorPaymentRequest {
  private methods: PaymentMethod[] = [];
  private details: any = null;
  private options: PaymentOptions = {};

  /**
   * Add a payment method
   */
  addMethod(method: PaymentMethod): this {
    this.methods.push(method);
    return this;
  }

  /**
   * Set payment details
   */
  setDetails(details: any): this {
    this.details = details;
    return this;
  }

  /**
   * Set payment options
   */
  setOptions(options: PaymentOptions): this {
    this.options = options;
    return this;
  }

  /**
   * Build and show the payment request
   */
  async show(): Promise<PaymentResponse> {
    if (!('PaymentRequest' in window)) {
      throw new Error('Payment Request API not supported');
    }

    if (this.methods.length === 0) {
      throw new Error('At least one payment method is required');
    }

    if (!this.details) {
      throw new Error('Payment details are required');
    }

    const request = new PaymentRequest(this.methods, this.details, this.options);
    
    // Check if payment can be made
    const canMake = await request.canMakePayment();
    if (!canMake) {
      throw new Error('No supported payment methods available');
    }

    // Show the payment UI
    return await request.show();
  }

  /**
   * Create a payment request for a simple transaction
   */
  static simple(
    amount: number,
    currency: string,
    label: string,
    methods: PaymentMethod[] = [{ supportedMethods: 'basic-card' }]
  ): MotorPaymentRequest {
    const request = new MotorPaymentRequest();
    
    methods.forEach(method => request.addMethod(method));
    
    request.setDetails({
      total: {
        label,
        amount: {
          currency,
          value: amount.toFixed(2),
        },
      },
    });

    return request;
  }
}

/**
 * Create a payment gateway client
 */
export function createPaymentGatewayClient(config?: MotorServiceWorkerConfig): PaymentGatewayClient {
  return new PaymentGatewayClient(config);
}

/**
 * Create a payment handler
 */
export function createPaymentHandler(config?: MotorServiceWorkerConfig): MotorPaymentHandler {
  return new MotorPaymentHandler(config);
}

/**
 * Check if Payment Handler API is supported
 */
export function isPaymentHandlerSupported(): boolean {
  return typeof window !== 'undefined' && 
         'serviceWorker' in navigator && 
         'PaymentManager' in window;
}

/**
 * Check if Payment Request API is supported
 */
export function isPaymentRequestSupported(): boolean {
  return typeof window !== 'undefined' && 'PaymentRequest' in window;
}