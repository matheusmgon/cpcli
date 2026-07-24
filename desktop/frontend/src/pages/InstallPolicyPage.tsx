import { useMemo, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { toast } from "sonner";

import { EmptyState } from "@/components/shared/EmptyState";
import { PageHeader } from "@/components/shared/PageHeader";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { getService } from "@/lib/wailsService";

function errorMessage(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

/** Install-policy screen — pick a policy package, pick target gateways,
 * then verify and/or install. Read side effects are handled by the backend;
 * this page only orchestrates the two calls and surfaces their result. */
export function InstallPolicyPage() {
  const [pkg, setPkg] = useState("");
  const [targets, setTargets] = useState<string[]>([]);

  const { data: packages = [] } = useQuery({
    queryKey: ["packages"],
    queryFn: () => getService().listPackages(),
  });

  const { data: gateways = [] } = useQuery({
    queryKey: ["gateways"],
    queryFn: () => getService().listGateways(),
  });

  const gatewayTargets = useMemo(() => {
    const onlyGateways = gateways.filter((g) => /gateway/i.test(String(g.type ?? "")));
    return onlyGateways.length > 0 ? onlyGateways : gateways;
  }, [gateways]);

  const verifyMutation = useMutation({
    mutationFn: () => getService().verifyPolicy(pkg),
    onSuccess: () => toast.success("Política verificada com sucesso"),
    onError: (error: unknown) => toast.error(errorMessage(error, "Falha na verificação")),
  });

  const installMutation = useMutation({
    mutationFn: () => getService().installPolicy(pkg, targets),
    onSuccess: () => toast.success("Instalação de política concluída"),
    onError: (error: unknown) => toast.error(errorMessage(error, "Falha ao instalar política")),
  });

  function toggleTarget(name: string) {
    setTargets((prev) => (prev.includes(name) ? prev.filter((t) => t !== name) : [...prev, name]));
  }

  function handleVerify() {
    if (!pkg) {
      toast.warning("Selecione um pacote antes de verificar");
      return;
    }
    verifyMutation.mutate();
  }

  function handleInstall() {
    if (!pkg) {
      toast.warning("Selecione um pacote antes de instalar");
      return;
    }
    if (targets.length === 0) {
      toast.warning("Selecione ao menos um gateway de destino");
      return;
    }
    installMutation.mutate();
  }

  return (
    <div>
      <PageHeader title="Instalar política" subtitle="Escolha o pacote e os gateways de destino." />

      <div className="flex max-w-lg flex-col gap-6">
        <div className="flex flex-col gap-1.5">
          <label className="text-sm font-medium text-text">Pacote</label>
          <Select value={pkg} onValueChange={setPkg}>
            <SelectTrigger>
              <SelectValue placeholder="Selecione um pacote" />
            </SelectTrigger>
            <SelectContent>
              {packages.map((p) => (
                <SelectItem key={String(p.name)} value={String(p.name)}>
                  {String(p.name)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="flex flex-col gap-1.5">
          <label className="text-sm font-medium text-text">Gateways de destino</label>
          {gatewayTargets.length === 0 ? (
            <EmptyState title="Nenhum gateway encontrado" />
          ) : (
            <div className="flex flex-wrap gap-2">
              {gatewayTargets.map((g) => {
                const name = String(g.name);
                return (
                  <label
                    key={name}
                    className="flex items-center gap-2 rounded-md border border-border px-2 py-1 text-sm text-text"
                  >
                    <input
                      type="checkbox"
                      className="size-3.5 rounded border-border accent-accent"
                      checked={targets.includes(name)}
                      onChange={() => toggleTarget(name)}
                    />
                    {name}
                  </label>
                );
              })}
            </div>
          )}
        </div>

        <div className="flex gap-2">
          <Button variant="outline" disabled={verifyMutation.isPending} onClick={handleVerify}>
            {verifyMutation.isPending ? "Verificando…" : "Verificar"}
          </Button>
          <Button variant="default" disabled={installMutation.isPending} onClick={handleInstall}>
            {installMutation.isPending ? "Instalando…" : "Instalar política"}
          </Button>
        </div>
      </div>
    </div>
  );
}
