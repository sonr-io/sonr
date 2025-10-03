export class HighwaySDK {
  private apiUrl: string;

  constructor(config: { apiUrl?: string } = {}) {
    this.apiUrl = config.apiUrl || 'https://api.highway.sonr.io';
  }

  async health(): Promise<any> {
    const response = await fetch(`${this.apiUrl}/health`);
    return response.json();
  }
}

export default HighwaySDK;

// Export WebAuthn module
export * from './webauthn';
