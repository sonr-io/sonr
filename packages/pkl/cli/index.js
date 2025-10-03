#!/usr/bin/env node

import { setTimeout } from 'node:timers/promises';
import { exec } from 'node:child_process';
import { promisify } from 'node:util';
import * as p from '@clack/prompts';
import color from 'picocolors';
import { getExePath } from '@pkl-community/pkl';

const execAsync = promisify(exec);
const pklPath = getExePath();

// Static list of available packages on mod.pkl.sh
const AVAILABLE_PACKAGES = [
  'sonr.beam',
  'sonr.core',
  'sonr.hway',
  'sonr.testnet'
];

// Module options for each package type
const MODULE_OPTIONS = {
  default: [
    { value: 'Config.pkl', label: 'Config', hint: 'Main configuration module' },
    { value: 'PklProject', label: 'Project', hint: 'Project definition' },
  ],
  'sonr.beam': [
    { value: 'Config.pkl', label: 'Config', hint: 'Main configuration' },
    { value: 'Element.pkl', label: 'Element', hint: 'Element configuration' },
    { value: 'Hookshot.pkl', label: 'Hookshot', hint: 'Hookshot configuration' },
    { value: 'Synapse.pkl', label: 'Synapse', hint: 'Synapse configuration' },
    { value: 'PklProject', label: 'Project', hint: 'Project definition' },
  ],
  'sonr.core': [
    { value: 'Config.pkl', label: 'Config', hint: 'Core configuration' },
    { value: 'Keys.pkl', label: 'Keys', hint: 'Key management' },
    { value: 'UCAN.pkl', label: 'UCAN', hint: 'UCAN configuration' },
    { value: 'Wallet.pkl', label: 'Wallet', hint: 'Wallet configuration' },
    { value: 'PklProject', label: 'Project', hint: 'Project definition' },
  ],
  'sonr.testnet': [
    { value: 'Starship.pkl', label: 'Starship', hint: 'Starship K8s configuration' },
    { value: 'Docker.pkl', label: 'Docker', hint: 'Docker Compose configuration' },
  ],
};

async function getAvailablePackages() {
  // Always return static list of packages available on mod.pkl.sh
  return AVAILABLE_PACKAGES;
}

async function main() {
  console.clear();

  await setTimeout(500);

  p.intro(`${color.bgCyan(color.black(' Sonr PKL '))} - Configuration CLI`);

  try {
    // Get available packages
    const availablePackages = await getAvailablePackages();

    if (availablePackages.length === 0) {
      p.cancel('No packages found.');
      process.exit(1);
    }

    const config = await p.group(
      {
        package: () =>
          p.select({
            message: 'Select a Pkl package to evaluate',
            options: availablePackages.map(pkg => ({
              value: pkg,
              label: pkg,
              hint: `https://mod.pkl.sh/${pkg}/`
            })),
            maxItems: 10
          }),
        module: ({ results }) => {
          const moduleOpts = MODULE_OPTIONS[results.package] || MODULE_OPTIONS.default;
          return p.select({
            message: 'Select module to evaluate',
            options: moduleOpts,
            maxItems: 8
          });
        },
        outputPath: () =>
          p.text({
            message: 'Output directory path',
            placeholder: './data',
            initialValue: './data',
            validate: value => {
              if (!value) return 'Output path is required';
              return;
            }
          }),
        format: () =>
          p.select({
            message: 'Output format',
            options: [
              { value: 'auto', label: 'Auto (from module)', hint: 'Use module\'s default format' },
              { value: 'yaml', label: 'YAML' },
              { value: 'json', label: 'JSON' },
              { value: 'jsonnet', label: 'Jsonnet' },
              { value: 'pcf', label: 'PCF (Pkl Configuration Format)' },
              { value: 'plist', label: 'Property List' },
              { value: 'properties', label: 'Java Properties' },
              { value: 'textproto', label: 'Text Proto' },
              { value: 'xml', label: 'XML' },
            ],
            initialValue: 'auto',
            maxItems: 8
          }),
      },
      {
        onCancel: () => {
          p.cancel('Operation cancelled.');
          process.exit(0);
        },
      }
    );

    const s = p.spinner();

    // Always use remote packages from mod.pkl.sh
    const moduleUrl = `https://mod.pkl.sh/${config.package}/${config.module}`;

    let pklCommand = `"${pklPath}" eval -m ${config.outputPath}`;

    if (config.format !== 'auto') {
      pklCommand += ` -f ${config.format}`;
    }

    pklCommand += ` ${moduleUrl}`;

    s.start(`Evaluating ${config.package}/${config.module}...`);

    try {
      const { stdout, stderr } = await execAsync(pklCommand);

      s.stop(`Successfully evaluated ${config.package}`);

      if (stdout) {
        p.note(stdout.trim(), 'Output');
      }

      if (stderr) {
        p.log.warning(stderr.trim());
      }

      p.outro(`Configuration generated in ${color.cyan(config.outputPath)}`);

    } catch (error) {
      s.stop('Evaluation failed');
      p.log.error(`Failed to evaluate module: ${error.message}`);

      if (error.stderr) {
        p.log.error(error.stderr);
      }

      process.exit(1);
    }

  } catch (error) {
    p.log.error(`An error occurred: ${error.message}`);
    process.exit(1);
  }
}

main().catch(console.error);
