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

/** Reference fields (original-source, method, etc.) come back as a single
 * `{name, uid}` object, an array of those, or — rarely — a plain string.
 * Normalize all three shapes to a display name. */
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

function pkgLabel(pkg: JsonRecord): string {
  return typeof pkg.name === "string" ? pkg.name : refName(pkg);
}

const METHODS = ["static", "hide"] as const;
type MethodOption = (typeof METHODS)[number];

interface NatFormState {
  originalSource: string[];
  originalDestination: string[];
  originalService: string[];
  translatedSource: string[];
  translatedDestination: string[];
  translatedService: string[];
  method: MethodOption;
  position: "top" | "bottom";
  enabled: boolean;
}

const emptyForm: NatFormState = {
  originalSource: [],
  originalDestination: [],
  originalService: [],
  translatedSource: [],
  translatedDestination: [],
  translatedService: [],
  method: "hide",
  position: "bottom",
  enabled: true,
};

/** NAT rulebase for a single policy package — package picker, rule table
 * (with section dividers), and an add/edit dialog. */
export function NatPage() {
  const queryClient = useQueryClient();
  const [pkgName, setPkgName] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingUid, setEditingUid] = useState<string | null>(null);
  const [form, setForm] = useState<NatFormState>(emptyForm);

  const packagesQuery = useQuery({
    queryKey: ["packages"],
    queryFn: () => getService().listPackages(),
  });
  const packages = packagesQuery.data ?? [];

  useEffect(() => {
    if (!pkgName && packages.length > 0) {
      setPkgName(pkgLabel(packages[0]));
    }
  }, [packages, pkgName]);

  const natQuery = useQuery({
    queryKey: ["nat-rulebase", pkgName],
    queryFn: () => getService().listNatRulebase(pkgName),
    enabled: pkgName.length > 0,
  });
  const rows = natQuery.data ?? [];

  function invalidateRulebase() {
    queryClient.invalidateQueries({ queryKey: ["nat-rulebase", pkgName] });
  }

  const addMutation = useMutation({
    mutationFn: (fields: JsonRecord) => getService().addNatRule(pkgName, fields),
    onSuccess: () => {
      toast.success("Regra NAT criada");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao criar regra NAT: ${getErrorMessage(error)}`),
  });

  const editMutation = useMutation({
    mutationFn: ({ uid, fields }: { uid: string; fields: JsonRecord }) =>
      getService().setNatRule(pkgName, uid, fields),
    onSuccess: () => {
      toast.success("Regra NAT atualizada");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao atualizar regra NAT: ${getErrorMessage(error)}`),
  });

  const deleteMutation = useMutation({
    mutationFn: (uid: string) => getService().deleteNatRule(pkgName, uid),
    onSuccess: () => {
      toast.success("Regra NAT removida");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao remover regra NAT: ${getErrorMessage(error)}`),
  });

  const toggleEnabledMutation = useMutation({
    mutationFn: ({ uid, enabled }: { uid: string; enabled: boolean }) =>
      getService().setNatRule(pkgName, uid, { enabled }),
    onSuccess: (_data, { enabled }) => {
      toast.success(enabled ? "Regra NAT ativada" : "Regra NAT desativada");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao alterar regra NAT: ${getErrorMessage(error)}`),
  });

  function openCreateDialog() {
    setEditingUid(null);
    setForm(emptyForm);
    setDialogOpen(true);
  }

  function openEditDialog(row: JsonRecord) {
    setEditingUid(String(row.uid));
    const methodName = refName(row.method).toLowerCase();
    setForm({
      originalSource: refNames(row["original-source"]),
      originalDestination: refNames(row["original-destination"]),
      originalService: refNames(row["original-service"]),
      translatedSource: refNames(row["translated-source"]),
      translatedDestination: refNames(row["translated-destination"]),
      translatedService: refNames(row["translated-service"]),
      method: (METHODS as readonly string[]).includes(methodName) ? (methodName as MethodOption) : "hide",
      position: "bottom",
      enabled: row.enabled !== false,
    });
    setDialogOpen(true);
  }

  function handleDelete(uid: string) {
    if (!window.confirm("Remover esta regra NAT?")) return;
    deleteMutation.mutate(uid);
  }

  function handleSubmit() {
    // Always send explicit values instead of omitting empty fields. For
    // NAT, `original-*` defaults to "Any" (match anything) and
    // `translated-*` defaults to "Original" (don't translate this field) —
    // both are Check Point's own builtin object names.
    const anyOrList = (arr: string[]) => (arr.length > 0 ? arr : ["Any"]);
    const originalOrList = (arr: string[]) => (arr.length > 0 ? arr : ["Original"]);

    const payload: JsonRecord = {
      method: form.method,
      "original-source": anyOrList(form.originalSource),
      "original-destination": anyOrList(form.originalDestination),
      "original-service": anyOrList(form.originalService),
      "translated-source": originalOrList(form.translatedSource),
      "translated-destination": originalOrList(form.translatedDestination),
      "translated-service": originalOrList(form.translatedService),
    };

    if (editingUid) {
      editMutation.mutate({ uid: editingUid, fields: { ...payload, enabled: form.enabled } });
      return;
    }
    addMutation.mutate({ ...payload, position: form.position });
  }

  const isSaving = addMutation.isPending || editMutation.isPending;

  return (
    <div>
      <PageHeader
        title="NAT"
        subtitle="Regras de NAT de um pacote de política."
        actions={
          <Button onClick={openCreateDialog} disabled={!pkgName}>
            <Plus className="size-4" />
            Nova regra NAT
          </Button>
        }
      />

      <div className="mb-4 max-w-xs">
        <Label htmlFor="package-select" className="mb-1.5 block">
          Package
        </Label>
        <Select value={pkgName} onValueChange={setPkgName}>
          <SelectTrigger id="package-select">
            <SelectValue placeholder="Selecione um package" />
          </SelectTrigger>
          <SelectContent>
            {packages.map((pkg) => {
              const label = pkgLabel(pkg);
              return (
                <SelectItem key={label || String(pkg.uid)} value={label}>
                  {label}
                </SelectItem>
              );
            })}
          </SelectContent>
        </Select>
      </div>

      <RulebaseTable
        rows={rows}
        emptyMessage={pkgName ? "Este pacote ainda não tem regras de NAT." : "Selecione um pacote para ver as regras."}
        columns={[
          { header: "#", cell: (row) => (typeof row["rule-number"] === "number" ? String(row["rule-number"]) : "") },
          {
            header: "Ativa",
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
                  aria-label={enabled ? "Desativar regra NAT" : "Ativar regra NAT"}
                />
              );
            },
          },
          { header: "Orig. origem", cell: (row) => refName(row["original-source"]) },
          { header: "Orig. destino", cell: (row) => refName(row["original-destination"]) },
          { header: "Orig. serviço", cell: (row) => refName(row["original-service"]) },
          { header: "Método", cell: (row) => refName(row.method) },
        ]}
        renderActions={(row) => (
          <div className="flex justify-end gap-1">
            <Button variant="ghost" size="icon" onClick={() => openEditDialog(row)} aria-label="Editar regra NAT">
              <Pencil className="size-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => handleDelete(String(row.uid ?? ""))}
              aria-label="Remover regra NAT"
            >
              <Trash2 className="size-4" />
            </Button>
          </div>
        )}
      />

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen} modal={false}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingUid ? "Editar regra NAT" : "Nova regra NAT"}</DialogTitle>
            <DialogDescription>
              {editingUid
                ? "Campos de objetos vazios não são alterados."
                : "Campos vazios ficam de fora — só envie o que quer definir."}
            </DialogDescription>
          </DialogHeader>

          <div className="flex max-h-[60vh] flex-col gap-4 overflow-y-auto pr-1">
            <div className="flex flex-col gap-1.5">
              <Label htmlFor="nat-method">Método</Label>
              <Select
                value={form.method}
                onValueChange={(value) => setForm((prev) => ({ ...prev, method: value as MethodOption }))}
              >
                <SelectTrigger id="nat-method">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {METHODS.map((method) => (
                    <SelectItem key={method} value={method}>
                      {method}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {editingUid ? (
              <div className="flex items-center gap-2">
                <input
                  id="nat-enabled"
                  type="checkbox"
                  checked={form.enabled}
                  onChange={(e) => setForm((prev) => ({ ...prev, enabled: e.target.checked }))}
                  className="size-4 rounded border-border accent-accent"
                />
                <Label htmlFor="nat-enabled">Habilitada</Label>
              </div>
            ) : (
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="nat-position">Posição</Label>
                <Select
                  value={form.position}
                  onValueChange={(value) => setForm((prev) => ({ ...prev, position: value as "top" | "bottom" }))}
                >
                  <SelectTrigger id="nat-position">
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
              <Label htmlFor="nat-original-source">Origem original</Label>
              <ObjectPicker
                value={form.originalSource}
                onChange={(names) => setForm((prev) => ({ ...prev, originalSource: names }))}
                placeholder="Buscar objetos... (vazio = Any)"
                categories={["network"]}
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="nat-original-destination">Destino original</Label>
              <ObjectPicker
                value={form.originalDestination}
                onChange={(names) => setForm((prev) => ({ ...prev, originalDestination: names }))}
                placeholder="Buscar objetos... (vazio = Any)"
                categories={["network"]}
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="nat-original-service">Serviço original</Label>
              <ObjectPicker
                value={form.originalService}
                onChange={(names) => setForm((prev) => ({ ...prev, originalService: names }))}
                placeholder="Buscar objetos... (vazio = Any)"
                categories={["service"]}
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="nat-translated-source">Origem traduzida</Label>
              <ObjectPicker
                value={form.translatedSource}
                onChange={(names) => setForm((prev) => ({ ...prev, translatedSource: names }))}
                placeholder="Buscar objetos... (opcional)"
                categories={["network"]}
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="nat-translated-destination">Destino traduzido</Label>
              <ObjectPicker
                value={form.translatedDestination}
                onChange={(names) => setForm((prev) => ({ ...prev, translatedDestination: names }))}
                placeholder="Buscar objetos... (opcional)"
                categories={["network"]}
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="nat-translated-service">Serviço traduzido</Label>
              <ObjectPicker
                value={form.translatedService}
                onChange={(names) => setForm((prev) => ({ ...prev, translatedService: names }))}
                placeholder="Buscar objetos... (opcional)"
                categories={["service"]}
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
