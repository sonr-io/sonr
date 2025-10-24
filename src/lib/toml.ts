// TOML utilities for configuration management
import TOML from '@iarna/toml';

export class TomlManager {
  /**
   * Read and parse a TOML file
   */
  static async read(filePath: string): Promise<any> {
    const file = Bun.file(filePath);
    const content = await file.text();
    return TOML.parse(content);
  }

  /**
   * Write an object to a TOML file
   */
  static async write(filePath: string, data: any): Promise<void> {
    const content = TOML.stringify(data);
    await Bun.write(filePath, content);
  }

  /**
   * Set a value in a TOML file (supports nested paths with dot notation)
   */
  static async setValue(filePath: string, path: string, value: any): Promise<void> {
    const data = await this.read(filePath);
    const keys = path.split('.');

    let current = data;
    for (let i = 0; i < keys.length - 1; i++) {
      const key = keys[i];
      if (!current[key]) {
        current[key] = {};
      }
      current = current[key];
    }

    current[keys[keys.length - 1]] = value;

    await this.write(filePath, data);
  }

  /**
   * Get a value from a TOML file
   */
  static async getValue(filePath: string, path: string): Promise<any> {
    const data = await this.read(filePath);
    const keys = path.split('.');

    let current = data;
    for (const key of keys) {
      if (current[key] === undefined) {
        return undefined;
      }
      current = current[key];
    }

    return current;
  }

  /**
   * Merge configurations
   */
  static merge(base: any, override: any): any {
    const result = { ...base };

    for (const [key, value] of Object.entries(override)) {
      if (typeof value === 'object' && !Array.isArray(value) && value !== null) {
        result[key] = this.merge(result[key] || {}, value);
      } else {
        result[key] = value;
      }
    }

    return result;
  }
}
