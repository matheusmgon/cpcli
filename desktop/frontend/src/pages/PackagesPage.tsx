import { useQuery, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { RefreshCw } from "lucide-react";

import { DataTable } from "@/components/shared/DataTable";
import { PageHeader } from "@/components/shared/PageHeader";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { getService, type JsonRecord } from "@/lib/wailsService";

const columns: ColumnDef<JsonRecord>[] = [
  { accessorKey: "name", header: "Name" },
  {
    id: "access",
    header: "Access",
    cell: ({ row }) => (row.original.access ? <Badge variant="success">yes</Badge> : null),
  },
  {
    id: "threat-prevention",
    header: "Threat Prevention",
    cell: ({ row }) =>
      row.original["threat-prevention"] ? <Badge variant="success">yes</Badge> : null,
  },
];

/** Read-only list of policy packages. */
export function PackagesPage() {
  const queryClient = useQueryClient();

  const { data = [] } = useQuery({
    queryKey: ["packages"],
    queryFn: () => getService().listPackages(),
  });

  return (
    <div>
      <PageHeader
        title="Policy Packages"
        subtitle="Read-only."
        actions={
          <Button
            variant="outline"
            onClick={() => queryClient.invalidateQueries({ queryKey: ["packages"] })}
          >
            <RefreshCw /> Refresh
          </Button>
        }
      />

      <DataTable columns={columns} data={data} searchPlaceholder="Search package..." />
    </div>
  );
}
