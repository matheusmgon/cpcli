import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { Network, RefreshCw } from "lucide-react";
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

interface AntiSpoofingSettings {
  action?: string;
  "spoof-tracking"?: string;
  "exclude-packets"?: boolean;
}

/** Interfaces dialog for a single gateway — lists interfaces and lets the
 * user toggle anti-spoofing and its advanced settings (action, spoof
 * tracking) per interface. The Go backend (`internal/mgmt/gateway.go` →
 * `MergeGatewayInterface`, called from `SetGatewayInterface`) preserves the
 * other interfaces on write.
 *
 * `anti-spoofing-settings` itself is a nested object the API replaces
 * wholesale rather than deep-merging (same footgun as the interfaces array
 * one level up) — so every write sends the complete settings object, built
 * from the interface's current values plus the one field being changed,
 * never just the delta. */
function InterfacesDialog({
  gateway,
  open,
  onOpenChange,
}: {
  gateway: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
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

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Interfaces — {gateway}</DialogTitle>
          <DialogDescription>
            Ative o anti-spoofing por interface e ajuste ação/rastreamento. Interfaces sem IP
            configurado no Gaia (estado "off") não aparecem aqui — isso se resolve na própria
            appliance (clish/WebUI do Gaia), não pela Management API.
          </DialogDescription>
        </DialogHeader>

        {data.length === 0 ? (
          <EmptyState title={isLoading ? "Carregando..." : "Nenhuma interface encontrada"} />
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nome</TableHead>
                <TableHead>IP</TableHead>
                <TableHead>Anti-spoofing</TableHead>
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
                return (
                  <TableRow key={ifaceName}>
                    <TableCell>{ifaceName}</TableCell>
                    <TableCell>{String(iface["ipv4-address"] ?? "")}</TableCell>
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
                        <SelectTrigger className="h-8 w-28">
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
                        <SelectTrigger className="h-8 w-28">
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
      </DialogContent>
    </Dialog>
  );
}

/** Read-only list of managed gateways, with a drill-down dialog into each
 * gateway's interfaces (anti-spoofing toggle). */
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
            Ver interfaces
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

      <InterfacesDialog
        gateway={interfacesGateway}
        open={interfacesGateway !== null}
        onOpenChange={(open) => {
          if (!open) setInterfacesGateway(null);
        }}
      />
    </div>
  );
}
