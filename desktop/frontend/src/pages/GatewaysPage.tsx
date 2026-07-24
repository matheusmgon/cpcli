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
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { EmptyState } from "@/components/shared/EmptyState";
import { getService, type JsonRecord } from "@/lib/wailsService";
import { useSessionStore } from "@/stores/session";

function getErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

/** Interfaces dialog for a single gateway — lists interfaces and lets the
 * user toggle anti-spoofing per interface. The Go backend
 * (`internal/mgmt/gateway.go` → `MergeGatewayInterface`, called from
 * `SetGatewayInterface`) preserves the other interfaces on write, so this
 * only ever sends the single changed field for the single interface. */
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

  const toggleMutation = useMutation({
    mutationFn: ({ ifaceName, antiSpoofing }: { ifaceName: string; antiSpoofing: boolean }) =>
      getService().setGatewayInterface(gateway ?? "", ifaceName, { "anti-spoofing": antiSpoofing }),
    onSuccess: () => {
      toast.success("Interface atualizada");
      queryClient.invalidateQueries({ queryKey });
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao atualizar interface: ${getErrorMessage(error)}`),
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Interfaces — {gateway}</DialogTitle>
          <DialogDescription>Ative ou desative o anti-spoofing por interface.</DialogDescription>
        </DialogHeader>

        {data.length === 0 ? (
          <EmptyState
            title={isLoading ? "Carregando..." : "Nenhuma interface encontrada"}
          />
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nome</TableHead>
                <TableHead>IP</TableHead>
                <TableHead>Topologia</TableHead>
                <TableHead className="text-right">Anti-spoofing</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.map((iface, idx) => {
                const ifaceName = typeof iface.name === "string" ? iface.name : String(idx);
                const checked = iface["anti-spoofing"] === true;
                return (
                  <TableRow key={ifaceName}>
                    <TableCell>{ifaceName}</TableCell>
                    <TableCell>{String(iface["ipv4-address"] ?? "")}</TableCell>
                    <TableCell>{String(iface.topology ?? "")}</TableCell>
                    <TableCell className="text-right">
                      <input
                        type="checkbox"
                        checked={checked}
                        disabled={toggleMutation.isPending}
                        onChange={(e) =>
                          toggleMutation.mutate({ ifaceName, antiSpoofing: e.target.checked })
                        }
                        className="size-4 rounded border-border accent-accent"
                        aria-label={`Anti-spoofing em ${ifaceName}`}
                      />
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
