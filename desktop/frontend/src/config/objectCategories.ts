import { objectKinds } from "@/config/objectKinds";

export interface ObjectCategory {
  key: string;
  label: string;
  /** Every `show-objects` `type` value that belongs to this category — the
   * picker fires one search/count call per type and merges the results,
   * since the Management API only accepts a single `type` per call. */
  types: string[];
}

/** Fixed type→category mapping, covering every object type the backend
 * already knows about (`internal/cli/object.go`), including types the UI
 * can't create yet (e.g. `security-zone`, `application-site`) — those are
 * still searchable/selectable, just not offered in the inline "criar novo"
 * form (see `creatableTypes` below). */
export const objectCategories: ObjectCategory[] = [
  {
    key: "network",
    label: "Rede",
    types: [
      "host",
      "network",
      "group",
      "address-range",
      "security-zone",
      "dns-domain",
      "wildcard",
      "dynamic-object",
      // Gateways/servers — needed so a rule's Source/Destination picker can
      // find the firewall itself (e.g. "CheckPointA" for the RC060 admin
      // rule). show-objects accepts all four types below individually.
      "simple-gateway",
      "checkpoint-host",
      "simple-cluster",
      "cluster-member",
    ],
  },
  {
    key: "service",
    label: "Serviços",
    types: ["service-tcp", "service-udp", "service-icmp", "service-other", "service-group"],
  },
  {
    key: "application",
    label: "Aplicações",
    types: ["application-site"],
  },
];

export type CategoryKey = (typeof objectCategories)[number]["key"];

/** Types within a category that the UI actually knows how to create
 * (present in `objectKinds`) — powers the picker's inline "+ Criar novo"
 * affordance and its type selector when more than one kind qualifies. */
export function creatableTypes(category: ObjectCategory): string[] {
  return category.types.filter((type) => type in objectKinds);
}

export function findCategory(key: string): ObjectCategory | undefined {
  return objectCategories.find((category) => category.key === key);
}
