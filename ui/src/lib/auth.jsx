import { useQueryClient } from "@tanstack/react-query";
import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { setUnauthorizedHandler } from "./api";

const STORAGE_KEY = "iwx-ui-session";
const AuthContext = createContext(null);

function decodeJwtPayload(token) {
  if (typeof window === "undefined") return null;

  const parts = String(token || "").split(".");
  if (parts.length < 2) return null;

  try {
    const base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, "=");
    return JSON.parse(window.atob(padded));
  } catch (_error) {
    return null;
  }
}

function sessionExpiryMs(session) {
  const expiresAtMs = Date.parse(String(session?.expiresAt || ""));
  if (Number.isFinite(expiresAtMs)) {
    return expiresAtMs;
  }

  const payload = decodeJwtPayload(session?.accessToken);
  const expSeconds = Number(payload?.exp);
  if (Number.isFinite(expSeconds) && expSeconds > 0) {
    return expSeconds * 1000;
  }

  return null;
}

function isSessionExpired(session) {
  const expiresAtMs = sessionExpiryMs(session);
  return Number.isFinite(expiresAtMs) ? expiresAtMs <= Date.now() : false;
}

function readStoredSession() {
  if (typeof window === "undefined") return null;

  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const session = JSON.parse(raw);
    return isSessionExpired(session) ? null : session;
  } catch (_error) {
    return null;
  }
}

export function AuthProvider({ children }) {
  const queryClient = useQueryClient();
  const [session, setSession] = useState(readStoredSession);

  useEffect(() => {
    if (!session || !isSessionExpired(session)) return;
    setSession(null);
  }, [session]);

  useEffect(() => {
    if (typeof window === "undefined") return;

    if (session) {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
      return;
    }

    window.localStorage.removeItem(STORAGE_KEY);
  }, [session]);

  useEffect(() => {
    setUnauthorizedHandler(() => {
      setSession((current) => (current ? null : current));
      queryClient.clear();
    });

    return () => {
      setUnauthorizedHandler(null);
    };
  }, [queryClient]);

  useEffect(() => {
    if (!session) return undefined;

    const expiresAtMs = sessionExpiryMs(session);
    if (!Number.isFinite(expiresAtMs)) return undefined;

    const delayMs = Math.max(expiresAtMs - Date.now(), 0);
    const timeoutId = window.setTimeout(() => {
      setSession(null);
      queryClient.clear();
    }, delayMs);

    return () => window.clearTimeout(timeoutId);
  }, [queryClient, session]);

  const logout = () => {
    setSession(null);
    queryClient.clear();
  };

  const value = useMemo(
    () => ({
      session,
      token: session?.accessToken ?? null,
      isAuthenticated: Boolean(session?.accessToken),
      login(nextSession) {
        setSession(isSessionExpired(nextSession) ? null : nextSession);
      },
      logout,
    }),
    [logout, session],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error("useAuth must be used inside AuthProvider");
  }

  return value;
}
