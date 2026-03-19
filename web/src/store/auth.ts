import { create } from "zustand";

const TOKEN_KEY = "sboard_token";
const TOKEN_EXPIRES_AT_KEY = "sboard_token_expires_at";

const DEFAULT_TOKEN_TTL_MS = 24 * 60 * 60 * 1000;

type AuthState = {
  token: string | null;
  expiresAt: string | null;
  setToken: (token: string, expiresAt?: string) => void;
  clearToken: () => void;
};

type PersistedAuth = {
  token: string | null;
  expiresAt: string | null;
};

function clearPersistedAuth() {
  try {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(TOKEN_EXPIRES_AT_KEY);
  } catch {
    // Ignore storage access failures and keep in-memory state authoritative.
  }
}

function defaultExpiresAt(): string {
  return new Date(Date.now() + DEFAULT_TOKEN_TTL_MS).toISOString();
}

export function isTokenExpired(expiresAt: string | null | undefined): boolean {
  if (!expiresAt) {
    return true;
  }

  const parsed = Date.parse(expiresAt);
  if (Number.isNaN(parsed)) {
    return true;
  }

  return parsed <= Date.now();
}

function readAuthFromStorage(): PersistedAuth {
  try {
    const token = localStorage.getItem(TOKEN_KEY);
    const expiresAt = localStorage.getItem(TOKEN_EXPIRES_AT_KEY);

    if (!token) {
      if (expiresAt) {
        clearPersistedAuth();
      }
      return { token: null, expiresAt: null };
    }

    if (isTokenExpired(expiresAt)) {
      clearPersistedAuth();
      return { token: null, expiresAt: null };
    }

    return { token, expiresAt };
  } catch {
    return { token: null, expiresAt: null };
  }
}

export function getValidAuthSnapshot(): PersistedAuth {
  const snapshot = useAuthStore.getState();
  if (!snapshot.token) {
    return { token: null, expiresAt: null };
  }

  if (isTokenExpired(snapshot.expiresAt)) {
    snapshot.clearToken();
    return { token: null, expiresAt: null };
  }

  return { token: snapshot.token, expiresAt: snapshot.expiresAt };
}

export const useAuthStore = create<AuthState>((set) => ({
  ...readAuthFromStorage(),
  setToken: (token, expiresAt = defaultExpiresAt()) => {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(TOKEN_EXPIRES_AT_KEY, expiresAt);
    set({ token, expiresAt });
  },
  clearToken: () => {
    clearPersistedAuth();
    set({ token: null, expiresAt: null });
  },
}));

export function resetAuthStore() {
  useAuthStore.setState(readAuthFromStorage());
}
