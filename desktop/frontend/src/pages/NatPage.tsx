import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2 } from "lucide-react";
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

function getErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

function pkgLabel(pkg: JsonRecord): string {
  return typeof pkg.name === "string" ? pkg.name : refName(pkg);
}

const METHODS = ["static", "hide"] as const;
type MethodOption = (typeof METHODS)[number];

interface NatFormState {
  originalSource: string;
  originalDestination: string;
  originalService: string;
  translatedSource: string;
  translatedDestination: string;
  translatedService: string;
  method: MethodOption;
  position: "top" | "bottom";
}

const emptyForm: NatFormState = {
  originalSource: "",
  originalDestination: "",
  originalService: "",
  translatedSource: "",
  translatedDestination: "",
  translatedService: "",
  method: "hide",
  position: "bottom",
};

/** NAT rulebase for a single policy package — package picker, rule table
 * (with section dividers), and an add dialog. No edit flow: the service
 * exposes `setNatRule`, but the UX here only requires add + delete. */
export function NatPage() {
  const queryClient = useQueryClient();
  const [pkgName, setPkgName] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
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

  const deleteMutation = useMutation({
    mutationFn: (uid: string) => getService().deleteNatRule(pkgName, uid),
    onSuccess: () => {
      toast.success("Regra NAT removida");
      invalidateRulebase();
      useSessionStore.getState().markPending(1);
    },
    onError: (error) => toast.error(`Falha ao remover regra NAT: ${getErrorMessage(error)}`),
  });

  function openCreateDialog() {
    setForm(emptyForm);
    setDialogOpen(true);
  }

  function handleDelete(uid: string) {
    if (!window.confirm("Remover esta regra NAT?")) return;
    deleteMutation.mutate(uid);
  }

  function handleSubmit() {
    const fields: JsonRecord = {
      method: form.method,
      position: form.position,
      ...(form.originalSource && { "original-source": form.originalSource }),
      ...(form.originalDestination && { "original-destination": form.originalDestination }),
      ...(form.originalService && { "original-service": form.originalService }),
      ...(form.translatedSource && { "translated-source": form.translatedSource }),
      ...(form.translatedDestination && { "translated-destination": form.translatedDestination }),
      ...(form.translatedService && { "translated-service": form.translatedService }),
    };
    addMutation.mutate(fields);
  }

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
          { header: "Orig. origem", cell: (row) => refName(row["original-source"]) },
          { header: "Orig. destino", cell: (row) => refName(row["original-destination"]) },
          { header: "Orig. serviço", cell: (row) => refName(row["original-service"]) },
          { header: "Método", cell: (row) => refName(row.method) },
        ]}
        renderActions={(row) => (
          <Button
            variant="ghost"
            size="icon"
            onClick={() => handleDelete(String(row.uid ?? ""))}
            aria-label="Remover regra NAT"
          >
            <Trash2 className="size-4" />
          </Button>
        )}
      />

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Nova regra NAT</DialogTitle>
            <DialogDescription>Campos vazios ficam de fora — só envie o que quer definir.</DialogDescription>
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

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="nat-original-source">Origem original</Label>
              <Input
                id="nat-original-source"
                value={form.originalSource}
                onChange={(e) => setForm((prev) => ({ ...prev, originalSource: e.target.value }))}
                placeholder="nome do objeto (vazio = Any)"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="nat-original-destination">Destino original</Label>
              <Input
                id="nat-original-destination"
                value={form.originalDestination}
                onChange={(e) => setForm((prev) => ({ ...prev, originalDestination: e.target.value }))}
                placeholder="nome do objeto (vazio = Any)"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="nat-original-service">Serviço original</Label>
              <Input
                id="nat-original-service"
                value={form.originalService}
                onChange={(e) => setForm((prev) => ({ ...prev, originalService: e.target.value }))}
                placeholder="nome do objeto (vazio = Any)"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="nat-translated-source">Origem traduzida</Label>
              <Input
                id="nat-translated-source"
                value={form.translatedSource}
                onChange={(e) => setForm((prev) => ({ ...prev, translatedSource: e.target.value }))}
                placeholder="nome do objeto (opcional)"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="nat-translated-destination">Destino traduzido</Label>
              <Input
                id="nat-translated-destination"
                value={form.translatedDestination}
                onChange={(e) => setForm((prev) => ({ ...prev, translatedDestination: e.target.value }))}
                placeholder="nome do objeto (opcional)"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor="nat-translated-service">Serviço traduzido</Label>
              <Input
                id="nat-translated-service"
                value={form.translatedService}
                onChange={(e) => setForm((prev) => ({ ...prev, translatedService: e.target.value }))}
                placeholder="nome do objeto (opcional)"
              />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>
              Cancelar
            </Button>
            <Button type="button" onClick={handleSubmit} disabled={addMutation.isPending}>
              Salvar
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
