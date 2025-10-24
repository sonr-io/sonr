// Configuration management using bun:sqlite
import { Database } from 'bun:sqlite';
import { homedir } from 'os';
import { join } from 'path';
import { DEFAULT_CHAIN_ID, DEFAULT_HOME_DIR, DEFAULT_BINARY } from './constants';

const CONFIG_DIR = join(homedir(), '.config', 'sonr');
const CONFIG_DB = join(CONFIG_DIR, 'config.db');

export class ConfigManager {
  private db: Database;

  constructor() {
    // Ensure config directory exists
    if (!Bun.file(CONFIG_DIR).size) {
      Bun.spawnSync(['mkdir', '-p', CONFIG_DIR]);
    }

    this.db = new Database(CONFIG_DB, { create: true });
    this.init();
  }

  /**
   * Initialize database tables
   */
  private init() {
    this.db.run(`
      CREATE TABLE IF NOT EXISTS config (
        key TEXT PRIMARY KEY,
        value TEXT NOT NULL
      )
    `);

    this.db.run(`
      CREATE TABLE IF NOT EXISTS nodes (
        name TEXT PRIMARY KEY,
        type TEXT NOT NULL,
        chain_id TEXT NOT NULL,
        home TEXT NOT NULL,
        moniker TEXT NOT NULL,
        created_at INTEGER NOT NULL
      )
    `);

    // Set defaults if not exists
    const defaults = {
      chain_id: DEFAULT_CHAIN_ID,
      home: DEFAULT_HOME_DIR,
      binary: DEFAULT_BINARY,
    };

    for (const [key, value] of Object.entries(defaults)) {
      const existing = this.get(key);
      if (!existing) {
        this.set(key, value);
      }
    }
  }

  /**
   * Get a configuration value
   */
  get(key: string): string | null {
    const result = this.db.query('SELECT value FROM config WHERE key = ?').get(key) as { value: string } | null;
    return result?.value || null;
  }

  /**
   * Set a configuration value
   */
  set(key: string, value: string) {
    this.db.run('INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)', [key, value]);
  }

  /**
   * Get all configuration
   */
  getAll(): Record<string, string> {
    const rows = this.db.query('SELECT key, value FROM config').all() as Array<{ key: string; value: string }>;
    return Object.fromEntries(rows.map(row => [row.key, row.value]));
  }

  /**
   * Add a node to the registry
   */
  addNode(name: string, type: string, chainId: string, home: string, moniker: string) {
    this.db.run(
      'INSERT OR REPLACE INTO nodes (name, type, chain_id, home, moniker, created_at) VALUES (?, ?, ?, ?, ?, ?)',
      [name, type, chainId, home, moniker, Date.now()]
    );
  }

  /**
   * Get a node from the registry
   */
  getNode(name: string): any {
    return this.db.query('SELECT * FROM nodes WHERE name = ?').get(name);
  }

  /**
   * List all nodes
   */
  listNodes(): any[] {
    return this.db.query('SELECT * FROM nodes ORDER BY created_at DESC').all();
  }

  /**
   * Remove a node
   */
  removeNode(name: string) {
    this.db.run('DELETE FROM nodes WHERE name = ?', [name]);
  }

  /**
   * Close database connection
   */
  close() {
    this.db.close();
  }
}
