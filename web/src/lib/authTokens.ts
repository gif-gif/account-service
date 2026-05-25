export type AuthTokens = {
  accessToken: string;
  refreshToken: string;
};

const storageKey = "account-service.auth.tokens";

let authExpiredHandler: (() => void) | null = null;

export function getAuthTokens(): AuthTokens | null {
  const raw = localStorage.getItem(storageKey);
  if (!raw) {
    return null;
  }
  try {
    const parsed = JSON.parse(raw) as Partial<AuthTokens>;
    if (!parsed.accessToken || !parsed.refreshToken) {
      return null;
    }
    return { accessToken: parsed.accessToken, refreshToken: parsed.refreshToken };
  } catch {
    return null;
  }
}

export function setAuthTokens(tokens: AuthTokens) {
  localStorage.setItem(storageKey, JSON.stringify(tokens));
}

export function clearAuthTokens() {
  localStorage.removeItem(storageKey);
}

export function setAuthExpiredHandler(handler: (() => void) | null) {
  authExpiredHandler = handler;
}

export function notifyAuthExpired() {
  authExpiredHandler?.();
}
