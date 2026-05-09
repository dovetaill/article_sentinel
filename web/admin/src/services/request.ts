import { redirectToFixedLogin } from '@/utils/auth';

export interface ApiEnvelope<T> {
  code: number;
  message: string;
  data?: T;
}

export interface ApiRequestErrorOptions {
  message: string;
  status: number;
}

export class ApiRequestError extends Error {
  status: number;

  constructor(options: ApiRequestErrorOptions) {
    super(options.message);
    this.name = 'ApiRequestError';
    this.status = options.status;
  }
}

export async function apiRequest<T>(input: string, init?: RequestInit): Promise<T> {
  const response = await fetch(input, {
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      ...(init?.headers ?? {})
    },
    ...init,
    credentials: 'same-origin'
  });

  let envelope: ApiEnvelope<T> | null = null;
  try {
    envelope = (await response.json()) as ApiEnvelope<T>;
  } catch {
    envelope = null;
  }

  if (response.status === 401) {
    redirectToFixedLogin();
  }

  if (!response.ok || !envelope || envelope.code !== 0) {
    throw new ApiRequestError({
      message: envelope?.message?.trim() || `Request failed with status ${response.status}`,
      status: response.status
    });
  }

  return envelope.data as T;
}
