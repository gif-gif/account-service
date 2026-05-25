import type { APIErrorBody } from "./types";
import { apiBaseURL } from "./env";

export class APIError extends Error {
  code: string;
  requestId: string;
  status: number;

  constructor(code: string, message: string, requestId: string, status: number) {
    super(message);
    this.name = "APIError";
    this.code = code;
    this.requestId = requestId;
    this.status = status;
  }
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const baseURL = apiBaseURL();
  const response = await fetch(`${baseURL}${path}`, {
    ...init,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(init.headers ?? {}),
    },
  });

  const body = await parseJSON<unknown>(response);
  if (!response.ok) {
    const errorBody = body as APIErrorBody;
    throw new APIError(
      errorBody.error?.code ?? "request_failed",
      errorBody.error?.message ?? "Request failed",
      errorBody.error?.request_id ?? response.headers.get("X-Request-ID") ?? "",
      response.status,
    );
  }

  return body as T;
}

async function parseJSON<T>(response: Response): Promise<T> {
  const text = await response.text();
  if (!text) {
    return {} as T;
  }
  return JSON.parse(text) as T;
}
