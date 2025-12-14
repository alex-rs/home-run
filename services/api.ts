// API service for communicating with the backend

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api';

// Generic fetch wrapper with credentials
async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    credentials: 'include', // Important for session cookies
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: 'Request failed' }));
    throw new Error(error.error || `HTTP ${response.status}`);
  }

  return response.json();
}

// Auth API
export interface LoginResponse {
  success: boolean;
  user?: { username: string };
  error?: string;
}

export async function login(username: string, password: string): Promise<LoginResponse> {
  return apiFetch<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export async function logout(): Promise<void> {
  await apiFetch('/auth/logout', { method: 'POST' });
}

export async function checkAuth(): Promise<LoginResponse> {
  return apiFetch<LoginResponse>('/auth/check');
}

// Services API
import { Service, ServiceConfig } from '../types';

export interface ServicesResponse {
  services: Service[];
  total: number;
  running: number;
}

export async function getServices(): Promise<ServicesResponse> {
  return apiFetch<ServicesResponse>('/services');
}

export async function getService(id: string): Promise<Service> {
  return apiFetch<Service>(`/services/${id}`);
}

export async function getServiceConfig(serviceId: string, configIndex: number): Promise<ServiceConfig> {
  return apiFetch<ServiceConfig>(`/services/${serviceId}/configs/${configIndex}`);
}

// Host Stats API
export interface HostStats {
  cpu: {
    usage: number;
    cores: number;
    threads: number;
  };
  memory: {
    usedGB: number;
    totalGB: number;
  };
  storage: {
    usedGB: number;
    totalGB: number;
  };
}

export async function getHostStats(): Promise<HostStats> {
  return apiFetch<HostStats>('/host/stats');
}
