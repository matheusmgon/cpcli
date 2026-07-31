import { useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/shared/DataTable";
import { EntityFormDialog } from "@/components/shared/EntityFormDialog";
import { PageHeader } from "@/components/shared/PageHeader";
import { objectKinds } from "@/config/objectKinds";
import { getService, type JsonRecord } from "@/lib/wailsService";
import { useSessionStore } from "@/stores/session";

function toText(value: unknown): string {
  if (value === null || value === undefined) return "";
  return String(value);
}

/** Generic object list/CRUD screen — one route (`/objects/:kind`) serves all
 * 8 "simple object" kinds; per-kind fields/columns come from
 * `config/objectKinds.ts`. */
export function ObjectsPage() {
  const { kind = "" } = useParams<{ kind: string }>();
  const config = objectKinds[kind];
  const queryClient = useQueryClient();
  const markPending = useSessionStore((s) => s.markPending);

  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<JsonRecord | null>(null);

  const queryKey = ["objects", kind];

  const { data, isLoading } = useQuery({
    queryKey,
    queryFn: () => getService().listObjects(kind, ""),
    enabled: Boolean(kind && config),
  });

  function onMutationSettled() {
    queryClient.invalidateQueries({ queryKey });
    markPending(1);
  }

  const addMutation = useMutation({
    mutationFn: (fields: JsonRecord) => getService().addObject(kind, fields),
    onSuccess: () => {
      onMutationSettled();
      toast.success("Object created.");
    },
    onError: (error: Error) => toast.error(`Failed to create object: ${error.message}`),
  });

  const editMutation = useMutation({
    mutationFn: (fields: JsonRecord) => getService().setObject(kind, fields),
    onSuccess: () => {
      onMutationSettled();
      toast.success("Object updated.");
    },
    onError: (error: Error) => toast.error(`Failed to update object: ${error.message}`),
  });

  const deleteMutation = useMutation({
    mutationFn: (name: string) => getService().deleteObject(kind, name),
    onSuccess: () => {
      onMutationSettled();
      toast.success("Object removed.");
    },
    onError: (error: Error) => toast.error(`Failed to remove object: ${error.message}`),
  });

  const canEdit = (config?.fields.length ?? 0) > 1;

  const columns = useMemo<ColumnDef<JsonRecord>[]>(() => {
    if (!config) return [];

    const dataColumns: ColumnDef<JsonRecord>[] = config.columns.map((col, idx) => ({
      id: `col-${idx}`,
      header: col.header,
      accessorFn: (row: JsonRecord) => col.accessor(row),
    }));

    const actionsColumn: ColumnDef<JsonRecord> = {
      id: "actions",
      header: "Actions",
      cell: ({ row }) => {
        const record = row.original;
        const name = toText(record.name);
        return (
          <div className="flex items-center gap-2">
            {canEdit && (
              <Button variant="outline" size="sm" onClick={() => setEditing(record)}>
                Edit
              </Button>
            )}
            <Button
              variant="ghost"
              size="sm"
              className="text-danger hover:text-danger"
              onClick={() => {
                if (window.confirm(`Delete "${name}"? This change stays pending until Publish.`)) {
                  deleteMutation.mutate(name);
                }
              }}
            >
              Delete
            </Button>
          </div>
        );
      },
    };

    return [...dataColumns, actionsColumn];
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [config, canEdit]);

  if (!config) {
    return (
      <div>
        <PageHeader title="Unknown object type" subtitle={`Objects: ${kind}`} />
      </div>
    );
  }

  const editFields = config.fields.filter((field) => field.key !== "name");
  const editInitial = editing
    ? Object.fromEntries(config.fields.map((field) => [field.key, toText(editing[field.key])]))
    : undefined;

  return (
    <div>
      <PageHeader
        title={config.title}
        subtitle="Changes stay pending until Publish."
        actions={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus /> Add
          </Button>
        }
      />

      <DataTable
        columns={columns}
        data={data ?? []}
        searchPlaceholder={`Search in ${config.title.toLowerCase()}...`}
        emptyTitle={isLoading ? "Loading..." : `No items found in ${config.title}`}
        emptyDescription={
          isLoading ? undefined : "Use the Add button to create the first record."
        }
      />

      <EntityFormDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        title={`Add — ${config.title}`}
        fields={config.fields}
        onSubmit={async (values) => {
          await addMutation.mutateAsync(values);
        }}
      />

      <EntityFormDialog
        open={editing !== null}
        onOpenChange={(open) => {
          if (!open) setEditing(null);
        }}
        title={`Edit — ${toText(editing?.name)}`}
        description="The name cannot be changed."
        fields={editFields}
        initial={editInitial}
        onSubmit={async (values) => {
          if (!editing) return;
          await editMutation.mutateAsync({ name: toText(editing.name), ...values });
        }}
      />
    </div>
  );
}
