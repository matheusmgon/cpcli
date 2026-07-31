import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { Pencil, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { DataTable } from "@/components/shared/DataTable";
import { EntityFormDialog } from "@/components/shared/EntityFormDialog";
import { ObjectPicker } from "@/components/shared/ObjectPicker";
import { PageHeader } from "@/components/shared/PageHeader";
import { RulebaseTable } from "@/components/shared/RulebaseTable";
import { getService, type JsonRecord } from "@/lib/wailsService";
import { useSessionStore } from "@/stores/session";

/** Reference fields on Check Point objects (action/protected-scope, etc.)
 * come back as a single `{name, uid}` object, an array of those, or —
 * rarely — a plain string. Normalize all three shapes to a display name. */
function refName(v: unknown): string {
  if (v === null || v === undefined) return "";
  if (Array.isArray(v)) {
    return v
      .map((item) => refName(item))
      .filter(Boolean)
      .join(", ");
  }
  if (typeof v === "object") {
    const obj = v as Record<string, unknown>;
    if (typeof obj.name === "string") return obj.name;
    if (typeof obj.uid === "string") return obj.uid;
    return "";
  }
  return String(v);
}

/** Same normalization as `refName`, but returns each object name as its own
 * array entry instead of joining them — used to seed `ObjectPicker` state
 * (which is `string[]`) when opening the edit dialog. */
function refNames(v: unknown): string[] {
  if (v === null || v === undefined) return [];
  if (Array.isArray(v)) {
    return v.flatMap((item) => refNames(item));
  }
  if (typeof v === "object") {
    const obj = v as Record<string, unknown>;
    if (typeof obj.name === "string") return [obj.name];
    if (typeof obj.uid === "string") return [obj.uid];
    return [];
  }
  const s = String(v);
  return s ? [s] : [];
}

function getErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

interface RuleFormState {
  name: string;
  position: "top" | "bottom";
  enabled: boolean;
  protectedScope: string[];
}

const emptyForm: RuleFormState = {
  name: "",
  position: "bottom",
  enabled: true,
  protectedScope: [],
};

/** Layers coming back from the Management API only expose `name`/`uid`, so
 * treat the shape loosely rather than modeling it. */
function layerLabel(layer: JsonRecord): string {
  return typeof layer.name === "string" ? layer.name : refName(layer);
}

/** Threat Prevention rulebase — layer picker, rule table (with section
 * dividers), and an add/edit dialog. Mirrors AccessRulesPage's pattern. */
function ThreatRulesTab() {
  const queryClient = useQueryClient();
  const [layerName, setLayerName] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingUid, setEditingUid] = useState<string | null>(null);
  const [form, setForm] = useState<RuleFormState>(emptyForm);

  const layersQuery = useQuery({
    queryKey: ["threat-layers"],
    queryFn: () => getService().listThreatLayers(),
  });
  const layers = layersQuery.data ?? [];

  useEffect(() => {
    if (!layerName && layers.length > 0) {
      setLayerName(layerLabel(layers[0]));
    }
  }, [layers, layerName]);

  const rulebaseQuery = useQuery({
    queryKey: ["threat-rulebase", layerName],
    queryFn: () => getService().listThreatRulebase(layerName),
    enabled: layerName.length > 0,
  });
  const rows = rulebaseQuery.data ?? [];

  function invalidateRulebase() {
    queryClient.invalidateQueries({ queryKey: ["threat-rulebase", layerName] });
  }

  const addMutation = useMutation({
    mutationFn: (fields: JsonRecord) => getService().addThreatRule(layerName, fields),
    onSuccess: () => {
      toast.success("Rule created");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Failed to create rule: ${getErrorMessage(error)}`),
  });

  const editMutation = useMutation({
    mutationFn: ({ uid, fields }: { uid: string; fields: JsonRecord }) =>
      getService().setThreatRule(layerName, uid, fields),
    onSuccess: () => {
      toast.success("Rule updated");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Failed to update rule: ${getErrorMessage(error)}`),
  });

  const deleteMutation = useMutation({
    mutationFn: (uid: string) => getService().deleteThreatRule(layerName, uid),
    onSuccess: () => {
      toast.success("Rule removed");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Failed to remove rule: ${getErrorMessage(error)}`),
  });

  const toggleEnabledMutation = useMutation({
    mutationFn: ({ uid, enabled }: { uid: string; enabled: boolean }) =>
      getService().setThreatRule(layerName, uid, { enabled }),
    onSuccess: (_data, { enabled }) => {
      toast.success(enabled ? "Rule enabled" : "Rule disabled");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Failed to change rule: ${getErrorMessage(error)}`),
  });

  function openCreateDialog() {
    setEditingUid(null);
    setForm(emptyForm);
    setDialogOpen(true);
  }

  function openEditDialog(row: JsonRecord) {
    setEditingUid(String(row.uid));
    setForm({
      name: typeof row.name === "string" ? row.name : "",
      position: "bottom",
      enabled: row.enabled !== false,
      protectedScope: refNames(row["protected-scope"]),
    });
    setDialogOpen(true);
  }

  function handleDelete(uid: string) {
    if (!window.confirm("Remove this rule?")) return;
    deleteMutation.mutate(uid);
  }

  function handleSubmit() {
    // Always send explicit "Any" for object-list fields left blank in the
    // form, instead of omitting them — keeps the UX consistent with Access
    // rules and avoids relying on server-side defaults (which vary by rule
    // type and API version).
    const protectedScope = form.protectedScope.length > 0 ? form.protectedScope : ["Any"];

    if (editingUid) {
      const fields: JsonRecord = {
        enabled: form.enabled,
        "protected-scope": protectedScope,
        ...(form.name && { name: form.name }),
      };
      editMutation.mutate({ uid: editingUid, fields });
      return;
    }

    if (!form.name.trim()) {
      toast.error("Name is required for Threat Prevention rules");
      return;
    }

    const fields: JsonRecord = {
      name: form.name,
      position: form.position,
      enabled: true,
      "protected-scope": protectedScope,
    };
    addMutation.mutate(fields);
  }

  const isSaving = addMutation.isPending || editMutation.isPending;

  return (
    <div>
      <div className="mb-4 flex items-end justify-between gap-4">
        <div className="max-w-xs flex-1">
          <Label htmlFor="threat-layer-select" className="mb-1.5 block">
            Layer
          </Label>
          <Select value={layerName} onValueChange={setLayerName}>
            <SelectTrigger id="threat-layer-select">
              <SelectValue placeholder="Select a layer" />
            </SelectTrigger>
            <SelectContent>
              {layers.map((layer) => {
                const label = layerLabel(layer);
                return (
                  <SelectItem key={label || String(layer.uid)} value={label}>
                    {label}
                  </SelectItem>
                );
              })}
            </SelectContent>
          </Select>
        </div>
        <Button onClick={openCreateDialog} disabled={!layerName}>
          <Plus className="size-4" />
          New rule
        </Button>
      </div>

      <RulebaseTable
        rows={rows}
        emptyMessage={layerName ? "This layer has no rules yet." : "Select a layer to see its rules."}
        columns={[
          { header: "#", cell: (row) => (typeof row["rule-number"] === "number" ? String(row["rule-number"]) : "") },
          {
            header: "Enabled",
            cell: (row) => {
              const uid = String(row.uid ?? "");
              const enabled = row.enabled !== false;
              return (
                <input
                  type="checkbox"
                  checked={enabled}
                  disabled={!uid || toggleEnabledMutation.isPending}
                  onChange={() => toggleEnabledMutation.mutate({ uid, enabled: !enabled })}
                  className="size-4 rounded border-border accent-accent"
                  aria-label={enabled ? "Disable rule" : "Enable rule"}
                />
              );
            },
          },
          { header: "Name", cell: (row) => (typeof row.name === "string" ? row.name : "") },
          { header: "Action", cell: (row) => refName(row.action) },
          { header: "Protected scope", cell: (row) => refName(row["protected-scope"]) },
        ]}
        renderActions={(row) => {
          const uid = String(row.uid ?? "");
          return (
            <div className="flex justify-end gap-1">
              <Button variant="ghost" size="icon" onClick={() => openEditDialog(row)} aria-label="Edit rule">
                <Pencil className="size-4" />
              </Button>
              <Button variant="ghost" size="icon" onClick={() => handleDelete(uid)} aria-label="Remove rule">
                <Trash2 className="size-4" />
              </Button>
            </div>
          );
        }}
      />

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen} modal={false}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingUid ? "Edit rule" : "New rule"}</DialogTitle>
            <DialogDescription>
              {editingUid
                ? "An empty protected scope is not changed."
                : "Name is required. An empty protected scope is equivalent to “Any”."}
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="threat-rule-name">Name</Label>
              <Input
                id="threat-rule-name"
                value={form.name}
                onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
                placeholder={editingUid ? "(optional)" : "rule name (required)"}
              />
            </div>

            {editingUid ? (
              <div className="flex items-center gap-2">
                <input
                  id="threat-rule-enabled"
                  type="checkbox"
                  checked={form.enabled}
                  onChange={(e) => setForm((prev) => ({ ...prev, enabled: e.target.checked }))}
                  className="size-4 rounded border-border accent-accent"
                />
                <Label htmlFor="threat-rule-enabled">Enabled</Label>
              </div>
            ) : (
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="threat-rule-position">Position</Label>
                <Select
                  value={form.position}
                  onValueChange={(value) => setForm((prev) => ({ ...prev, position: value as "top" | "bottom" }))}
                >
                  <SelectTrigger id="threat-rule-position">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="top">top</SelectItem>
                    <SelectItem value="bottom">bottom</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            )}

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="threat-rule-scope">Protected scope</Label>
              <ObjectPicker
                value={form.protectedScope}
                onChange={(names) => setForm((prev) => ({ ...prev, protectedScope: names }))}
                placeholder="Search objects... (empty = Any)"
                categories={["network"]}
              />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button type="button" onClick={handleSubmit} disabled={isSaving}>
              Save
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

/** Threat Prevention profiles — simple named entities (Optimized, Strict,
 * custom profiles), listed/created/removed via the generic DataTable +
 * EntityFormDialog pattern used on ObjectsPage. */
function ThreatProfilesTab() {
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);

  const queryKey = ["threat-profiles"];
  const { data, isLoading } = useQuery({
    queryKey,
    queryFn: () => getService().listThreatProfiles(),
  });

  function onMutationSettled() {
    queryClient.invalidateQueries({ queryKey });
    useSessionStore.getState().markPending(1);
  }

  const addMutation = useMutation({
    mutationFn: (fields: JsonRecord) => getService().addThreatProfile(fields),
    onSuccess: () => {
      onMutationSettled();
      toast.success("Profile created.");
    },
    onError: (error) => toast.error(`Failed to create profile: ${getErrorMessage(error)}`),
  });

  const deleteMutation = useMutation({
    mutationFn: (name: string) => getService().deleteThreatProfile(name),
    onSuccess: () => {
      onMutationSettled();
      toast.success("Profile removed.");
    },
    onError: (error) => toast.error(`Failed to remove profile: ${getErrorMessage(error)}`),
  });

  const columns = useMemo<ColumnDef<JsonRecord>[]>(
    () => [
      { accessorKey: "name", header: "Name" },
      {
        id: "actions",
        header: "Actions",
        cell: ({ row }) => {
          const name = typeof row.original.name === "string" ? row.original.name : "";
          return (
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="sm"
                className="text-danger hover:text-danger"
                onClick={() => {
                  if (window.confirm(`Delete profile "${name}"? This change stays pending until Publish.`)) {
                    deleteMutation.mutate(name);
                  }
                }}
              >
                Delete
              </Button>
            </div>
          );
        },
      },
    ],
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );

  return (
    <div>
      <div className="mb-4 flex justify-end">
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="size-4" />
          New profile
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={data ?? []}
        searchPlaceholder="Search profile..."
        emptyTitle={isLoading ? "Loading..." : "No profiles found"}
        emptyDescription={isLoading ? undefined : "Use the New profile button to create the first record."}
      />

      <EntityFormDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        title="New profile"
        fields={[{ key: "name", label: "Name", placeholder: "profile name" }]}
        onSubmit={async (values) => {
          await addMutation.mutateAsync(values);
        }}
      />
    </div>
  );
}

/** Threat Prevention — rules (per layer) and profiles, in tabs. */
export function ThreatPreventionPage() {
  return (
    <div>
      <PageHeader title="Threat Prevention" subtitle="Threat prevention rules and profiles." />

      <Tabs defaultValue="rules">
        <TabsList>
          <TabsTrigger value="rules">Rules</TabsTrigger>
          <TabsTrigger value="profiles">Profiles</TabsTrigger>
        </TabsList>
        <TabsContent value="rules">
          <ThreatRulesTab />
        </TabsContent>
        <TabsContent value="profiles">
          <ThreatProfilesTab />
        </TabsContent>
      </Tabs>
    </div>
  );
}
