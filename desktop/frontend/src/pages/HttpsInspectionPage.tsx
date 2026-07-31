import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
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
import { ObjectPicker } from "@/components/shared/ObjectPicker";
import { PageHeader } from "@/components/shared/PageHeader";
import { RulebaseTable } from "@/components/shared/RulebaseTable";
import { getService, type JsonRecord } from "@/lib/wailsService";
import { useSessionStore } from "@/stores/session";

/** Reference fields on Check Point objects (action/source/destination, etc.)
 * come back as a single `{name, uid}` object, an array of those, or —
 * for https-rule `action` in particular — a plain string ("Inspect" /
 * "Bypass"). Normalize all shapes to a display name. */
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
  source: string[];
  destination: string[];
}

const emptyForm: RuleFormState = {
  name: "",
  position: "bottom",
  enabled: true,
  source: [],
  destination: [],
};

/** Layers coming back from the Management API only expose `name`/`uid`, so
 * treat the shape loosely rather than modeling it. */
function layerLabel(layer: JsonRecord): string {
  return typeof layer.name === "string" ? layer.name : refName(layer);
}

/** HTTPS Inspection rulebase for a single layer — layer picker, rule table
 * (with section dividers), and an add/edit dialog. Same pattern as
 * AccessRulesPage/ThreatPreventionPage's rules tab. */
export function HttpsInspectionPage() {
  const queryClient = useQueryClient();
  const [layerName, setLayerName] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingUid, setEditingUid] = useState<string | null>(null);
  const [form, setForm] = useState<RuleFormState>(emptyForm);

  const layersQuery = useQuery({
    queryKey: ["https-layers"],
    queryFn: () => getService().listHttpsLayers(),
  });
  const layers = layersQuery.data ?? [];

  useEffect(() => {
    if (!layerName && layers.length > 0) {
      setLayerName(layerLabel(layers[0]));
    }
  }, [layers, layerName]);

  const rulebaseQuery = useQuery({
    queryKey: ["https-rulebase", layerName],
    queryFn: () => getService().listHttpsRulebase(layerName),
    enabled: layerName.length > 0,
  });
  const rows = rulebaseQuery.data ?? [];

  function invalidateRulebase() {
    queryClient.invalidateQueries({ queryKey: ["https-rulebase", layerName] });
  }

  const addMutation = useMutation({
    mutationFn: (fields: JsonRecord) => getService().addHttpsRule(layerName, fields),
    onSuccess: () => {
      toast.success("Rule created");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Failed to create rule: ${getErrorMessage(error)}`),
  });

  const editMutation = useMutation({
    mutationFn: ({ uid, fields }: { uid: string; fields: JsonRecord }) =>
      getService().setHttpsRule(layerName, uid, fields),
    onSuccess: () => {
      toast.success("Rule updated");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Failed to update rule: ${getErrorMessage(error)}`),
  });

  const deleteMutation = useMutation({
    mutationFn: (uid: string) => getService().deleteHttpsRule(layerName, uid),
    onSuccess: () => {
      toast.success("Rule removed");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Failed to remove rule: ${getErrorMessage(error)}`),
  });

  const toggleEnabledMutation = useMutation({
    mutationFn: ({ uid, enabled }: { uid: string; enabled: boolean }) =>
      getService().setHttpsRule(layerName, uid, { enabled }),
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
      source: refNames(row.source),
      destination: refNames(row.destination),
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
    // rules and avoids relying on server-side defaults.
    const source = form.source.length > 0 ? form.source : ["Any"];
    const destination = form.destination.length > 0 ? form.destination : ["Any"];

    if (editingUid) {
      const fields: JsonRecord = {
        enabled: form.enabled,
        source,
        destination,
        ...(form.name && { name: form.name }),
      };
      editMutation.mutate({ uid: editingUid, fields });
      return;
    }

    const fields: JsonRecord = {
      position: form.position,
      enabled: true,
      source,
      destination,
      ...(form.name && { name: form.name }),
    };
    addMutation.mutate(fields);
  }

  const isSaving = addMutation.isPending || editMutation.isPending;

  return (
    <div>
      <PageHeader
        title="HTTPS Inspection"
        subtitle="HTTPS inspection rules."
        actions={
          <Button onClick={openCreateDialog} disabled={!layerName}>
            <Plus className="size-4" />
            New rule
          </Button>
        }
      />

      <div className="mb-4 max-w-xs">
        <Label htmlFor="https-layer-select" className="mb-1.5 block">
          Layer
        </Label>
        <Select value={layerName} onValueChange={setLayerName}>
          <SelectTrigger id="https-layer-select">
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
          { header: "Source", cell: (row) => refName(row.source) },
          { header: "Destination", cell: (row) => refName(row.destination) },
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
                ? "Empty source/destination fields are not changed."
                : "Name is optional. Empty source/destination fields are equivalent to “Any”."}
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="https-rule-name">Name</Label>
              <Input
                id="https-rule-name"
                value={form.name}
                onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
                placeholder="(optional)"
              />
            </div>

            {editingUid ? (
              <div className="flex items-center gap-2">
                <input
                  id="https-rule-enabled"
                  type="checkbox"
                  checked={form.enabled}
                  onChange={(e) => setForm((prev) => ({ ...prev, enabled: e.target.checked }))}
                  className="size-4 rounded border-border accent-accent"
                />
                <Label htmlFor="https-rule-enabled">Enabled</Label>
              </div>
            ) : (
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="https-rule-position">Position</Label>
                <Select
                  value={form.position}
                  onValueChange={(value) => setForm((prev) => ({ ...prev, position: value as "top" | "bottom" }))}
                >
                  <SelectTrigger id="https-rule-position">
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
              <Label htmlFor="https-rule-source">Source</Label>
              <ObjectPicker
                value={form.source}
                onChange={(names) => setForm((prev) => ({ ...prev, source: names }))}
                placeholder="Search objects... (empty = Any)"
                categories={["network"]}
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="https-rule-destination">Destination</Label>
              <ObjectPicker
                value={form.destination}
                onChange={(names) => setForm((prev) => ({ ...prev, destination: names }))}
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
