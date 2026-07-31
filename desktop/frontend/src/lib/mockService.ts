import type { CpService, JsonRecord, LoginRequest, SessionInfo } from "./wailsService";

/** Canned sample data so the UI is fully explorable in a plain browser
 * (`vite dev` / `vite preview`, outside the native Wails shell) — used only
 * when `window.go` isn't present. Not a full fake backend: enough shape per
 * screen to render every page realistically. */

let session: SessionInfo = { connected: false, server: "", user: "", apiVersion: "" };
let pending = 0;

const delay = <T,>(v: T) => new Promise<T>((resolve) => setTimeout(() => resolve(v), 220));

const hosts: JsonRecord[] = [
  { name: "web-01", "ipv4-address": "10.10.1.11", comments: "Web server" },
  { name: "db-01", "ipv4-address": "10.10.1.20", comments: "Database" },
  { name: "mgmt-server", "ipv4-address": "192.168.56.10", comments: "" },
];
const networks: JsonRecord[] = [
  { name: "lan-servers", subnet4: "10.10.1.0", "mask-length4": 24 },
  { name: "dmz", subnet4: "10.10.2.0", "mask-length4": 24 },
];
const objectStore: Record<string, JsonRecord[]> = {
  host: hosts,
  network: networks,
  group: [{ name: "grp-servers", members: ["web-01", "db-01"] }],
  "service-tcp": [{ name: "svc-https", port: "443" }],
  "service-udp": [{ name: "svc-dns", port: "53" }],
  "address-range": [{ name: "dhcp-pool", "ipv4-address-first": "10.10.5.10", "ipv4-address-last": "10.10.5.100" }],
  "service-group": [{ name: "grp-web-svc", members: ["svc-https"] }],
  "access-role": [{ name: "role-admins", comments: "Administrators" }],
};

const layer = { name: "Network", uid: "layer-mock-1", type: "access-layer" };
const rulebase: JsonRecord[] = [
  {
    "rule-number": 1,
    uid: "rule-mock-1",
    name: "Allow LAN → Internet",
    action: { name: "Accept" },
    source: [{ name: "lan-servers" }],
    destination: [{ name: "Any" }],
    service: [{ name: "Any" }],
  },
  {
    "rule-number": 2,
    uid: "rule-mock-2",
    name: "Block DMZ → LAN",
    action: { name: "Drop" },
    source: [{ name: "dmz" }],
    destination: [{ name: "lan-servers" }],
    service: [{ name: "Any" }],
  },
];

const pkg = { name: "Standard", uid: "pkg-mock-1", access: true, "threat-prevention": true };
const natRulebase: JsonRecord[] = [
  {
    type: "nat-section",
    name: "Automatic Generated Rules : Network Hide NAT",
    uid: "nat-section-mock-1",
    rulebase: [
      {
        "rule-number": 1,
        uid: "nat-auto-mock-1",
        "original-source": "lan-servers",
        "original-destination": "Any",
        "original-service": "Any",
        method: "hide",
      },
    ],
  },
  {
    "rule-number": 1,
    uid: "nat-mock-1",
    "original-source": { name: "lan-servers" },
    "original-destination": { name: "Any" },
    "original-service": { name: "Any" },
    method: "hide",
  },
];

const gateways: JsonRecord[] = [{ name: "gw-fw01", type: "simple-gateway", "ipv4-address": "192.168.56.10" }];

/** Shared filter behind `searchObjects`/`countObjects` — mirrors the real
 * `show-objects` semantics closely enough for mock/offline UI work: `type`
 * narrows to one object kind, `filter` substring-matches the name. */
function matchObjects(filter: string, objType: string): JsonRecord[] {
  const all: JsonRecord[] = [
    ...Object.entries(objectStore).flatMap(([kind, rows]) => rows.map((r) => ({ ...r, type: kind }))),
    ...gateways,
  ];
  const f = filter.trim().toLowerCase();
  return all.filter((r) => {
    if (objType && r.type !== objType) return false;
    if (!f) return true;
    return String(r.name ?? "").toLowerCase().includes(f);
  });
}
const vpnMeshed: JsonRecord[] = [{ name: "MyIntranet", gateways: [{ name: "gw-fw01" }] }];
const vpnStar: JsonRecord[] = [];

const threatLayer = { name: "Standard Threat Prevention", uid: "threat-layer-mock-1", type: "threat-layer" };
const threatRulebase: JsonRecord[] = [
  {
    "rule-number": 1,
    uid: "threat-rule-mock-1",
    name: "Default profile",
    action: { name: "Optimized" },
    "protected-scope": [{ name: "Any" }],
  },
];
const threatProfiles: JsonRecord[] = [{ name: "Optimized", uid: "threat-profile-mock-1" }];

const httpsLayer = { name: "Default Layer", uid: "https-layer-mock-1", type: "https-layer" };
const httpsRulebase: JsonRecord[] = [
  {
    "rule-number": 1,
    uid: "https-rule-mock-1",
    name: "Inspect all",
    action: "Inspect",
    source: [{ name: "Any" }],
    destination: [{ name: "Any" }],
  },
];

const gatewayInterfaces: Record<string, JsonRecord[]> = {
  "gw-fw01": [
    { name: "eth0", "ipv4-address": "192.168.0.200", "ipv4-mask-length": 24, "anti-spoofing": false, topology: "automatic" },
    {
      name: "eth1",
      "ipv4-address": "10.10.1.1",
      "ipv4-mask-length": 24,
      "anti-spoofing": true,
      topology: "internal",
      "anti-spoofing-settings": { action: "prevent", "spoof-tracking": "log", "exclude-packets": false },
    },
  ],
};

const gatewayBlades: Record<string, JsonRecord> = {
  "gw-fw01": {
    firewall: true,
    vpn: false,
    "application-control": false,
    "url-filtering": false,
    ips: true,
    "anti-bot": false,
    "anti-virus": false,
    "threat-emulation": false,
    "threat-extraction": false,
    "content-awareness": false,
    "identity-awareness": false,
    "https-inspection": false,
  },
};

function requireSession() {
  if (!session.connected) throw new Error("not connected — please log in first");
}

export const mockService: CpService = {
  async login(req: LoginRequest) {
    session = { connected: true, server: req.server, user: req.user, apiVersion: "2.x (mock)" };
    return delay(session);
  },
  async status() {
    return delay(session);
  },
  async logout() {
    session = { connected: false, server: "", user: "", apiVersion: "" };
  },
  async objectKinds() {
    return delay(Object.keys(objectStore));
  },
  async listObjects(kind, filter) {
    requireSession();
    const rows = objectStore[kind] ?? [];
    const f = filter.trim().toLowerCase();
    return delay(f ? rows.filter((r) => String(r.name).toLowerCase().includes(f)) : rows);
  },
  async getObject(kind, name) {
    requireSession();
    return delay((objectStore[kind] ?? []).find((r) => r.name === name) ?? {});
  },
  async addObject(kind, fields) {
    requireSession();
    (objectStore[kind] ??= []).push(fields);
    pending++;
    return delay(fields);
  },
  async setObject(kind, fields) {
    requireSession();
    const rows = objectStore[kind] ?? [];
    const idx = rows.findIndex((r) => r.name === fields.name);
    if (idx >= 0) rows[idx] = { ...rows[idx], ...fields };
    pending++;
    return delay(fields);
  },
  async deleteObject(kind, name) {
    requireSession();
    objectStore[kind] = (objectStore[kind] ?? []).filter((r) => r.name !== name);
    pending++;
  },
  async listAccessLayers() {
    requireSession();
    return delay([layer]);
  },
  async listAccessRulebase() {
    requireSession();
    return delay(rulebase);
  },
  async listNatRulebase() {
    requireSession();
    return delay(natRulebase);
  },
  async listPackages() {
    requireSession();
    return delay([pkg]);
  },
  async listGateways() {
    requireSession();
    return delay(gateways);
  },
  async installPolicy() {
    requireSession();
    pending = 0;
    return delay({ status: "succeeded" });
  },
  async verifyPolicy() {
    requireSession();
    return delay({ status: "succeeded" });
  },
  async addAccessRule(_layer, fields) {
    requireSession();
    rulebase.unshift({ ...fields, uid: `rule-mock-${rulebase.length + 1}`, "rule-number": rulebase.length + 1 });
    pending++;
    return delay(fields);
  },
  async setAccessRule(_layer, uid, fields) {
    requireSession();
    const idx = rulebase.findIndex((r) => r.uid === uid);
    if (idx >= 0) rulebase[idx] = { ...rulebase[idx], ...fields };
    pending++;
    return delay(fields);
  },
  async deleteAccessRule(_layer, uid) {
    requireSession();
    const idx = rulebase.findIndex((r) => r.uid === uid);
    if (idx >= 0) rulebase.splice(idx, 1);
    pending++;
  },
  async addNatRule(_pkg, fields) {
    requireSession();
    natRulebase.unshift({ ...fields, uid: `nat-mock-${natRulebase.length + 1}`, "rule-number": natRulebase.length + 1 });
    pending++;
    return delay(fields);
  },
  async setNatRule(_pkg, uid, fields) {
    requireSession();
    const idx = natRulebase.findIndex((r) => r.uid === uid);
    if (idx >= 0) natRulebase[idx] = { ...natRulebase[idx], ...fields };
    pending++;
    return delay(fields);
  },
  async deleteNatRule(_pkg, uid) {
    requireSession();
    const idx = natRulebase.findIndex((r) => r.uid === uid);
    if (idx >= 0) natRulebase.splice(idx, 1);
    pending++;
  },
  async listThreatLayers() {
    requireSession();
    return delay([threatLayer]);
  },
  async listThreatRulebase() {
    requireSession();
    return delay(threatRulebase);
  },
  async addThreatRule(_layer, fields) {
    requireSession();
    threatRulebase.unshift({
      ...fields,
      uid: `threat-rule-mock-${threatRulebase.length + 1}`,
      "rule-number": threatRulebase.length + 1,
    });
    pending++;
    return delay(fields);
  },
  async setThreatRule(_layer, uid, fields) {
    requireSession();
    const idx = threatRulebase.findIndex((r) => r.uid === uid);
    if (idx >= 0) threatRulebase[idx] = { ...threatRulebase[idx], ...fields };
    pending++;
    return delay(fields);
  },
  async deleteThreatRule(_layer, uid) {
    requireSession();
    const idx = threatRulebase.findIndex((r) => r.uid === uid);
    if (idx >= 0) threatRulebase.splice(idx, 1);
    pending++;
  },
  async listThreatProfiles() {
    requireSession();
    return delay(threatProfiles);
  },
  async addThreatProfile(fields) {
    requireSession();
    threatProfiles.push({ ...fields, uid: `threat-profile-mock-${threatProfiles.length + 1}` });
    pending++;
    return delay(fields);
  },
  async setThreatProfile(fields) {
    requireSession();
    const idx = threatProfiles.findIndex((r) => r.name === fields.name);
    if (idx >= 0) threatProfiles[idx] = { ...threatProfiles[idx], ...fields };
    pending++;
    return delay(fields);
  },
  async deleteThreatProfile(name) {
    requireSession();
    const idx = threatProfiles.findIndex((r) => r.name === name);
    if (idx >= 0) threatProfiles.splice(idx, 1);
    pending++;
  },
  async listHttpsLayers() {
    requireSession();
    return delay([httpsLayer]);
  },
  async listHttpsRulebase() {
    requireSession();
    return delay(httpsRulebase);
  },
  async addHttpsRule(_layer, fields) {
    requireSession();
    httpsRulebase.unshift({
      ...fields,
      uid: `https-rule-mock-${httpsRulebase.length + 1}`,
      "rule-number": httpsRulebase.length + 1,
    });
    pending++;
    return delay(fields);
  },
  async setHttpsRule(_layer, uid, fields) {
    requireSession();
    const idx = httpsRulebase.findIndex((r) => r.uid === uid);
    if (idx >= 0) httpsRulebase[idx] = { ...httpsRulebase[idx], ...fields };
    pending++;
    return delay(fields);
  },
  async deleteHttpsRule(_layer, uid) {
    requireSession();
    const idx = httpsRulebase.findIndex((r) => r.uid === uid);
    if (idx >= 0) httpsRulebase.splice(idx, 1);
    pending++;
  },
  async listGatewayInterfaces(gateway) {
    requireSession();
    return delay(gatewayInterfaces[gateway] ?? []);
  },
  async setGatewayInterface(gateway, ifaceName, fields) {
    requireSession();
    const ifaces = gatewayInterfaces[gateway] ?? [];
    const idx = ifaces.findIndex((i) => i.name === ifaceName);
    if (idx >= 0) ifaces[idx] = { ...ifaces[idx], ...fields };
    pending++;
    return delay(fields);
  },
  async refreshGatewayTopology(gateway) {
    requireSession();
    pending++;
    return delay({ "target-name": gateway, task: "mock-get-interfaces" });
  },
  async getGatewayBlades(gateway) {
    requireSession();
    return delay(gatewayBlades[gateway] ?? {});
  },
  async setGatewayBlades(gateway, fields) {
    requireSession();
    gatewayBlades[gateway] = { ...(gatewayBlades[gateway] ?? {}), ...fields };
    pending++;
    return delay(gatewayBlades[gateway]);
  },
  async searchObjects(filter, objType) {
    requireSession();
    return delay(matchObjects(filter, objType).slice(0, 50));
  },
  async countObjects(filter, objType) {
    requireSession();
    return delay(matchObjects(filter, objType).length);
  },
  async vpnKinds() {
    return delay(["star", "meshed"]);
  },
  async listVpnCommunities(kind) {
    requireSession();
    return delay(kind === "star" ? vpnStar : vpnMeshed);
  },
  async getVpnCommunity(kind, name) {
    requireSession();
    const rows = kind === "star" ? vpnStar : vpnMeshed;
    return delay(rows.find((r) => r.name === name) ?? {});
  },
  async addVpnCommunity(kind, fields) {
    requireSession();
    (kind === "star" ? vpnStar : vpnMeshed).push(fields);
    pending++;
    return delay(fields);
  },
  async setVpnCommunity(kind, fields) {
    requireSession();
    const rows = kind === "star" ? vpnStar : vpnMeshed;
    const idx = rows.findIndex((r) => r.name === fields.name);
    if (idx >= 0) rows[idx] = { ...rows[idx], ...fields };
    pending++;
    return delay(fields);
  },
  async deleteVpnCommunity(kind, name) {
    requireSession();
    const rows = kind === "star" ? vpnStar : vpnMeshed;
    const idx = rows.findIndex((r) => r.name === name);
    if (idx >= 0) rows.splice(idx, 1);
    pending++;
  },
  async publish() {
    requireSession();
    const n = pending;
    pending = 0;
    return delay({ tasks: [{ status: "succeeded", "task-details": [{ publishResponse: { numberOfPublishedChanges: n } }] }] });
  },
  async discard() {
    requireSession();
    pending = 0;
    return delay({});
  },
  async readFirewallLogs(_gateway, filter, _limit) {
    requireSession();
    const sample: JsonRecord[] = [
      { time: "20:54:28", action: "accept", iface: "eth2", src: "192.168.56.1", dst: "192.168.56.10", proto: "tcp", service_id: "https", rule_name: "Admin Host-Only" },
      { time: "20:54:14", action: "drop", iface: "eth1", src: "10.0.10.100", dst: "8.8.8.8", proto: "udp", service_id: "domain-udp", rule_name: "Cleanup rule" },
      { time: "20:53:31", action: "accept", iface: "eth2", src: "192.168.56.1", dst: "192.168.56.10", proto: "tcp", service_id: "https", rule_name: "Admin Host-Only" },
      { time: "20:53:11", action: "drop", iface: "eth1", src: "10.0.10.100", dst: "8.8.8.8", proto: "udp", service_id: "domain-udp", rule_name: "Cleanup rule" },
    ];
    const f = filter.trim().toLowerCase();
    const out = f
      ? sample.filter((r) => Object.values(r).some((v) => String(v).toLowerCase().includes(f)))
      : sample;
    return delay(out);
  },
};
