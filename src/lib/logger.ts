// Beautiful logging utilities for sonrctl

const COLORS = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  dim: '\x1b[2m',

  // Foreground colors
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[37m',
  gray: '\x1b[90m',
} as const;

const ICONS = {
  info: 'ℹ',
  success: '✓',
  warning: '⚠',
  error: '✗',
  rocket: '🚀',
  gear: '⚙',
  chain: '⛓',
  node: '🔗',
  docker: '🐳',
};

export class Logger {
  static info(message: string, ...args: any[]) {
    console.log(`${COLORS.blue}${ICONS.info}${COLORS.reset} ${message}`, ...args);
  }

  static success(message: string, ...args: any[]) {
    console.log(`${COLORS.green}${ICONS.success}${COLORS.reset} ${message}`, ...args);
  }

  static warn(message: string, ...args: any[]) {
    console.log(`${COLORS.yellow}${ICONS.warning}${COLORS.reset} ${message}`, ...args);
  }

  static error(message: string, ...args: any[]) {
    console.error(`${COLORS.red}${ICONS.error}${COLORS.reset} ${message}`, ...args);
  }

  static debug(message: string, ...args: any[]) {
    if (process.env.DEBUG) {
      console.log(`${COLORS.gray}[DEBUG]${COLORS.reset} ${message}`, ...args);
    }
  }

  static header(message: string) {
    const line = '═'.repeat(message.length + 4);
    console.log(`\n${COLORS.cyan}${line}${COLORS.reset}`);
    console.log(`${COLORS.cyan}  ${message}${COLORS.reset}`);
    console.log(`${COLORS.cyan}${line}${COLORS.reset}\n`);
  }

  static section(message: string) {
    console.log(`\n${COLORS.bright}${COLORS.cyan}▸ ${message}${COLORS.reset}`);
  }

  static step(step: number, total: number, message: string) {
    console.log(`${COLORS.dim}[${step}/${total}]${COLORS.reset} ${message}`);
  }

  static command(cmd: string) {
    console.log(`${COLORS.dim}$ ${cmd}${COLORS.reset}`);
  }

  static json(obj: any) {
    console.log(JSON.stringify(obj, null, 2));
  }

  static table(data: Record<string, any>) {
    const maxKeyLength = Math.max(...Object.keys(data).map(k => k.length));

    console.log();
    for (const [key, value] of Object.entries(data)) {
      const paddedKey = key.padEnd(maxKeyLength);
      console.log(`  ${COLORS.cyan}${paddedKey}${COLORS.reset} : ${value}`);
    }
    console.log();
  }

  static loading(message: string): () => void {
    const frames = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];
    let i = 0;

    const timer = setInterval(() => {
      process.stdout.write(`\r${COLORS.blue}${frames[i]}${COLORS.reset} ${message}`);
      i = (i + 1) % frames.length;
    }, 80);

    return () => {
      clearInterval(timer);
      process.stdout.write('\r\x1b[K'); // Clear line
    };
  }
}
