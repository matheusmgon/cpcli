import type { JsonRecord } from "@/lib/wailsService";

export interface ObjectKindField {
  /** Key sent to the Go facade (`AddObject`/`SetObject` field map). */
  key: string;
  label: string;
  placeholder?: string;
}

export interface ObjectKindColumn {
  header: string;
  accessor: (row: JsonRecord) => string;
}

export interface ObjectKindConfig {
  /** Display name used in the page title, dialogs, and toasts. */
  title: string;
  fields: ObjectKindField[];
  columns: ObjectKindColumn[];
}

/** Renders any JSON value coming back from the Management API as plain text
 * — the API returns loosely-typed records, so table cells and edit-dialog
 * initial values need a defensive stringifier. */
function text(value: unknown): string {
  if (value === null || value === undefined) return "";
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean") return String(value);
  return String(value);
}

/** Config-driven definition of the 8 "simple object" kinds handled by the
 * generic `ObjectsPage` — one entry per `kind` string accepted by
 * `service.Service.ListObjects/AddObject/SetObject/DeleteObject`. */
export const objectKinds: Record<string, ObjectKindConfig> = {
  host: {
    title: "Hosts",
    fields: [
      { key: "name", label: "Nome", placeholder: "nome (ex: web-01)" },
      { key: "ip-address", label: "IP", placeholder: "ip (ex: 10.0.0.10)" },
    ],
    columns: [
      { header: "Nome", accessor: (row) => text(row.name) },
      {
        header: "IPv4",
        accessor: (row) => text(row["ipv4-address"] || row["ipv6-address"]),
      },
    ],
  },
  network: {
    title: "Networks",
    fields: [
      { key: "name", label: "Nome", placeholder: "nome (ex: lan)" },
      { key: "subnet4", label: "Subnet", placeholder: "subnet (ex: 10.0.0.0)" },
      { key: "mask-length4", label: "Máscara", placeholder: "máscara (ex: 24)" },
    ],
    columns: [
      { header: "Nome", accessor: (row) => text(row.name) },
      {
        header: "Subnet",
        accessor: (row) => `${text(row.subnet4)} / ${text(row["mask-length4"])}`,
      },
    ],
  },
  group: {
    title: "Groups",
    fields: [{ key: "name", label: "Nome", placeholder: "nome do grupo" }],
    columns: [
      { header: "Nome", accessor: (row) => text(row.name) },
      {
        header: "Membros",
        accessor: (row) => text(Array.isArray(row.members) ? row.members.length : 0),
      },
    ],
  },
  "service-tcp": {
    title: "Services TCP",
    fields: [
      { key: "name", label: "Nome", placeholder: "nome (ex: my-tcp)" },
      { key: "port", label: "Porta", placeholder: "porta (ex: 8080)" },
    ],
    columns: [
      { header: "Nome", accessor: (row) => text(row.name) },
      { header: "Porta", accessor: (row) => text(row.port) },
    ],
  },
  "service-udp": {
    title: "Services UDP",
    fields: [
      { key: "name", label: "Nome", placeholder: "nome (ex: my-udp)" },
      { key: "port", label: "Porta", placeholder: "porta (ex: 53)" },
    ],
    columns: [
      { header: "Nome", accessor: (row) => text(row.name) },
      { header: "Porta", accessor: (row) => text(row.port) },
    ],
  },
  "address-range": {
    title: "Address Ranges",
    fields: [
      { key: "name", label: "Nome", placeholder: "nome (ex: dhcp-pool)" },
      { key: "ipv4-address-first", label: "IP inicial", placeholder: "IP inicial" },
      { key: "ipv4-address-last", label: "IP final", placeholder: "IP final" },
    ],
    columns: [
      { header: "Nome", accessor: (row) => text(row.name) },
      {
        header: "Range",
        accessor: (row) =>
          `${text(row["ipv4-address-first"])} – ${text(row["ipv4-address-last"])}`,
      },
    ],
  },
  "service-group": {
    title: "Service Groups",
    fields: [{ key: "name", label: "Nome", placeholder: "nome do grupo de serviços" }],
    columns: [
      { header: "Nome", accessor: (row) => text(row.name) },
      {
        header: "Membros",
        accessor: (row) => text(Array.isArray(row.members) ? row.members.length : 0),
      },
    ],
  },
  "access-role": {
    title: "Access Roles",
    fields: [
      { key: "name", label: "Nome", placeholder: "nome do access-role" },
      { key: "comments", label: "Comentário", placeholder: "comentário (opcional)" },
    ],
    columns: [
      { header: "Nome", accessor: (row) => text(row.name) },
      { header: "Comentário", accessor: (row) => text(row.comments) },
    ],
  },
};
