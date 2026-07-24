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

/** Reference fields on Check Point objects (action/source/destination/service,
 * etc.) come back as a single `{name, uid}` object, an array of those, or —
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

function getErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

function splitCsv(value: string): string[] {
  return value
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

const ACTIONS = ["accept", "drop", "reject", "ask"] as const;
type ActionOption = (typeof ACTIONS)[number];

interface RuleFormState {
  name: string;
  action: ActionOption;
  position: "top" | "bottom";
  enabled: boolean;
  source: string;
  destination: string;
  service: string;
}

const emptyForm: RuleFormState = {
  name: "",
  action: "accept",
  position: "bottom",
  enabled: true,
  source: "",
  destination: "",
  service: "",
};

/** Layers/packages coming back from the Management API — access layers only
 * expose `name`/`uid`, so treat the shape loosely rather than modeling it. */
function layerLabel(layer: JsonRecord): string {
  return typeof layer.name === "string" ? layer.name : refName(layer);
}

/** Access Control rulebase for a single layer — layer picker, rule table
 * (with section dividers), and an add/edit dialog. */
export function AccessRulesPage() {
  const queryClient = useQueryClient();
  const [layerName, setLayerName] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingUid, setEditingUid] = useState<string | null>(null);
  const [form, setForm] = useState<RuleFormState>(emptyForm);

  const layersQuery = useQuery({
    queryKey: ["access-layers"],
    queryFn: () => getService().listAccessLayers(),
  });
  const layers = layersQuery.data ?? [];

  useEffect(() => {
    if (!layerName && layers.length > 0) {
      setLayerName(layerLabel(layers[0]));
    }
  }, [layers, layerName]);

  const rulebaseQuery = useQuery({
    queryKey: ["access-rulebase", layerName],
    queryFn: () => getService().listAccessRulebase(layerName),
    enabled: layerName.length > 0,
  });
  const rows = rulebaseQuery.data ?? [];

  function invalidateRulebase() {
    queryClient.invalidateQueries({ queryKey: ["access-rulebase", layerName] });
  }

  const addMutation = useMutation({
    mutationFn: (fields: JsonRecord) => getService().addAccessRule(layerName, fields),
    onSuccess: () => {
      toast.success("Regra criada");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao criar regra: ${getErrorMessage(error)}`),
  });

  const editMutation = useMutation({
    mutationFn: ({ uid, fields }: { uid: string; fields: JsonRecord }) =>
      getService().setAccessRule(layerName, uid, fields),
    onSuccess: () => {
      toast.success("Regra atualizada");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao atualizar regra: ${getErrorMessage(error)}`),
  });

  const deleteMutation = useMutation({
    mutationFn: (uid: string) => getService().deleteAccessRule(layerName, uid),
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
    const actionName = refName(row.action).toLowerCase();
    setForm({
      name: typeof row.name === "string" ? row.name : "",
      action: (ACTIONS as readonly string[]).includes(actionName) ? (actionName as ActionOption) : "accept",
      position: "bottom",
      enabled: row.enabled !== false,
      source: refName(row.source),
      destination: refName(row.destination),
      service: refName(row.service),
    });
    setDialogOpen(true);
  }

  function handleDelete(uid: string) {
    if (!window.confirm("Remover esta regra?")) return;
    deleteMutation.mutate(uid);
  }

  function handleSubmit() {
    if (editingUid) {
      const fields: JsonRecord = {
        enabled: form.enabled,
        action: form.action,
        ...(form.name && { name: form.name }),
        ...(form.source && { source: splitCsv(form.source) }),
        ...(form.destination && { destination: splitCsv(form.destination) }),
        ...(form.service && { service: splitCsv(form.service) }),
      };
      editMutation.mutate({ uid: editingUid, fields });
      return;
    }

    const source = splitCsv(form.source);
    const destination = splitCsv(form.destination);
    const service = splitCsv(form.service);
    const fields: JsonRecord = {
      action: form.action,
      position: form.position,
      enabled: true,
      ...(form.name && { name: form.name }),
      ...(source.length > 0 && { source }),
      ...(destination.length > 0 && { destination }),
      ...(service.length > 0 && { service }),
    };
    addMutation.mutate(fields);
  }

  const isSaving = addMutation.isPending || editMutation.isPending;

  return (
    <div>
      <PageHeader
        title="Access Control"
        subtitle="Regras de uma camada (layer)."
        actions={
          <Button onClick={openCreateDialog} disabled={!layerName}>
            <Plus className="size-4" />
            Nova regra
          </Button>
        }
      />

      <div className="mb-4 max-w-xs">
        <Label htmlFor="layer-select" className="mb-1.5 block">
          Layer
        </Label>
        <Select value={layerName} onValueChange={setLayerName}>
          <SelectTrigger id="layer-select">
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
          { header: "Serviço", cell: (row) => refName(row.service) },
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
                ? "Campos de origem/destino/serviço vazios não são alterados."
                : "Campos de origem/destino/serviço vazios equivalem a “Any”."}
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="rule-name">Nome</Label>
              <Input
                id="rule-name"
                value={form.name}
                onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
                placeholder="(opcional)"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="rule-action">Ação</Label>
              <Select
                value={form.action}
                onValueChange={(value) => setForm((prev) => ({ ...prev, action: value as ActionOption }))}
              >
                <SelectTrigger id="rule-action">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {ACTIONS.map((action) => (
                    <SelectItem key={action} value={action}>
                      {action}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {editingUid ? (
              <div className="flex items-center gap-2">
                <input
                  id="rule-enabled"
                  type="checkbox"
                  checked={form.enabled}
                  onChange={(e) => setForm((prev) => ({ ...prev, enabled: e.target.checked }))}
                  className="size-4 rounded border-border accent-accent"
                />
                <Label htmlFor="rule-enabled">Habilitada</Label>
              </div>
            ) : (
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="rule-position">Posição</Label>
                <Select
                  value={form.position}
                  onValueChange={(value) => setForm((prev) => ({ ...prev, position: value as "top" | "bottom" }))}
                >
                  <SelectTrigger id="rule-position">
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
              <Label htmlFor="rule-source">Origem</Label>
              <Input
                id="rule-source"
                value={form.source}
                onChange={(e) => setForm((prev) => ({ ...prev, source: e.target.value }))}
                placeholder="nomes separados por vírgula (vazio = Any)"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="rule-destination">Destino</Label>
              <Input
                id="rule-destination"
                value={form.destination}
                onChange={(e) => setForm((prev) => ({ ...prev, destination: e.target.value }))}
                placeholder="nomes separados por vírgula (vazio = Any)"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="rule-service">Serviço</Label>
              <Input
                id="rule-service"
                value={form.service}
                onChange={(e) => setForm((prev) => ({ ...prev, service: e.target.value }))}
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
