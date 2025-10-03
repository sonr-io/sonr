/**
 * OAuth scope utilities for UCAN integration
 */

import type { UCANCapability } from './delegation';

/**
 * OAuth scope definition
 */
export interface OAuthScope {
  name: string;
  title: string;
  description: string;
  icon?: string;
  required?: boolean;
  capabilities: UCANCapability[];
  children?: string[]; // Hierarchical scopes
}

/**
 * Scope registry
 */
export class ScopeRegistry {
  private scopes: Map<string, OAuthScope> = new Map();

  constructor() {
    this.registerDefaultScopes();
  }

  /**
   * Register default OAuth scopes
   */
  private registerDefaultScopes() {
    // OpenID Connect scopes
    this.register({
      name: 'openid',
      title: 'OpenID',
      description: 'Authenticate using your Sonr identity',
      required: true,
      capabilities: [{ with: 'did:*', can: 'identify' }],
    });

    this.register({
      name: 'profile',
      title: 'Profile',
      description: 'Access your basic profile information',
      capabilities: [
        { with: 'profile:*', can: 'read' },
        { with: 'did:*', can: 'resolve' },
      ],
    });

    this.register({
      name: 'email',
      title: 'Email',
      description: 'Access your email address',
      capabilities: [{ with: 'profile:email', can: 'read' }],
    });

    // Vault scopes
    this.register({
      name: 'vault:read',
      title: 'Read Vault',
      description: 'Read data from your encrypted vault',
      capabilities: [
        { with: 'vault:*', can: 'read' },
        { with: 'vault:*', can: 'list' },
      ],
    });

    this.register({
      name: 'vault:write',
      title: 'Write Vault',
      description: 'Create and modify data in your vault',
      capabilities: [
        { with: 'vault:*', can: 'write' },
        { with: 'vault:*', can: 'create' },
        { with: 'vault:*', can: 'update' },
      ],
    });

    this.register({
      name: 'vault:delete',
      title: 'Delete Vault Data',
      description: 'Delete data from your vault',
      capabilities: [{ with: 'vault:*', can: 'delete' }],
    });

    this.register({
      name: 'vault:sign',
      title: 'Sign with Vault',
      description: 'Sign messages and transactions using vault keys',
      capabilities: [
        { with: 'vault:*', can: 'sign' },
        { with: 'tx:*', can: 'sign' },
      ],
    });

    // Hierarchical vault scope
    this.register({
      name: 'vault:admin',
      title: 'Vault Admin',
      description: 'Full administrative access to your vault',
      capabilities: [{ with: 'vault:*', can: '*' }],
      children: ['vault:read', 'vault:write', 'vault:delete', 'vault:sign'],
    });

    // Service management scopes
    this.register({
      name: 'service:read',
      title: 'Read Services',
      description: 'List and view your registered services',
      capabilities: [
        { with: 'service:*', can: 'read' },
        { with: 'service:*', can: 'list' },
      ],
    });

    this.register({
      name: 'service:write',
      title: 'Manage Services',
      description: 'Create and update service registrations',
      capabilities: [
        { with: 'service:*', can: 'create' },
        { with: 'service:*', can: 'update' },
      ],
    });

    this.register({
      name: 'service:delete',
      title: 'Delete Services',
      description: 'Remove service registrations',
      capabilities: [{ with: 'service:*', can: 'delete' }],
    });

    this.register({
      name: 'service:manage',
      title: 'Service Admin',
      description: 'Full control over service registrations',
      capabilities: [{ with: 'service:*', can: '*' }],
      children: ['service:read', 'service:write', 'service:delete'],
    });

    // DWN (Decentralized Web Node) scopes
    this.register({
      name: 'dwn:read',
      title: 'Read DWN',
      description: 'Read data from your Decentralized Web Node',
      capabilities: [
        { with: 'dwn:*', can: 'read' },
        { with: 'dwn:*', can: 'query' },
      ],
    });

    this.register({
      name: 'dwn:write',
      title: 'Write DWN',
      description: 'Store data in your Decentralized Web Node',
      capabilities: [
        { with: 'dwn:*', can: 'write' },
        { with: 'dwn:*', can: 'create' },
      ],
    });

    // Offline access
    this.register({
      name: 'offline_access',
      title: 'Offline Access',
      description: "Access your account when you're not present",
      capabilities: [{ with: 'token:*', can: 'refresh' }],
    });
  }

  /**
   * Register a new scope
   */
  register(scope: OAuthScope): void {
    this.scopes.set(scope.name, scope);
  }

  /**
   * Get a scope by name
   */
  get(name: string): OAuthScope | undefined {
    return this.scopes.get(name);
  }

  /**
   * Get all registered scopes
   */
  getAll(): OAuthScope[] {
    return Array.from(this.scopes.values());
  }

  /**
   * Check if a scope exists
   */
  has(name: string): boolean {
    return this.scopes.has(name);
  }

  /**
   * Validate requested scopes
   */
  validate(requested: string[]): { valid: string[]; invalid: string[] } {
    const valid: string[] = [];
    const invalid: string[] = [];

    for (const scope of requested) {
      if (this.has(scope)) {
        valid.push(scope);
      } else {
        invalid.push(scope);
      }
    }

    return { valid, invalid };
  }

  /**
   * Expand hierarchical scopes
   */
  expand(scopes: string[]): string[] {
    const expanded = new Set<string>();

    for (const scope of scopes) {
      expanded.add(scope);

      const definition = this.get(scope);
      if (definition?.children) {
        for (const child of definition.children) {
          expanded.add(child);
        }
      }
    }

    return Array.from(expanded);
  }

  /**
   * Get capabilities for scopes
   */
  getCapabilities(scopes: string[]): UCANCapability[] {
    const capabilities: UCANCapability[] = [];
    const seen = new Set<string>();

    // Expand hierarchical scopes
    const expanded = this.expand(scopes);

    for (const scope of expanded) {
      const definition = this.get(scope);
      if (definition) {
        for (const cap of definition.capabilities) {
          const key = `${cap.with}:${cap.can}`;
          if (!seen.has(key)) {
            seen.add(key);
            capabilities.push(cap);
          }
        }
      }
    }

    return capabilities;
  }

  /**
   * Get required scopes from a list
   */
  getRequired(scopes: string[]): string[] {
    return scopes.filter((scope) => {
      const definition = this.get(scope);
      return definition?.required === true;
    });
  }

  /**
   * Group scopes by category
   */
  groupByCategory(scopes: string[]): Map<string, OAuthScope[]> {
    const groups = new Map<string, OAuthScope[]>();

    for (const scope of scopes) {
      const definition = this.get(scope);
      if (definition) {
        const category = scope.split(':')[0] || 'general';

        if (!groups.has(category)) {
          groups.set(category, []);
        }

        groups.get(category)!.push(definition);
      }
    }

    return groups;
  }
}

/**
 * Default scope registry instance
 */
export const scopeRegistry = new ScopeRegistry();

/**
 * Scope validation utilities
 */
export const ScopeValidator = {
  /**
   * Check if scopes are valid
   */
  isValid(scopes: string[]): boolean {
    const { invalid } = scopeRegistry.validate(scopes);
    return invalid.length === 0;
  },

  /**
   * Check if scope is hierarchical
   */
  isHierarchical(scope: string): boolean {
    const definition = scopeRegistry.get(scope);
    return !!(definition?.children && definition.children.length > 0);
  },

  /**
   * Check if one scope includes another
   */
  includes(parent: string, child: string): boolean {
    const definition = scopeRegistry.get(parent);
    if (!definition?.children) return false;

    return definition.children.includes(child);
  },

  /**
   * Get minimal scope set (remove redundant child scopes)
   */
  minimize(scopes: string[]): string[] {
    const minimal = new Set<string>(scopes);

    for (const scope of scopes) {
      const definition = scopeRegistry.get(scope);
      if (definition?.children) {
        // Remove children if parent is present
        for (const child of definition.children) {
          minimal.delete(child);
        }
      }
    }

    return Array.from(minimal);
  },

  /**
   * Check if scopes grant a specific capability
   */
  hasCapability(scopes: string[], resource: string, action: string): boolean {
    const capabilities = scopeRegistry.getCapabilities(scopes);

    return capabilities.some((cap) => {
      const resourceMatch =
        cap.with === resource ||
        cap.with === '*' ||
        (cap.with.endsWith('*') && resource.startsWith(cap.with.slice(0, -1)));

      const actionMatch = cap.can === action || cap.can === '*';

      return resourceMatch && actionMatch;
    });
  },
};

/**
 * Scope formatting utilities
 */
export const ScopeFormatter = {
  /**
   * Format scope for display
   */
  format(scope: string): string {
    const definition = scopeRegistry.get(scope);
    return definition?.title || scope;
  },

  /**
   * Format scope list
   */
  formatList(scopes: string[]): string {
    return scopes.map((s) => this.format(s)).join(', ');
  },

  /**
   * Get scope description
   */
  describe(scope: string): string {
    const definition = scopeRegistry.get(scope);
    return definition?.description || `Access to ${scope}`;
  },

  /**
   * Get scope icon
   */
  getIcon(scope: string): string | undefined {
    const definition = scopeRegistry.get(scope);
    return definition?.icon;
  },
};

/**
 * Export utilities
 */
export default {
  registry: scopeRegistry,
  validator: ScopeValidator,
  formatter: ScopeFormatter,
};
