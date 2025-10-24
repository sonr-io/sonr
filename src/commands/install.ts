// Install command - Install or update snrd binary
import { Logger } from '../lib/logger';
import { ConfigManager } from '../lib/config';

const GITHUB_REPO = 'sonr-io/sonr';
const BINARY_NAME = 'snrd';

export async function install(args: string[]) {
  Logger.header('Install Sonr Node Binary');

  const version = args[0] || 'latest';

  Logger.info(`Installing ${BINARY_NAME} version: ${version}`);

  try {
    // Check if running on supported platform
    const platform = process.platform;
    const arch = process.arch;

    Logger.info(`Detected platform: ${platform}/${arch}`);

    if (version === 'latest') {
      // Get latest release from GitHub
      Logger.step(1, 3, 'Fetching latest release information');
      const response = await fetch(`https://api.github.com/repos/${GITHUB_REPO}/releases/latest`);

      if (!response.ok) {
        throw new Error('Failed to fetch release information');
      }

      const release = await response.json();
      const latestVersion = release.tag_name;

      Logger.info(`Latest version: ${latestVersion}`);

      // Download binary
      Logger.step(2, 3, `Downloading ${BINARY_NAME}`);
      const assetName = `${BINARY_NAME}-${platform}-${arch}`;
      const asset = release.assets.find((a: any) => a.name.includes(assetName));

      if (!asset) {
        throw new Error(`No binary found for ${platform}/${arch}`);
      }

      const stop = Logger.loading('Downloading binary...');
      const binaryResponse = await fetch(asset.browser_download_url);

      if (!binaryResponse.ok) {
        stop();
        throw new Error('Failed to download binary');
      }

      stop();

      // Save binary to /usr/local/bin or ~/.local/bin
      Logger.step(3, 3, 'Installing binary');
      const binaryData = await binaryResponse.arrayBuffer();

      try {
        // Try to install to /usr/local/bin first (system-wide)
        await Bun.write(`/usr/local/bin/${BINARY_NAME}`, binaryData);
        await Bun.$`chmod +x /usr/local/bin/${BINARY_NAME}`;
        Logger.success(`Installed ${BINARY_NAME} to /usr/local/bin`);
      } catch {
        // Fall back to ~/.local/bin (user-level)
        const localBinDir = `${process.env.HOME}/.local/bin`;
        await Bun.$`mkdir -p ${localBinDir}`;
        await Bun.write(`${localBinDir}/${BINARY_NAME}`, binaryData);
        await Bun.$`chmod +x ${localBinDir}/${BINARY_NAME}`;
        Logger.success(`Installed ${BINARY_NAME} to ${localBinDir}`);
        Logger.warn(`Make sure ${localBinDir} is in your PATH`);
      }

      // Verify installation
      const versionCheck = await Bun.$`${BINARY_NAME} version`.quiet();
      const installedVersion = (await versionCheck.text()).trim();

      Logger.success(`Successfully installed ${BINARY_NAME} ${installedVersion}`);

      // Update config
      const config = new ConfigManager();
      config.set('binary', BINARY_NAME);
      config.set('version', installedVersion);
      config.close();
    } else {
      throw new Error('Specific version installation not yet supported. Use "latest" for now.');
    }
  } catch (error) {
    Logger.error(`Installation failed: ${error}`);
    process.exit(1);
  }
}
