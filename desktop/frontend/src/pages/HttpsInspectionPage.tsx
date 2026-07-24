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

function getErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

function splitCsv(value: string): string[] {
  return value
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

interface RuleFormState {
  name: string;
  position: "top" | "bottom";
  enabled: boolean;
  source: string;
  destination: string;
}

const emptyForm: RuleFormState = {
  name: "",
  position: "bottom",
  enabled: true,
  source: "",
  destination: "",
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
      toast.success("Regra criada");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao criar regra: ${getErrorMessage(error)}`),
  });

  const editMutation = useMutation({
    mutationFn: ({ uid, fields }: { uid: string; fields: JsonRecord }) =>
      getService().setHttpsRule(layerName, uid, fields),
    onSuccess: () => {
      toast.success("Regra atualizada");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao atualizar regra: ${getErrorMessage(error)}`),
  });

  const deleteMutation = useMutation({
    mutationFn: (uid: string) => getService().deleteHttpsRule(layerName, uid),
    onSuccess: () => {
      toast.success("Regra removida");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao remover regra: ${getErrorMessage(error)}`),
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
      source: refName(row.source),
      destination: refName(row.destination),
    });
    setDialogOpen(true);
  }

  function handleDelete(uid: string) {
    if (!window.confirm("Remover esta regra?")) return;
    deleteMutation.mutate(uid);
  }

  function handleSubmit() {
    const source = splitCsv(form.source);
    const destination = splitCsv(form.destination);

    if (editingUid) {
      const fields: JsonRecord = {
        enabled: form.enabled,
        ...(form.name && { name: form.name }),
        ...(source.length > 0 && { source }),
        ...(destination.length > 0 && { destination }),
      };
      editMutation.mutate({ uid: editingUid, fields });
      return;
    }

    const fields: JsonRecord = {
      position: form.position,
      enabled: true,
      ...(form.name && { name: form.name }),
      ...(source.length > 0 && { source }),
      ...(destination.length > 0 && { destination }),
    };
    addMutation.mutate(fields);
  }

  const isSaving = addMutation.isPending || editMutation.isPending;

  return (
    <div>
      <PageHeader
        title="HTTPS Inspection"
        subtitle="Regras de inspeção HTTPS."
        actions={
          <Button onClick={openCreateDialog} disabled={!layerName}>
            <Plus className="size-4" />
            Nova regra
          </Button>
        }
      />

      <div className="mb-4 max-w-xs">
        <Label htmlFor="https-layer-select" className="mb-1.5 block">
          Layer
        </Label>
        <Select value={layerName} onValueChange={setLayerName}>
          <SelectTrigger id="https-layer-select">
            <SelectValue placeholder="Selecione uma layer" />
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
        emptyMessage={layerName ? "Esta layer ainda não tem regras." : "Selecione uma layer para ver as regras."}
        columns={[
          { header: "#", cell: (row) => (typeof row["rule-number"] === "number" ? String(row["rule-number"]) : "") },
          { header: "Nome", cell: (row) => (typeof row.name === "string" ? row.name : "") },
          { header: "Ação", cell: (row) => refName(row.action) },
          { header: "Origem", cell: (row) => refName(row.source) },
          { header: "Destino", cell: (row) => refName(row.destination) },
        ]}
        renderActions={(row) => {
          const uid = String(row.uid ?? "");
          return (
            <div className="flex justify-end gap-1">
              <Button variant="ghost" size="icon" onClick={() => openEditDialog(row)} aria-label="Editar regra">
                <Pencil className="size-4" />
              </Button>
              <Button variant="ghost" size="icon" onClick={() => handleDelete(uid)} aria-label="Remover regra">
                <Trash2 className="size-4" />
              </Button>
            </div>
          );
        }}
      />

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingUid ? "Editar regra" : "Nova regra"}</DialogTitle>
            <DialogDescription>
              {editingUid
                ? "Campos de origem/destino vazios não são alterados."
                : "Nome é opcional. Campos de origem/destino vazios equivalem a “Any”."}
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="https-rule-name">Nome</Label>
              <Input
                id="https-rule-name"
                value={form.name}
                onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
                placeholder="(opcional)"
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
                <Label htmlFor="https-rule-enabled">Habilitada</Label>
              </div>
            ) : (
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="https-rule-position">Posição</Label>
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
              <Label htmlFor="https-rule-source">Origem</Label>
              <Input
                id="https-rule-source"
                value={form.source}
                onChange={(e) => setForm((prev) => ({ ...prev, source: e.target.value }))}
                placeholder="nomes separados por vírgula (vazio = Any)"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="https-rule-destination">Destino</Label>
              <Input
                id="https-rule-destination"
                value={form.destination}
                onChange={(e) => setForm((prev) => ({ ...prev, destination: e.target.value }))}
                placeholder="nomes separados por vírgula (vazio = Any)"
              />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>
              Cancelar
            </Button>
            <Button type="button" onClick={handleSubmit} disabled={isSaving}>
              Salvar
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
