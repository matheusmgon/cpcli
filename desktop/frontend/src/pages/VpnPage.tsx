import { useState } from "react";
import { useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/shared/DataTable";
import { EntityFormDialog, type EntityField } from "@/components/shared/EntityFormDialog";
import { PageHeader } from "@/components/shared/PageHeader";
import { getService, type JsonRecord } from "@/lib/wailsService";
import { useSessionStore } from "@/stores/session";

type VpnKind = "star" | "meshed";

/** Formats a Check Point API reference field — it comes back as a plain
 * string, an object with a `name`, or an array of either — into a single
 * human-readable string, joining multiple entries with ", ". */
function refName(value: unknown): string {
  if (value == null) return "";
  if (Array.isArray(value)) return value.map(refName).filter(Boolean).join(", ");
  if (typeof value === "object") {
    const name = (value as { name?: unknown }).name;
    return name != null ? String(name) : "";
  }
  return String(value);
}

function toList(value: string | undefined): string[] {
  return (value ?? "")
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

function errorMessage(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

/** VPN communities (Star or Meshed, chosen by the `:kind` route param) —
 * list + create + delete, backed by the site-to-site VPN community API. */
export function VpnPage() {
  const { kind: rawKind } = useParams<{ kind: string }>();
  const kind: VpnKind = rawKind === "star" ? "star" : "meshed";

  const queryClient = useQueryClient();
  const markPending = useSessionStore((s) => s.markPending);
  const [dialogOpen, setDialogOpen] = useState(false);

  const { data = [] } = useQuery({
    queryKey: ["vpn", kind],
    queryFn: () => getService().listVpnCommunities(kind),
  });

  const addMutation = useMutation({
    mutationFn: (values: Record<string, string>) => {
      const fields: JsonRecord = { name: values.name };
      if (kind === "star") {
        const centerGateways = toList(values["center-gateways"]);
        const satelliteGateways = toList(values["satellite-gateways"]);
        if (centerGateways.length > 0) fields["center-gateways"] = centerGateways;
        if (satelliteGateways.length > 0) fields["satellite-gateways"] = satelliteGateways;
      } else {
        const gateways = toList(values.gateways);
        if (gateways.length > 0) fields.gateways = gateways;
      }
      return getService().addVpnCommunity(kind, fields);
    },
    onSuccess: () => {
      toast.success("Community created");
      markPending(1);
      queryClient.invalidateQueries({ queryKey: ["vpn", kind] });
    },
    onError: (error: unknown) => {
      toast.error(errorMessage(error, "Failed to create community"));
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (name: string) => getService().deleteVpnCommunity(kind, name),
    onSuccess: () => {
      toast.success("Community removed");
      markPending(1);
      queryClient.invalidateQueries({ queryKey: ["vpn", kind] });
    },
    onError: (error: unknown) => {
      toast.error(errorMessage(error, "Failed to remove community"));
    },
  });

  const fields: EntityField[] =
    kind === "star"
      ? [
          { key: "name", label: "Name" },
          { key: "center-gateways", label: "Center gateways", placeholder: "gw1, gw2" },
          { key: "satellite-gateways", label: "Satellite gateways", placeholder: "gw3, gw4" },
        ]
      : [
          { key: "name", label: "Name" },
          { key: "gateways", label: "Gateways", placeholder: "gw1, gw2" },
        ];

  const actionsColumn: ColumnDef<JsonRecord> = {
    id: "actions",
    header: "",
    cell: ({ row }) => (
      <Button
        variant="ghost"
        size="icon"
        title="Delete"
        onClick={() => deleteMutation.mutate(String(row.original.name))}
      >
        <Trash2 className="size-4" />
      </Button>
    ),
  };

  const columns: ColumnDef<JsonRecord>[] =
    kind === "star"
      ? [
          { accessorKey: "name", header: "Name" },
          {
            id: "center-gateways",
            header: "Center",
            cell: ({ row }) => refName(row.original["center-gateways"]),
          },
          {
            id: "satellite-gateways",
            header: "Satellites",
            cell: ({ row }) => refName(row.original["satellite-gateways"]),
          },
          actionsColumn,
        ]
      : [
          { accessorKey: "name", header: "Name" },
          {
            id: "gateways",
            header: "Gateways",
            cell: ({ row }) => refName(row.original.gateways),
          },
          actionsColumn,
        ];

  return (
    <div>
      <PageHeader
        title={`VPN ${kind === "star" ? "Star" : "Meshed"}`}
        subtitle="Site-to-site VPN communities."
        actions={
          <Button onClick={() => setDialogOpen(true)}>
            <Plus /> Create community
          </Button>
        }
      />

      <DataTable columns={columns} data={data} searchPlaceholder="Search community..." />

      <EntityFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        title="Create community"
        fields={fields}
        onSubmit={async (values) => {
          await addMutation.mutateAsync(values);
        }}
      />
    </div>
  );
}
