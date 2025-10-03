/**
 * Global type declarations for experimental browser APIs
 */

declare global {
  interface PaymentManager {
    instruments: PaymentInstruments;
    userHint?: string;
    canMakePayment?: boolean;
  }

  interface PaymentInstruments {
    set(key: string, details: any): Promise<void>;
    get(key: string): Promise<any | undefined>;
    keys(): Promise<string[]>;
    has(key: string): Promise<boolean>;
    delete(key: string): Promise<boolean>;
    clear(): Promise<void>;
  }

  interface ServiceWorkerRegistration {
    paymentManager?: PaymentManager;
  }

  interface Window {
    PaymentManager?: any;
  }
}

export {};