// Docker utilities using Bun.$
import type { DockerNode } from '../types';
import { Logger } from './logger';

export class DockerManager {
  /**
   * Check if Docker is installed and running
   */
  static async isAvailable(): Promise<boolean> {
    try {
      const result = await Bun.$`docker info`.quiet();
      return result.exitCode === 0;
    } catch {
      return false;
    }
  }

  /**
   * Check if docker-compose is available
   */
  static async isComposeAvailable(): Promise<boolean> {
    try {
      const result = await Bun.$`docker compose version`.quiet();
      return result.exitCode === 0;
    } catch {
      return false;
    }
  }

  /**
   * Pull a Docker image
   */
  static async pullImage(image: string): Promise<void> {
    Logger.info(`Pulling Docker image: ${image}`);
    const stop = Logger.loading('Pulling image...');

    try {
      await Bun.$`docker pull ${image}`.quiet();
      stop();
      Logger.success(`Image pulled: ${image}`);
    } catch (error) {
      stop();
      throw new Error(`Failed to pull image: ${error}`);
    }
  }

  /**
   * Check if a container is running
   */
  static async isContainerRunning(name: string): Promise<boolean> {
    try {
      const result = await Bun.$`docker ps --filter name=${name} --format {{.Names}}`.quiet();
      const output = await result.text();
      return output.trim() === name;
    } catch {
      return false;
    }
  }

  /**
   * Start a container
   */
  static async startContainer(name: string): Promise<void> {
    Logger.info(`Starting container: ${name}`);
    await Bun.$`docker start ${name}`;
    Logger.success(`Container started: ${name}`);
  }

  /**
   * Stop a container
   */
  static async stopContainer(name: string): Promise<void> {
    Logger.info(`Stopping container: ${name}`);
    await Bun.$`docker stop ${name}`;
    Logger.success(`Container stopped: ${name}`);
  }

  /**
   * Remove a container
   */
  static async removeContainer(name: string, force: boolean = false): Promise<void> {
    Logger.info(`Removing container: ${name}`);
    const forceFlag = force ? '-f' : '';
    await Bun.$`docker rm ${forceFlag} ${name}`;
    Logger.success(`Container removed: ${name}`);
  }

  /**
   * Run a command in a container
   */
  static async exec(container: string, command: string): Promise<string> {
    const result = await Bun.$`docker exec ${container} sh -c ${command}`.quiet();
    return await result.text();
  }

  /**
   * Get container logs
   */
  static async logs(container: string, tail: number = 50): Promise<string> {
    const result = await Bun.$`docker logs ${container} --tail ${tail}`.quiet();
    return await result.text();
  }

  /**
   * Run docker-compose command
   */
  static async composeUp(composePath: string, detached: boolean = true): Promise<void> {
    const cwd = Bun.file(composePath).name ? composePath.split('/').slice(0, -1).join('/') : composePath;
    const flags = detached ? '-d' : '';

    Logger.info('Starting network with docker-compose...');
    await Bun.$`cd ${cwd} && docker compose up ${flags}`;
    Logger.success('Network started successfully');
  }

  /**
   * Stop docker-compose services
   */
  static async composeDown(composePath: string): Promise<void> {
    const cwd = Bun.file(composePath).name ? composePath.split('/').slice(0, -1).join('/') : composePath;

    Logger.info('Stopping network...');
    await Bun.$`cd ${cwd} && docker compose down`;
    Logger.success('Network stopped successfully');
  }

  /**
   * Get container status
   */
  static async getContainerStatus(name: string): Promise<{ running: boolean; status: string }> {
    try {
      const result = await Bun.$`docker ps -a --filter name=${name} --format {{.Status}}`.quiet();
      const status = (await result.text()).trim();
      const running = status.startsWith('Up');
      return { running, status };
    } catch {
      return { running: false, status: 'Not found' };
    }
  }

  /**
   * List all containers with a prefix
   */
  static async listContainers(prefix?: string): Promise<Array<{ name: string; status: string; image: string }>> {
    try {
      const filter = prefix ? `--filter name=${prefix}` : '';
      const result = await Bun.$`docker ps -a ${filter} --format {{.Names}}|||{{.Status}}|||{{.Image}}`.quiet();
      const output = await result.text();

      if (!output.trim()) {
        return [];
      }

      return output
        .trim()
        .split('\n')
        .map(line => {
          const [name, status, image] = line.split('|||');
          return { name, status, image };
        });
    } catch {
      return [];
    }
  }
}
