import { create } from "zustand";

const TOKEN_KEY = "sboard_token";

type AuthState = {
  token: string | null;
  setToken: (token: string) => void;
  clearToken: () => void;
};

function readTokenFromStorage(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY);
  } catch {
    return null;
  }
}

export const useAuthStore = create<AuthState>((set) => ({
  token: readTokenFromStorage(),
  setToken: (token) => {
    localStorage.setItem(TOKEN_KEY, token);
    set({ token });
  },
  clearToken: () => {
    localStorage.removeItem(TOKEN_KEY);
    set({ token: null });
  },
}));

export function resetAuthStore() {
  useAuthStore.setState({ token: readTokenFromStorage() });
}
