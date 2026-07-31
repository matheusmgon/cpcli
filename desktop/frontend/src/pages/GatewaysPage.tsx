import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { DownloadCloud, Network, RefreshCw } from "lucide-react";
import { toast } from "sonner";

import { DataTable } from "@/components/shared/DataTable";
import { PageHeader } from "@/components/shared/PageHeader";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { EmptyState } from "@/components/shared/EmptyState";
import { getService, type JsonRecord } from "@/lib/wailsService";
import { useSessionStore } from "@/stores/session";

function getErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

/** Check Point's standard terms for these two enums (same wording
 * SmartConsole itself uses). */
const SPOOF_ACTIONS = ["prevent", "detect"] as const;
const SPOOF_TRACKING = ["none", "log", "alert"] as const;

/** Topology values accepted by set-simple-gateway's `topology` field.
 * `automatic` is what Get Interfaces sets by default, but it often leaves
 * `ip-address-behind-this-interface: "not defined"` — which then blocks
 * install-policy with "Topology information must be configured". Explicit
 * `internal`/`external` classification is what unblocks it. */
const TOPOLOGIES = ["automatic", "external", "internal"] as const;

/** Options for `topology-settings.ip-address-behind-this-interface` — only
 * meaningful when topology is `internal`. "network defined by ... ip and
 * net mask" is the safe default for a LAN interface (uses the interface's
 * own IP/mask to define what's behind it). */
const BEHIND_OPTIONS: { value: string; label: string }[] = [
  { value: "not defined", label: "não definida" },
  { value: "network defined by the interface ip and net mask", label: "rede pela máscara" },
];

/** Software blade boolean fields on a simple-gateway object — confirmed
 * live (`GetGatewayBlades`) that "firewall"/"application-control" are real
 * field names this Management Server returns; the rest follow the same
 * Check Point naming convention used throughout the Management API. */
const BLADES: { key: string; label: string }[] = [
  { key: "firewall", label: "Firewall" },
  { key: "vpn", label: "VPN" },
  { key: "application-control", label: "Application Control" },
  { key: "url-filtering", label: "URL Filtering" },
  { key: "ips", label: "IPS" },
  { key: "anti-bot", label: "Anti-Bot" },
  { key: "anti-virus", label: "Anti-Virus" },
  { key: "threat-emulation", label: "Threat Emulation" },
  { key: "threat-extraction", label: "Threat Extraction" },
  { key: "content-awareness", label: "Content Awareness" },
  { key: "identity-awareness", label: "Identity Awareness" },
  { key: "https-inspection", label: "HTTPS Inspection" },
];

interface AntiSpoofingSettings {
  action?: string;
  "spoof-tracking"?: string;
  "exclude-packets"?: boolean;
}

interface TopologySettings {
  "ip-address-behind-this-interface"?: string;
  "interface-leads-to-dmz"?: boolean;
}

/** Interfaces tab for a single gateway — lists interfaces and lets the user
 * toggle anti-spoofing and its advanced settings (action, spoof tracking)
 * per interface. The Go backend (`internal/mgmt/gateway.go` →
 * `MergeGatewayInterface`, called from `SetGatewayInterface`) preserves the
 * other interfaces on write.
 *
 * `anti-spoofing-settings` itself is a nested object the API replaces
 * wholesale rather than deep-merging (same footgun as the interfaces array
 * one level up) — so every write sends the complete settings object, built
 * from the interface's current values plus the one field being changed,
 * never just the delta. */
function InterfacesTab({ gateway, open }: { gateway: string | null; open: boolean }) {
  const queryClient = useQueryClient();
  const queryKey = ["gateway-interfaces", gateway];

  const { data = [], isLoading } = useQuery({
    queryKey,
    queryFn: () => getService().listGatewayInterfaces(gateway ?? ""),
    enabled: open && Boolean(gateway),
  });

  const updateMutation = useMutation({
    mutationFn: ({ ifaceName, fields }: { ifaceName: string; fields: JsonRecord }) =>
      getService().setGatewayInterface(gateway ?? "", ifaceName, fields),
    onSuccess: () => {
      toast.success("Interface atualizada");
      queryClient.invalidateQueries({ queryKey });
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao atualizar interface: ${getErrorMessage(error)}`),
  });

  const refreshTopologyMutation = useMutation({
    mutationFn: () => getService().refreshGatewayTopology(gateway ?? ""),
    onSuccess: () => {
      toast.success("Topologia sincronizada — interfaces relidas do gateway via SIC");
      queryClient.invalidateQueries({ queryKey });
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha em Get Interfaces: ${getErrorMessage(error)}`),
  });

  const autoClassifyMutation = useMutation({
    mutationFn: async () => {
      const svc = getService();
      for (const iface of data) {
        const ifaceName = String(iface.name);
        if (!ifaceName) continue;
        await svc.setGatewayInterface(gateway ?? "", ifaceName, {
          topology: "internal",
          "topology-settings": {
            "ip-address-behind-this-interface": "network defined by the interface ip and net mask",
            "interface-leads-to-dmz": false,
          },
        });
      }
    },
    onSuccess: () => {
      toast.success(
        "Topologia auto-classificada. Marque manualmente a interface voltada pra Internet como 'external'.",
      );
      queryClient.invalidateQueries({ queryKey });
      useSessionStore.getState().markPending(data.length);
    },
    onError: (error) => toast.error(`Falha ao auto-classificar: ${getErrorMessage(error)}`),
  });

  function toggleAntiSpoofing(ifaceName: string, checked: boolean) {
    updateMutation.mutate({ ifaceName, fields: { "anti-spoofing": checked } });
  }

  function updateSpoofSetting(
    ifaceName: string,
    current: AntiSpoofingSettings,
    patch: Partial<AntiSpoofingSettings>,
  ) {
    const settings: AntiSpoofingSettings = {
      action: current.action ?? "prevent",
      "spoof-tracking": current["spoof-tracking"] ?? "log",
      "exclude-packets": current["exclude-packets"] ?? false,
      ...patch,
    };
    updateMutation.mutate({ ifaceName, fields: { "anti-spoofing-settings": settings } });
  }

  function updateTopology(ifaceName: string, topology: string) {
    updateMutation.mutate({ ifaceName, fields: { topology } });
  }

  function updateBehindInterface(
    ifaceName: string,
    current: TopologySettings,
    value: string,
  ) {
    const settings: TopologySettings = {
      "ip-address-behind-this-interface": value,
      "interface-leads-to-dmz": current["interface-leads-to-dmz"] ?? false,
    };
    updateMutation.mutate({ ifaceName, fields: { "topology-settings": settings } });
  }

  return (
    <div>
      <p className="mb-3 text-sm text-muted">
        Ative o anti-spoofing por interface e ajuste ação/rastreamento. Se acabou de habilitar
        uma interface no Gaia (ex.: Eth2), clique em <strong>Get Interfaces</strong> abaixo para
        o management re-ler as interfaces do gateway via SIC.
      </p>

      <div className="mb-3 flex justify-end gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => autoClassifyMutation.mutate()}
          disabled={!gateway || data.length === 0 || autoClassifyMutation.isPending}
          title="Marca todas as interfaces como internal com rede pela máscara. Depois marque a Internet-facing como external."
        >
          {autoClassifyMutation.isPending ? "Classificando..." : "Auto-classify topologia"}
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => refreshTopologyMutation.mutate()}
          disabled={!gateway || refreshTopologyMutation.isPending}
        >
          <DownloadCloud className="size-4" />
          {refreshTopologyMutation.isPending ? "Sincronizando..." : "Get Interfaces"}
        </Button>
      </div>

      {data.length === 0 ? (
          <EmptyState title={isLoading ? "Carregando..." : "Nenhuma interface encontrada"} />
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nome</TableHead>
                <TableHead>IP</TableHead>
                <TableHead>Topologia</TableHead>
                <TableHead>Rede atrás</TableHead>
                <TableHead>Anti-spoof</TableHead>
                <TableHead>Ação</TableHead>
                <TableHead>Rastreamento</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.map((iface, idx) => {
                const ifaceName = typeof iface.name === "string" ? iface.name : String(idx);
                const checked = iface["anti-spoofing"] === true;
                const settings = (iface["anti-spoofing-settings"] as AntiSpoofingSettings) ?? {};
                const action = settings.action ?? "prevent";
                const tracking = settings["spoof-tracking"] ?? "log";
                const topology = typeof iface.topology === "string" ? iface.topology : "automatic";
                const topoSettings = (iface["topology-settings"] as TopologySettings) ?? {};
                const behind = topoSettings["ip-address-behind-this-interface"] ?? "not defined";
                return (
                  <TableRow key={ifaceName}>
                    <TableCell>{ifaceName}</TableCell>
                    <TableCell>{String(iface["ipv4-address"] ?? "")}</TableCell>
                    <TableCell>
                      <Select
                        value={TOPOLOGIES.includes(topology as (typeof TOPOLOGIES)[number]) ? topology : "automatic"}
                        disabled={updateMutation.isPending}
                        onValueChange={(value) => updateTopology(ifaceName, value)}
                      >
                        <SelectTrigger className="h-8 w-28">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {TOPOLOGIES.map((t) => (
                            <SelectItem key={t} value={t}>
                              {t}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </TableCell>
                    <TableCell>
                      <Select
                        value={behind}
                        disabled={topology !== "internal" || updateMutation.isPending}
                        onValueChange={(value) => updateBehindInterface(ifaceName, topoSettings, value)}
                      >
                        <SelectTrigger className="h-8 w-44">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {BEHIND_OPTIONS.map((o) => (
                            <SelectItem key={o.value} value={o.value}>
                              {o.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </TableCell>
                    <TableCell>
                      <input
                        type="checkbox"
                        checked={checked}
                        disabled={updateMutation.isPending}
                        onChange={(e) => toggleAntiSpoofing(ifaceName, e.target.checked)}
                        className="size-4 rounded border-border accent-accent"
                        aria-label={`Anti-spoofing em ${ifaceName}`}
                      />
                    </TableCell>
                    <TableCell>
                      <Select
                        value={action}
                        disabled={!checked || updateMutation.isPending}
                        onValueChange={(value) =>
                          updateSpoofSetting(ifaceName, settings, { action: value })
                        }
                      >
                        <SelectTrigger className="h-8 w-24">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {SPOOF_ACTIONS.map((a) => (
                            <SelectItem key={a} value={a}>
                              {a}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </TableCell>
                    <TableCell>
                      <Select
                        value={tracking}
                        disabled={!checked || updateMutation.isPending}
                        onValueChange={(value) =>
                          updateSpoofSetting(ifaceName, settings, { "spoof-tracking": value })
                        }
                      >
                        <SelectTrigger className="h-8 w-24">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {SPOOF_TRACKING.map((t) => (
                            <SelectItem key={t} value={t}>
                              {t}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        )}
    </div>
  );
}

/** Blades tab for a single gateway — checkbox grid to enable/disable
 * software blades via `SetGatewayBlades`. Unlike interfaces/anti-spoofing,
 * blade fields are plain top-level booleans on the simple-gateway object —
 * confirmed live that set-simple-gateway only touches the fields sent, so
 * no read-merge-write dance is needed here. */
function BladesTab({ gateway, open }: { gateway: string | null; open: boolean }) {
  const queryClient = useQueryClient();
  const queryKey = ["gateway-blades", gateway];

  const { data = {}, isLoading } = useQuery({
    queryKey,
    queryFn: () => getService().getGatewayBlades(gateway ?? ""),
    enabled: open && Boolean(gateway),
  });

  const updateMutation = useMutation({
    mutationFn: (fields: JsonRecord) => getService().setGatewayBlades(gateway ?? "", fields),
    onSuccess: () => {
      toast.success("Blade atualizada");
      queryClient.invalidateQueries({ queryKey });
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao atualizar blade: ${getErrorMessage(error)}`),
  });

  if (Object.keys(data).length === 0) {
    return <EmptyState title={isLoading ? "Carregando..." : "Nenhuma informação de blade encontrada"} />;
  }

  return (
    <div className="grid grid-cols-2 gap-x-6 gap-y-3 py-1">
      {BLADES.map((blade) => {
        const checked = data[blade.key] === true;
        return (
          <label key={blade.key} className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={checked}
              disabled={updateMutation.isPending}
              onChange={(e) => updateMutation.mutate({ [blade.key]: e.target.checked })}
              className="size-4 rounded border-border accent-accent"
              aria-label={blade.label}
            />
            {blade.label}
          </label>
        );
      })}
    </div>
  );
}

/** Drill-down dialog for a single gateway — Interfaces (anti-spoofing) and
 * Blades tabs. */
function GatewayDialog({
  gateway,
  open,
  onOpenChange,
}: {
  gateway: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-5xl">
        <DialogHeader>
          <DialogTitle>{gateway}</DialogTitle>
          <DialogDescription>Interfaces (anti-spoofing) e blades habilitadas neste gateway.</DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="interfaces">
          <TabsList>
            <TabsTrigger value="interfaces">Interfaces</TabsTrigger>
            <TabsTrigger value="blades">Blades</TabsTrigger>
          </TabsList>
          <TabsContent value="interfaces">
            <InterfacesTab gateway={gateway} open={open} />
          </TabsContent>
          <TabsContent value="blades">
            <BladesTab gateway={gateway} open={open} />
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}

/** Read-only list of managed gateways, with a drill-down dialog into each
 * gateway's interfaces (anti-spoofing toggle) and software blades. */
export function GatewaysPage() {
  const queryClient = useQueryClient();
  const [interfacesGateway, setInterfacesGateway] = useState<string | null>(null);

  const { data = [] } = useQuery({
    queryKey: ["gateways"],
    queryFn: () => getService().listGateways(),
  });

  const columns: ColumnDef<JsonRecord>[] = [
    { accessorKey: "name", header: "Nome" },
    { id: "type", header: "Tipo", cell: ({ row }) => String(row.original.type ?? "") },
    {
      id: "ipv4-address",
      header: "IPv4",
      cell: ({ row }) => String(row.original["ipv4-address"] ?? ""),
    },
    {
      id: "actions",
      header: "Ações",
      cell: ({ row }) => {
        const name = typeof row.original.name === "string" ? row.original.name : "";
        return (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setInterfacesGateway(name)}
            disabled={!name}
          >
            <Network className="size-4" />
            Ver detalhes
          </Button>
        );
      },
    },
  ];

  return (
    <div>
      <PageHeader
        title="Gateways"
        subtitle="Somente leitura."
        actions={
          <Button
            variant="outline"
            onClick={() => queryClient.invalidateQueries({ queryKey: ["gateways"] })}
          >
            <RefreshCw /> Atualizar
          </Button>
        }
      />

      <DataTable columns={columns} data={data} searchPlaceholder="Buscar gateway..." />

      <GatewayDialog
        gateway={interfacesGateway}
        open={interfacesGateway !== null}
        onOpenChange={(open) => {
          if (!open) setInterfacesGateway(null);
        }}
      />
    </div>
  );
}
