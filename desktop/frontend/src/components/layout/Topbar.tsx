import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Moon, Sun, UploadCloud, Undo2, LogOut } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { getService } from "@/lib/wailsService";
import { useSessionStore } from "@/stores/session";

/** Connection pill + pending-changes badge + theme toggle + Publish/Discard
 * (mirrors SmartConsole's install-policy flow) + logout. */
export function Topbar() {
  const navigate = useNavigate();
  const session = useSessionStore((s) => s.session);
  const pending = useSessionStore((s) => s.pending);
  const theme = useSessionStore((s) => s.theme);
  const toggleTheme = useSessionStore((s) => s.toggleTheme);
  const resetPending = useSessionStore((s) => s.resetPending);
  const clearSession = useSessionStore((s) => s.clearSession);
  const [busy, setBusy] = useState<"publish" | "discard" | null>(null);

  async function handlePublish() {
    setBusy("publish");
    try {
      await getService().publish();
      resetPending();
      toast.success("Changes published successfully");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to publish");
    } finally {
      setBusy(null);
    }
  }

  async function handleDiscard() {
    setBusy("discard");
    try {
      await getService().discard();
      resetPending();
      toast.success("Changes discarded");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to discard");
    } finally {
      setBusy(null);
    }
  }

  async function handleLogout() {
    try {
      await getService().logout();
    } catch {
      /* best-effort — still clear local session and navigate away */
    }
    clearSession();
    navigate("/login");
  }

  return (
    <header className="flex h-14 shrink-0 items-center justify-between border-b border-border bg-surface px-5">
      <div className="flex items-center gap-2 rounded-full border border-border bg-surface-2 px-3 py-1 text-sm">
        <span className="size-2 rounded-full bg-ok" />
        <span className="font-medium text-text">{session.server || "disconnected"}</span>
        {session.user && <span className="text-muted">· {session.user}</span>}
      </div>

      <div className="flex items-center gap-2">
        {pending > 0 && (
          <Badge variant="warning">
            {pending} pending change{pending === 1 ? "" : "s"}
          </Badge>
        )}

        <Button variant="ghost" size="icon" onClick={toggleTheme} title="Toggle theme">
          {theme === "dark" ? <Sun className="size-4" /> : <Moon className="size-4" />}
        </Button>

        <Button
          variant="outline"
          size="sm"
          onClick={handleDiscard}
          disabled={pending === 0 || busy !== null}
        >
          <Undo2 className="size-4" />
          Discard
        </Button>

        <Button size="sm" onClick={handlePublish} disabled={pending === 0 || busy !== null}>
          <UploadCloud className="size-4" />
          Publish
        </Button>

        <Button variant="ghost" size="icon" onClick={handleLogout} title="Log out">
          <LogOut className="size-4" />
        </Button>
      </div>
    </header>
  );
}
