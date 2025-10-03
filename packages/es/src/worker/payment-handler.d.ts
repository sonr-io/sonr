/**
 * W3C Payment Handler API type declarations
 * These are experimental APIs not yet in standard TypeScript definitions
 */

interface PaymentManager {
  instruments: PaymentInstruments;
  userHint?: string;
  canMakePayment?: boolean;
}

interface PaymentInstruments {
  set(key: string, details: PaymentInstrument): Promise<void>;
  get(key: string): Promise<PaymentInstrument | undefined>;
  keys(): Promise<string[]>;
  has(key: string): Promise<boolean>;
  delete(key: string): Promise<boolean>;
  clear(): Promise<void>;
}

interface PaymentInstrument {
  name: string;
  icons?: Array<{
    src: string;
    sizes?: string;
    type?: string;
  }>;
  method?: string;
  capabilities?: {
    [key: string]: any;
  };
}

interface ServiceWorkerRegistration {
  paymentManager?: PaymentManager;
}

interface Window {
  PaymentManager?: any;
}

export {};