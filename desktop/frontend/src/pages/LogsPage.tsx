import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { ChevronDown, ChevronRight, RefreshCw, Search } from "lucide-react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { EmptyState } from "@/components/shared/EmptyState";
import { PageHeader } from "@/components/shared/PageHeader";
import { getService, type JsonRecord } from "@/lib/wailsService";

function getErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

function actionBadge(action: string) {
  const a = action.toLowerCase();
  if (a === "accept")
    return <Badge className="bg-emerald-600 px-2 py-0.5 text-xs font-semibold uppercase text-white">accept</Badge>;
  if (a === "drop" || a === "reject")
    return <Badge className="bg-rose-600 px-2 py-0.5 text-xs font-semibold uppercase text-white">{action}</Badge>;
  return <Badge variant="outline" className="px-2 py-0.5 text-xs uppercase">{action}</Badge>;
}

/** Compose the NAT-translated tuple `xlatesrc:xlatesport -> xlatedst:xlatedport`
 * when any of those fields are present; empty string otherwise. */
function natString(r: JsonRecord): string {
  const xs = String(r.xlatesrc ?? "");
  const xd = String(r.xlatedst ?? "");
  const xsp = String(r.xlatesport ?? "");
  const xdp = String(r.xlatedport ?? "");
  if (!xs && !xd) return "";
  const left = xs + (xsp ? ":" + xsp : "");
  const right = xd + (xdp ? ":" + xdp : "");
  return `${left} → ${right}`;
}

/** Some fw log entries aren't per-packet events but periodic rule/NAT hit
 * counters or policy accounting records — they carry `hit:`, `policy:`,
 * `first_hit_time:` etc. but no `src`/`dst`. Detect those so the UI can
 * label them explicitly instead of showing an all-blank row. */
function isCounterEntry(r: JsonRecord): boolean {
  const hasNoTraffic = !r.src && !r.dst;
  const hasCounter = Boolean(r.hit || r.first_hit_time || r.last_hit_time);
  return hasNoTraffic && hasCounter;
}

/** Turn the raw fw log line into a list of `{key, value}` pairs for the
 * expanded detail panel. Splits at `LogId:` (kept as key), then walks the
 * `key: value;` pairs. */
function parseRawPairs(raw: string): { key: string; value: string }[] {
  const out: { key: string; value: string }[] = [];
  const idx = raw.indexOf("LogId:");
  if (idx < 0) return out;
  const tail = raw.slice(idx);
  for (const chunk of tail.split(";")) {
    const s = chunk.trim();
    if (!s) continue;
    const eq = s.indexOf(":");
    if (eq < 0) continue;
    const k = s.slice(0, eq).trim();
    const v = s.slice(eq + 1).trim();
    if (k && v) out.push({ key: k, value: v });
  }
  return out;
}

/** Firewall log viewer — reads the last N lines of `fw log` from a chosen
 * gateway via the Management API's run-script mechanism (bypasses the
 * `show-logs` command, which needs a Smart-1 log server this Standalone
 * lab doesn't have). Explicit "Read logs" button — we don't auto-poll,
 * since run-script + fw log on a busy appliance can take a few seconds. */
export function LogsPage() {
  const [gateway, setGateway] = useState("");
  const [filter, setFilter] = useState("");
  const [limit, setLimit] = useState(100);
  const [logs, setLogs] = useState<JsonRecord[]>([]);
  const [expandedIdx, setExpandedIdx] = useState<number | null>(null);
  useMemo(() => setExpandedIdx(null), [logs]);

  const gatewaysQuery = useQuery({
    queryKey: ["gateways"],
    queryFn: () => getService().listGateways(),
  });
  const gateways = gatewaysQuery.data ?? [];

  useEffect(() => {
    if (!gateway && gateways.length > 0) {
      const first = gateways.find((g) => /gateway/i.test(String(g.type ?? ""))) ?? gateways[0];
      const name = typeof first.name === "string" ? first.name : "";
      if (name) setGateway(name);
    }
  }, [gateways, gateway]);

  const readMutation = useMutation({
    mutationFn: () => getService().readFirewallLogs(gateway, filter, limit),
    onSuccess: (data) => {
      setLogs(data);
      toast.success(`${data.length} entries loaded`);
    },
    onError: (error) => toast.error(`Failed to read logs: ${getErrorMessage(error)}`),
  });

  return (
    <div>
      <PageHeader
        title="Logs & Monitor"
        subtitle="Reads fw log directly from the gateway via SIC (run-script). The standard show-logs needs a Smart-1 log server, which this standalone lab does not have."
      />

      <div className="mb-4 flex flex-wrap items-end gap-3">
        <div className="flex min-w-[200px] flex-col gap-1.5">
          <Label htmlFor="log-gateway">Gateway</Label>
          <Select value={gateway} onValueChange={setGateway}>
            <SelectTrigger id="log-gateway">
              <SelectValue placeholder="Select" />
            </SelectTrigger>
            <SelectContent>
              {gateways.map((g) => {
                const name = typeof g.name === "string" ? g.name : "";
                return (
                  <SelectItem key={name || String(g.uid)} value={name}>
                    {name}
                  </SelectItem>
                );
              })}
            </SelectContent>
          </Select>
        </div>

        <div className="flex min-w-[260px] flex-1 flex-col gap-1.5">
          <Label htmlFor="log-filter">Filter (grep over the raw line)</Label>
          <Input
            id="log-filter"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            placeholder='e.g. "10.0.10.10" or "drop"'
            onKeyDown={(e) => {
              if (e.key === "Enter" && gateway) readMutation.mutate();
            }}
          />
        </div>

        <div className="flex w-24 flex-col gap-1.5">
          <Label htmlFor="log-limit">Limit</Label>
          <Input
            id="log-limit"
            type="number"
            min={10}
            max={500}
            value={limit}
            onChange={(e) => setLimit(Number(e.target.value) || 100)}
          />
        </div>

        <Button onClick={() => readMutation.mutate()} disabled={!gateway || readMutation.isPending}>
          {readMutation.isPending ? (
            <>
              <RefreshCw className="size-4 animate-spin" />
              Reading...
            </>
          ) : (
            <>
              <Search className="size-4" />
              Read logs
            </>
          )}
        </Button>
      </div>

      {logs.length === 0 ? (
        <EmptyState
          title={readMutation.isPending ? "Loading..." : "No logs loaded yet"}
          description="Click 'Read logs' to fetch the last N lines of fw log from the gateway. Filter by IP, rule, action, etc."
        />
      ) : (
        <>
        <div className="overflow-x-auto rounded-md border border-border">
          <Table>
            <TableHeader>
              <TableRow className="bg-surface-2">
                <TableHead className="w-8"></TableHead>
                <TableHead className="w-24">Time</TableHead>
                <TableHead className="w-28">Action</TableHead>
                <TableHead className="min-w-[140px]">Source</TableHead>
                <TableHead className="min-w-[140px]">Destination</TableHead>
                <TableHead className="min-w-[120px]">Service</TableHead>
                <TableHead className="w-20">Proto</TableHead>
                <TableHead className="w-24">Interface</TableHead>
                <TableHead className="min-w-[160px]">Rule</TableHead>
                <TableHead className="min-w-[220px]">NAT (translated)</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.map((r, i) => {
                const nat = natString(r);
                const counter = isCounterEntry(r);
                const isOpen = expandedIdx === i;
                const raw = typeof r.raw === "string" ? r.raw : "";
                return (
                  <>
                    <TableRow
                      key={`row-${i}`}
                      onClick={() => setExpandedIdx(isOpen ? null : i)}
                      className={`cursor-pointer hover:bg-surface-2 ${isOpen ? "bg-surface-2" : ""}`}
                    >
                      <TableCell className="text-muted">
                        {isOpen ? <ChevronDown className="size-4" /> : <ChevronRight className="size-4" />}
                      </TableCell>
                      <TableCell className="font-mono text-sm tabular-nums">{String(r.time ?? "")}</TableCell>
                      <TableCell>{actionBadge(String(r.action ?? ""))}</TableCell>
                      {counter ? (
                        <TableCell colSpan={5} className="text-sm italic text-muted">
                          <Badge variant="outline" className="mr-2 px-2 py-0.5 text-xs uppercase">counter</Badge>
                          Statistical record ({String(r.policy ?? "policy")} — hits={String(r.hit ?? "?")}). No src/dst
                          because this is not a packet event.
                        </TableCell>
                      ) : (
                        <>
                          <TableCell className="font-mono text-sm">{String(r.src ?? "")}</TableCell>
                          <TableCell className="font-mono text-sm">{String(r.dst ?? "")}</TableCell>
                          <TableCell className="text-sm">{String(r.service_id ?? r.svc ?? "")}</TableCell>
                          <TableCell className="text-sm uppercase">{String(r.proto ?? "")}</TableCell>
                          <TableCell className="font-mono text-sm">{String(r.iface ?? "")}</TableCell>
                        </>
                      )}
                      <TableCell className="text-sm font-medium">{String(r.rule_name ?? "")}</TableCell>
                      <TableCell className="font-mono text-sm text-accent">{nat || "—"}</TableCell>
                    </TableRow>
                    {isOpen && (
                      <TableRow key={`det-${i}`} className="bg-surface/40">
                        <TableCell colSpan={10} className="p-4">
                          <div className="rounded-md border border-border bg-surface p-4">
                            <div className="mb-2 flex items-center justify-between">
                              <span className="text-xs font-semibold uppercase text-muted">Log details</span>
                              <button
                                type="button"
                                className="text-xs text-accent hover:underline"
                                onClick={() => setExpandedIdx(null)}
                              >
                                close
                              </button>
                            </div>
                            <dl className="grid grid-cols-1 gap-x-6 gap-y-1 md:grid-cols-2">
                              {parseRawPairs(raw).map((p) => (
                                <div key={p.key} className="flex gap-2 border-b border-border/50 py-1 last:border-b-0">
                                  <dt className="min-w-[140px] shrink-0 text-xs font-medium text-muted">{p.key}</dt>
                                  <dd className="break-all font-mono text-xs text-text">{p.value}</dd>
                                </div>
                              ))}
                            </dl>
                          </div>
                        </TableCell>
                      </TableRow>
                    )}
                  </>
                );
              })}
            </TableBody>
          </Table>
        </div>

        <p className="mt-2 text-xs text-muted">
          Click a row to see the full breakdown below (all key:value pairs from fw log).
          Rows marked <span className="rounded bg-surface-2 px-1">counter</span> are periodic
          statistical records for rules, not traffic events.
        </p>
        </>
      )}
    </div>
  );
}
