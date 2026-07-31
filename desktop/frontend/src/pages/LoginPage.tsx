import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { Moon, ShieldCheck, Sun } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { getService } from "@/lib/wailsService";
import { useSessionStore } from "@/stores/session";

/** Faithful port of the existing login screen: server/user/port/password +
 * an "ignore TLS verification" checkbox, centered card, theme toggle. */
export function LoginPage() {
  const navigate = useNavigate();
  const theme = useSessionStore((s) => s.theme);
  const toggleTheme = useSessionStore((s) => s.toggleTheme);
  const setSession = useSessionStore((s) => s.setSession);

  const [server, setServer] = useState("");
  const [user, setUser] = useState("admin");
  const [port, setPort] = useState("443");
  const [password, setPassword] = useState("");
  const [insecure, setInsecure] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    try {
      const session = await getService().login({
        server,
        port: Number(port) || 443,
        user,
        password,
        apiKey: "",
        domain: "",
        readOnly: false,
        insecure,
      });
      setSession(session);
      navigate("/dashboard");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Falha ao conectar");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="flex h-screen w-screen items-center justify-center bg-bg text-text" data-theme={theme}>
      <Button
        variant="ghost"
        size="icon"
        onClick={toggleTheme}
        className="absolute right-6 top-6"
        title="Alternar tema"
      >
        {theme === "dark" ? <Sun className="size-4" /> : <Moon className="size-4" />}
      </Button>

      <div className="w-full max-w-sm rounded-lg border border-border bg-surface p-8 shadow-[var(--shadow)]">
        <div className="mb-6 flex flex-col items-center gap-2 text-center">
          <div className="flex size-11 items-center justify-center rounded-full bg-accent-soft">
            <ShieldCheck className="size-5 text-accent" />
          </div>
          <h1 className="text-lg font-semibold tracking-tight">CheckPoint Console</h1>
          <p className="text-sm text-muted">Conecte-se ao Security Management Server</p>
        </div>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <Label htmlFor="server">Servidor</Label>
            <Input
              id="server"
              value={server}
              onChange={(e) => setServer(e.target.value)}
              placeholder="192.168.56.10"
              required
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="user">Usuário</Label>
              <Input
                id="user"
                value={user}
                onChange={(e) => setUser(e.target.value)}
                placeholder="admin"
                required
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="port">Porta</Label>
              <Input
                id="port"
                type="number"
                value={port}
                onChange={(e) => setPort(e.target.value)}
                placeholder="443"
                required
              />
            </div>
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor="password">Senha</Label>
            <Input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              required
            />
          </div>

          <label className="flex items-center gap-2 text-sm text-muted">
            <input
              type="checkbox"
              checked={insecure}
              onChange={(e) => setInsecure(e.target.checked)}
              className="size-3.5 rounded border-border accent-accent"
            />
            Ignorar verificação TLS
          </label>

          <Button type="submit" className="mt-2" disabled={submitting}>
            {submitting ? "Conectando..." : "Conectar"}
          </Button>
        </form>
      </div>
    </div>
  );
}
