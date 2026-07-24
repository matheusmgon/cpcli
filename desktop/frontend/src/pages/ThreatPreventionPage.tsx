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
  protectedScope: string;
}

const emptyForm: RuleFormState = {
  name: "",
  position: "bottom",
  enabled: true,
  protectedScope: "",
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
      toast.success("Regra criada");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao criar regra: ${getErrorMessage(error)}`),
  });

  const editMutation = useMutation({
    mutationFn: ({ uid, fields }: { uid: string; fields: JsonRecord }) =>
      getService().setThreatRule(layerName, uid, fields),
    onSuccess: () => {
      toast.success("Regra atualizada");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao atualizar regra: ${getErrorMessage(error)}`),
  });

  const deleteMutation = useMutation({
    mutationFn: (uid: string) => getService().deleteThreatRule(layerName, uid),
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
      protectedScope: refName(row["protected-scope"]),
    });
    setDialogOpen(true);
  }

  function handleDelete(uid: string) {
    if (!window.confirm("Remover esta regra?")) return;
    deleteMutation.mutate(uid);
  }

  function handleSubmit() {
    const protectedScope = splitCsv(form.protectedScope);

    if (editingUid) {
      const fields: JsonRecord = {
        enabled: form.enabled,
        ...(form.name && { name: form.name }),
        ...(protectedScope.length > 0 && { "protected-scope": protectedScope }),
      };
      editMutation.mutate({ uid: editingUid, fields });
      return;
    }

    if (!form.name.trim()) {
      toast.error("Nome é obrigatório para regras de Threat Prevention");
      return;
    }

    const fields: JsonRecord = {
      name: form.name,
      position: form.position,
      enabled: true,
      ...(protectedScope.length > 0 && { "protected-scope": protectedScope }),
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
        <Button onClick={openCreateDialog} disabled={!layerName}>
          <Plus className="size-4" />
          Nova regra
        </Button>
      </div>

      <RulebaseTable
        rows={rows}
        emptyMessage={layerName ? "Esta layer ainda não tem regras." : "Selecione uma layer para ver as regras."}
        columns={[
          { header: "#", cell: (row) => (typeof row["rule-number"] === "number" ? String(row["rule-number"]) : "") },
          { header: "Nome", cell: (row) => (typeof row.name === "string" ? row.name : "") },
          { header: "Ação", cell: (row) => refName(row.action) },
          { header: "Escopo protegido", cell: (row) => refName(row["protected-scope"]) },
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
                ? "O escopo protegido vazio não é alterado."
                : "Nome é obrigatório. Escopo protegido vazio equivale a “Any”."}
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="threat-rule-name">Nome</Label>
              <Input
                id="threat-rule-name"
                value={form.name}
                onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
                placeholder={editingUid ? "(opcional)" : "nome da regra (obrigatório)"}
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
                <Label htmlFor="threat-rule-enabled">Habilitada</Label>
              </div>
            ) : (
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="threat-rule-position">Posição</Label>
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
              <Label htmlFor="threat-rule-scope">Escopo protegido</Label>
              <Input
                id="threat-rule-scope"
                value={form.protectedScope}
                onChange={(e) => setForm((prev) => ({ ...prev, protectedScope: e.target.value }))}
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
      toast.success("Profile criado.");
    },
    onError: (error) => toast.error(`Falha ao criar profile: ${getErrorMessage(error)}`),
  });

  const deleteMutation = useMutation({
    mutationFn: (name: string) => getService().deleteThreatProfile(name),
    onSuccess: () => {
      onMutationSettled();
      toast.success("Profile removido.");
    },
    onError: (error) => toast.error(`Falha ao remover profile: ${getErrorMessage(error)}`),
  });

  const columns = useMemo<ColumnDef<JsonRecord>[]>(
    () => [
      { accessorKey: "name", header: "Nome" },
      {
        id: "actions",
        header: "Ações",
        cell: ({ row }) => {
          const name = typeof row.original.name === "string" ? row.original.name : "";
          return (
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="sm"
                className="text-danger hover:text-danger"
                onClick={() => {
                  if (window.confirm(`Apagar profile "${name}"? Esta ação fica pendente até o Publish.`)) {
                    deleteMutation.mutate(name);
                  }
                }}
              >
                Apagar
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
          Novo profile
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={data ?? []}
        searchPlaceholder="Buscar profile..."
        emptyTitle={isLoading ? "Carregando..." : "Nenhum profile encontrado"}
        emptyDescription={isLoading ? undefined : "Use o botão Novo profile para criar o primeiro registro."}
      />

      <EntityFormDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        title="Novo profile"
        fields={[{ key: "name", label: "Nome", placeholder: "nome do profile" }]}
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
      <PageHeader title="Threat Prevention" subtitle="Regras e perfis de prevenção de ameaças." />

      <Tabs defaultValue="regras">
        <TabsList>
          <TabsTrigger value="regras">Regras</TabsTrigger>
          <TabsTrigger value="profiles">Profiles</TabsTrigger>
        </TabsList>
        <TabsContent value="regras">
          <ThreatRulesTab />
        </TabsContent>
        <TabsContent value="profiles">
          <ThreatProfilesTab />
        </TabsContent>
      </Tabs>
    </div>
  );
}
