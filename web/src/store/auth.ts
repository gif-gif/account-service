import { create, type StoreApi, type UseBoundStore } from "zustand";

import { clearAuthTokens, setAuthExpiredHandler, setAuthTokens } from "../lib/authTokens";
import { APIError, apiFetch } from "../lib/api";
import type { AdminUser } from "../lib/types";

type AuthResponse = {
  user: AdminUser;
  accessToken: string;
  refreshToken: string;
};

export type AuthState = {
  user: AdminUser | null;
  loading: boolean;
  error: string | null;
  login: (username: string, password: string) => Promise<void>;
  restore: () => Promise<void>;
  logout: () => Promise<void>;
};

export type AuthStore = UseBoundStore<StoreApi<AuthState>>;

export function createAuthStore(): AuthStore {
  const store = create<AuthState>((set) => ({
    user: null,
    loading: false,
    error: null,
    async login(username, password) {
      set({ loading: true, error: null });
      try {
        const response = await apiFetch<AuthResponse>("/api/v1/admin/login", {
          method: "POST",
          body: JSON.stringify({ username, password }),
        });
        setAuthTokens({ accessToken: response.accessToken, refreshToken: response.refreshToken });
        set({ user: response.user, loading: false, error: null });
      } catch (error) {
        const message = error instanceof APIError ? error.message : "Login failed";
        set({ loading: false, error: message });
        throw error;
      }
    },
    async restore() {
      set({ loading: true, error: null });
      try {
        const response = await apiFetch<AuthResponse>("/api/v1/admin/me");
        set({ user: response.user, loading: false, error: null });
      } catch (error) {
        clearAuthTokens();
        set({ user: null, loading: false, error: null });
      }
    },
    async logout() {
      set({ loading: true, error: null });
      try {
        await apiFetch<{ ok: boolean }>("/api/v1/admin/logout", { method: "POST" });
      } finally {
        clearAuthTokens();
        set({ user: null, loading: false, error: null });
      }
    },
  }));
  setAuthExpiredHandler(() => store.setState({ user: null, loading: false, error: null }));
  return store;
}

export const useAuthStore = createAuthStore();
