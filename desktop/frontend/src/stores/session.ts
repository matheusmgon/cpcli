import { create } from "zustand";
import type { SessionInfo } from "@/lib/wailsService";

export type Theme = "light" | "dark";

function readStoredTheme(): Theme {
  try {
    const t = localStorage.getItem("cpcli-theme");
    return t === "light" || t === "dark" ? t : "dark";
  } catch {
    return "dark";
  }
}

interface SessionState {
  session: SessionInfo;
  pending: number;
  theme: Theme;
  setSession: (s: SessionInfo) => void;
  clearSession: () => void;
  markPending: (delta: number) => void;
  resetPending: () => void;
  toggleTheme: () => void;
}

export const useSessionStore = create<SessionState>((set) => ({
  session: { connected: false, server: "", user: "", apiVersion: "" },
  pending: 0,
  theme: readStoredTheme(),
  setSession: (s) => set({ session: s }),
  clearSession: () =>
    set({ session: { connected: false, server: "", user: "", apiVersion: "" }, pending: 0 }),
  markPending: (delta) => set((state) => ({ pending: Math.max(0, state.pending + delta) })),
  resetPending: () => set({ pending: 0 }),
  toggleTheme: () =>
    set((state) => {
      const next: Theme = state.theme === "dark" ? "light" : "dark";
      try {
        localStorage.setItem("cpcli-theme", next);
      } catch {
        /* ignore — localStorage unavailable */
      }
      return { theme: next };
    }),
}));
