import { useEffect, type ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { Boxes, FolderKanban, Network, Server, ShieldCheck, type LucideIcon } from "lucide-react";
import { toast } from "sonner";

import { PageHeader } from "@/components/shared/PageHeader";
import { getService, type JsonRecord } from "@/lib/wailsService";

interface StatCardProps {
  icon: LucideIcon;
  label: string;
  value: number;
  isLoading: boolean;
  isError: boolean;
  children?: ReactNode;
}

/** One dashboard tile: icon + label + big count, with loading/error states. */
function StatCard({ icon: Icon, label, value, isLoading, isError, children }: StatCardProps) {
  return (
    <div className="rounded-lg border border-border bg-surface p-4 shadow-sm">
      <div className="flex items-center gap-2 text-muted">
        <Icon className="size-4" />
        <span className="text-xs font-medium uppercase tracking-wide">{label}</span>
      </div>
      <div className="mt-2 text-3xl font-semibold tracking-tight text-text">
        {isLoading ? (
          <div className="h-8 w-16 animate-pulse rounded bg-surface-2" />
        ) : isError ? (
          "—"
        ) : (
          value
        )}
      </div>
      {children}
    </div>
  );
}

function firstItemName(rows: JsonRecord[] | undefined): string {
  const first = rows?.[0];
  if (!first) return "";
  const name = first.name ?? first.uid;
  return typeof name === "string" ? name : "";
}

/** Post-login overview: counts of the environment's main object types,
 * fetched in parallel via independent react-query hooks. */
export function DashboardPage() {
  const hostsQuery = useQuery({
    queryKey: ["dashboard", "hosts"],
    queryFn: () => getService().listObjects("host", ""),
  });
  const networksQuery = useQuery({
    queryKey: ["dashboard", "networks"],
    queryFn: () => getService().listObjects("network", ""),
  });
  const layersQuery = useQuery({
    queryKey: ["dashboard", "layers"],
    queryFn: () => getService().listAccessLayers(),
  });
  const gatewaysQuery = useQuery({
    queryKey: ["dashboard", "gateways"],
    queryFn: () => getService().listGateways(),
  });
  const packagesQuery = useQuery({
    queryKey: ["dashboard", "packages"],
    queryFn: () => getService().listPackages(),
  });

  const layerName = firstItemName(layersQuery.data);

  const rulesQuery = useQuery({
    queryKey: ["dashboard", "rules", layerName],
    queryFn: () => getService().listAccessRulebase(layerName),
    enabled: Boolean(layerName),
  });

  useEffect(() => {
    const failures: Array<[boolean, string]> = [
      [hostsQuery.isError, "Falha ao carregar hosts."],
      [networksQuery.isError, "Falha ao carregar networks."],
      [layersQuery.isError, "Falha ao carregar access layers."],
      [rulesQuery.isError, "Falha ao carregar regras de acesso."],
      [gatewaysQuery.isError, "Falha ao carregar gateways."],
      [packagesQuery.isError, "Falha ao carregar pacotes de política."],
    ];
    for (const [isError, message] of failures) {
      if (isError) toast.error(message);
    }
  }, [
    hostsQuery.isError,
    networksQuery.isError,
    layersQuery.isError,
    rulesQuery.isError,
    gatewaysQuery.isError,
    packagesQuery.isError,
  ]);

  const packages = packagesQuery.data ?? [];
  const packageNames = packages
    .map((pkg) => (typeof pkg.name === "string" ? pkg.name : ""))
    .filter(Boolean);
  const visibleNames = packageNames.slice(0, 3);
  const extraCount = packageNames.length - visibleNames.length;

  // Layers loaded but there is no layer at all → show 0, not an error/spinner.
  const rulesLoading = layersQuery.isLoading || (Boolean(layerName) && rulesQuery.isLoading);
  const rulesError = layersQuery.isError || (Boolean(layerName) && rulesQuery.isError);
  const rulesValue = layerName ? (rulesQuery.data?.length ?? 0) : 0;

  return (
    <div>
      <PageHeader title="Dashboard" subtitle="Visão geral do ambiente" />

      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard
          icon={Server}
          label="Hosts"
          value={hostsQuery.data?.length ?? 0}
          isLoading={hostsQuery.isLoading}
          isError={hostsQuery.isError}
        />
        <StatCard
          icon={Network}
          label="Networks"
          value={networksQuery.data?.length ?? 0}
          isLoading={networksQuery.isLoading}
          isError={networksQuery.isError}
        />
        <StatCard
          icon={ShieldCheck}
          label="Regras de acesso"
          value={rulesValue}
          isLoading={rulesLoading}
          isError={rulesError}
        />
        <StatCard
          icon={Boxes}
          label="Gateways"
          value={gatewaysQuery.data?.length ?? 0}
          isLoading={gatewaysQuery.isLoading}
          isError={gatewaysQuery.isError}
        />
        <StatCard
          icon={FolderKanban}
          label="Pacotes de política"
          value={packageNames.length}
          isLoading={packagesQuery.isLoading}
          isError={packagesQuery.isError}
        >
          {!packagesQuery.isLoading && !packagesQuery.isError && packageNames.length > 0 && (
            <p className="mt-2 truncate text-xs text-muted">
              {visibleNames.join(", ")}
              {extraCount > 0 && ` +${extraCount}`}
            </p>
          )}
        </StatCard>
      </div>
    </div>
  );
}
