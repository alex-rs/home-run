export enum ServiceStatus {
  RUNNING = 'RUNNING',
  STOPPED = 'STOPPED',
  ERROR = 'ERROR',
  MAINTENANCE = 'MAINTENANCE',
}

export enum ConfigType {
  YAML = 'YAML',
  DOCKERFILE = 'DOCKERFILE',
  JSON = 'JSON',
  INI = 'INI',
}

export interface ServiceConfig {
  type: ConfigType;
  content: string;
  path: string;
  lastEdited: string;
}

export interface Service {
  id: string;
  name: string;
  status: ServiceStatus;
  port: number;
  url: string;
  configs: ServiceConfig[];
  uptime: string;
  cpuUsage: number; // Percent
  memoryUsage: number; // MB
  host?: string; // For federated services - 'local' or remote host name
}

export interface User {
  username: string;
  isAuthenticated: boolean;
}