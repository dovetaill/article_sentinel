export interface ApiEnvelope<T> {
  code: number;
  message: string;
  data?: T;
}

export const FIXED_LOGIN_URL = 'https://appadmin.cq.qiludev.com/cq-admin/index.html';

export function redirectToFixedLogin() {
  if (typeof window !== 'undefined' && typeof window.location?.assign === 'function') {
    window.location.assign(FIXED_LOGIN_URL);
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
    const message = envelope?.message?.trim() || `Request failed with status ${response.status}`;
    throw new Error(message);
  }

  return envelope.data as T;
}
