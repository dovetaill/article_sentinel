export interface ApiEnvelope<T> {
  code: number;
  message: string;
  data?: T;
}

export async function apiRequest<T>(input: string, init?: RequestInit): Promise<T> {
  const response = await fetch(input, {
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      ...(init?.headers ?? {})
    },
    ...init
  });

  let envelope: ApiEnvelope<T> | null = null;
  try {
    envelope = (await response.json()) as ApiEnvelope<T>;
  } catch {
    envelope = null;
  }

  if (!response.ok || !envelope || envelope.code !== 0) {
    const message = envelope?.message?.trim() || `Request failed with status ${response.status}`;
    throw new Error(message);
  }

  return envelope.data as T;
}
