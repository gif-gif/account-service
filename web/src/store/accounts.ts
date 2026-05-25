import { create, type StoreApi, type UseBoundStore } from "zustand";

import { APIError, apiFetch } from "../lib/api";

export const accountTypes = ["claude", "aws", "gpt", "kiro", "claudecode", "codex"] as const;

export type AccountType = (typeof accountTypes)[number];

export type Account = {
  id: string;
  username: string;
  password?: string;
  login_url?: string;
  access_token?: string;
  refresh_token?: string;
  region: string;
  account_type: AccountType;
  status: string;
  quota_total?: number;
  quota_used?: number;
  quota_remaining: number;
  max_concurrent_leases?: number;
  tags: string[];
  notes?: string;
  created_at?: string;
  updated_at?: string;
};

export type AccountFilters = {
  region: string;
  accountType: string;
  status: string;
  tags: string;
  minQuotaRemaining: number;
  limit: number;
};

type QueryResponse = {
  accounts: Account[];
};

export type AccountsState = {
  accounts: Account[];
  filters: AccountFilters;
  loading: boolean;
  error: string | null;
  setFilter: <K extends keyof AccountFilters>(key: K, value: AccountFilters[K]) => void;
  load: () => Promise<void>;
};

export type AccountsStore = UseBoundStore<StoreApi<AccountsState>>;

const defaultFilters: AccountFilters = {
  region: "",
  accountType: "",
  status: "",
  tags: "",
  minQuotaRemaining: 0,
  limit: 10,
};

export function createAccountsStore(): AccountsStore {
  return create<AccountsState>((set, get) => ({
    accounts: [],
    filters: defaultFilters,
    loading: false,
    error: null,
    setFilter(key, value) {
      set((state) => ({ filters: { ...state.filters, [key]: value } }));
    },
    async load() {
      set({ loading: true, error: null });
      const filters = get().filters;
      try {
        const response = await apiFetch<QueryResponse>("/api/v1/accounts/query", {
          method: "POST",
          body: JSON.stringify({
            region: filters.region || undefined,
            account_type: filters.accountType || undefined,
            statuses: filters.status ? [filters.status] : undefined,
            tags: filters.tags ? filters.tags.split(",").map((tag) => tag.trim()).filter(Boolean) : undefined,
            min_quota_remaining: filters.minQuotaRemaining,
            limit: filters.limit,
          }),
        });
        set({ accounts: response.accounts, loading: false, error: null });
      } catch (error) {
        const message = error instanceof APIError ? error.message : "Failed to load accounts";
        set({ loading: false, error: message });
      }
    },
  }));
}

export const useAccountsStore = createAccountsStore();
