import { useQuery, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { RefreshCw } from "lucide-react";

import { DataTable } from "@/components/shared/DataTable";
import { PageHeader } from "@/components/shared/PageHeader";
import { Button } from "@/components/ui/button";
import { getService, type JsonRecord } from "@/lib/wailsService";

const columns: ColumnDef<JsonRecord>[] = [
  { accessorKey: "name", header: "Nome" },
  { id: "type", header: "Tipo", cell: ({ row }) => String(row.original.type ?? "") },
  {
    id: "ipv4-address",
    header: "IPv4",
    cell: ({ row }) => String(row.original["ipv4-address"] ?? ""),
  },
];

/** Read-only list of managed gateways. */
export function GatewaysPage() {
  const queryClient = useQueryClient();

  const { data = [] } = useQuery({
    queryKey: ["gateways"],
    queryFn: () => getService().listGateways(),
  });

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
    </div>
  );
}
