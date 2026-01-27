/**
 * Configuration Management
 */

import { SDKConfig } from '../types';

const DEFAULT_CONFIG: Partial<SDKConfig> = {
  batchSize: 10,
  batchInterval: 5000, // 5 seconds
  enableDebug: false,
  environment: undefined, // Auto-detect by default
};

export class ConfigManager {
  private config: SDKConfig;

  constructor(config: SDKConfig) {
    this.config = {
      ...DEFAULT_CONFIG,
      ...config,
    };

    this.validate();
  }

  /**
   * Validate configuration
   */
  private validate(): void {
    if (!this.config.apiKey || this.config.apiKey.trim() === '') {
      throw new Error('HawkEye SDK: apiKey is required');
    }

    if (!this.config.ingestionUrl || this.config.ingestionUrl.trim() === '') {
      throw new Error('HawkEye SDK: ingestionUrl is required');
    }

    try {
      new URL(this.config.ingestionUrl);
    } catch {
      throw new Error('HawkEye SDK: ingestionUrl must be a valid URL');
    }
  }

  /**
   * Get configuration
   */
  getConfig(): SDKConfig {
    return { ...this.config };
  }

  /**
   * Get API key
   */
  getApiKey(): string {
    return this.config.apiKey;
  }

  /**
   * Get ingestion URL
   */
  getIngestionUrl(): string {
    return this.config.ingestionUrl;
  }

  /**
   * Get batch size
   */
  getBatchSize(): number {
    return this.config.batchSize || 10;
  }

  /**
   * Get batch interval (ms)
   */
  getBatchInterval(): number {
    return this.config.batchInterval || 5000;
  }

  /**
   * Is debug enabled
   */
  isDebugEnabled(): boolean {
    return this.config.enableDebug || false;
  }

  /**
   * Get project ID
   */
  getProjectId(): string | undefined {
    return this.config.projectId;
  }

  /**
   * Get app ID
   */
  getAppId(): string | undefined {
    return this.config.appId;
  }

  /**
   * Get environment (auto-detect if not configured)
   */
  getEnvironment(): string {
    if (this.config.environment) {
      return this.config.environment;
    }
    // Auto-detect environment based on URL
    if (typeof window !== 'undefined') {
      const hostname = window.location.hostname;
      if (hostname === 'localhost' || hostname === '127.0.0.1') {
        return 'development';
      }
      if (hostname.includes('staging') || hostname.includes('stage') || hostname.includes('test')) {
        return 'staging';
      }
    }
    return 'production';
  }
}
