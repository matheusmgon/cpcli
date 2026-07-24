import { useEffect } from "react";
import { Navigate, Outlet } from "react-router-dom";

import { Sidebar } from "@/components/layout/Sidebar";
import { Topbar } from "@/components/layout/Topbar";
import { useSessionStore } from "@/stores/session";

/** Layout-route shell: sidebar + topbar + routed content, applies the
 * light/dark theme to <html>, and guards every nested route behind an
 * active session (redirects to /login otherwise). */
export function AppShell() {
  const theme = useSessionStore((s) => s.theme);
  const connected = useSessionStore((s) => s.session.connected);

  useEffect(() => {
    document.documentElement.setAttribute("data-theme", theme);
  }, [theme]);

  if (!connected) {
    return <Navigate to="/login" replace />;
  }

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-bg text-text">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <Topbar />
        <main className="flex-1 overflow-y-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
