import type { APIErrorBody } from "./types";
import { apiBaseURL } from "./env";
import { clearAuthTokens, getAuthTokens, notifyAuthExpired, setAuthTokens, type AuthTokens } from "./authTokens";

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
  const response = await request(path, init, getAuthTokens()?.accessToken);
  if (response.status === 401 && path !== "/api/v1/admin/refresh") {
    const refreshed = await refreshTokens();
    if (refreshed) {
      return handleResponse<T>(await request(path, init, refreshed.accessToken));
    }
  }

  return handleResponse<T>(response);
}

async function request(path: string, init: RequestInit, accessToken?: string): Promise<Response> {
  const baseURL = apiBaseURL();
  return fetch(`${baseURL}${path}`, {
    ...init,
    credentials: "omit",
    headers: {
      "Content-Type": "application/json",
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
      ...(init.headers ?? {}),
    },
  });
}

async function handleResponse<T>(response: Response): Promise<T> {
  const body = await parseResponseBody(response);
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

async function refreshTokens(): Promise<AuthTokens | null> {
  const tokens = getAuthTokens();
  if (!tokens?.refreshToken) {
    return null;
  }
  const baseURL = apiBaseURL();
  const response = await fetch(`${baseURL}/api/v1/admin/refresh`, {
    method: "POST",
    credentials: "omit",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ refreshToken: tokens.refreshToken }),
  });
  if (!response.ok) {
    clearAuthTokens();
    notifyAuthExpired();
    return null;
  }
  const body = (await parseResponseBody(response)) as { accessToken?: string; refreshToken?: string };
  if (!body.accessToken || !body.refreshToken) {
    clearAuthTokens();
    notifyAuthExpired();
    return null;
  }
  const refreshed = { accessToken: body.accessToken, refreshToken: body.refreshToken };
  setAuthTokens(refreshed);
  return refreshed;
}

async function parseResponseBody(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text) {
    return {};
  }
  try {
    return JSON.parse(text) as unknown;
  } catch (error) {
    if (response.ok) {
      throw error;
    }
    return {
      error: {
        code: "request_failed",
        message: text || response.statusText || "Request failed",
        request_id: response.headers.get("X-Request-ID") ?? "",
      },
    };
  }
}
