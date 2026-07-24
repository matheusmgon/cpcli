import * as Real from "../../wailsjs/go/service/Service";
import { mockService } from "./mockService";

/** Loosely-typed API payload — the Go facade accepts/returns arbitrary
 * Check Point Management API JSON, so a fixed schema isn't available. */
export type JsonRecord = Record<string, unknown>;

export interface SessionInfo {
  connected: boolean;
  server: string;
  user: string;
  apiVersion: string;
}

export interface LoginRequest {
  server: string;
  port: number;
  user: string;
  password: string;
  apiKey: string;
  domain: string;
  readOnly: boolean;
  insecure: boolean;
}

/** Mirrors service.Service (service/service.go) — the same surface bound
 * into the Wails window, generated as `wailsjs/go/service/Service`. */
export interface CpService {
  login(req: LoginRequest): Promise<SessionInfo>;
  status(): Promise<SessionInfo>;
  logout(): Promise<void>;
  objectKinds(): Promise<string[]>;
  listObjects(kind: string, filter: string): Promise<JsonRecord[]>;
  getObject(kind: string, name: string): Promise<JsonRecord>;
  addObject(kind: string, fields: JsonRecord): Promise<JsonRecord>;
  setObject(kind: string, fields: JsonRecord): Promise<JsonRecord>;
  deleteObject(kind: string, name: string): Promise<void>;
  listAccessLayers(): Promise<JsonRecord[]>;
  listAccessRulebase(layer: string): Promise<JsonRecord[]>;
  listNatRulebase(pkg: string): Promise<JsonRecord[]>;
  listPackages(): Promise<JsonRecord[]>;
  listGateways(): Promise<JsonRecord[]>;
  installPolicy(pkg: string, targets: string[]): Promise<JsonRecord>;
  verifyPolicy(pkg: string): Promise<JsonRecord>;
  addAccessRule(layer: string, fields: JsonRecord): Promise<JsonRecord>;
  setAccessRule(layer: string, uid: string, fields: JsonRecord): Promise<JsonRecord>;
  deleteAccessRule(layer: string, uid: string): Promise<void>;
  addNatRule(pkg: string, fields: JsonRecord): Promise<JsonRecord>;
  setNatRule(pkg: string, uid: string, fields: JsonRecord): Promise<JsonRecord>;
  deleteNatRule(pkg: string, uid: string): Promise<void>;
  // Threat Prevention
  listThreatLayers(): Promise<JsonRecord[]>;
  listThreatRulebase(layer: string): Promise<JsonRecord[]>;
  addThreatRule(layer: string, fields: JsonRecord): Promise<JsonRecord>;
  setThreatRule(layer: string, uid: string, fields: JsonRecord): Promise<JsonRecord>;
  deleteThreatRule(layer: string, uid: string): Promise<void>;
  listThreatProfiles(): Promise<JsonRecord[]>;
  addThreatProfile(fields: JsonRecord): Promise<JsonRecord>;
  setThreatProfile(fields: JsonRecord): Promise<JsonRecord>;
  deleteThreatProfile(name: string): Promise<void>;
  // HTTPS Inspection
  listHttpsLayers(): Promise<JsonRecord[]>;
  listHttpsRulebase(layer: string): Promise<JsonRecord[]>;
  addHttpsRule(layer: string, fields: JsonRecord): Promise<JsonRecord>;
  setHttpsRule(layer: string, uid: string, fields: JsonRecord): Promise<JsonRecord>;
  deleteHttpsRule(layer: string, uid: string): Promise<void>;
  // Gateway interfaces (anti-spoofing)
  listGatewayInterfaces(gateway: string): Promise<JsonRecord[]>;
  setGatewayInterface(gateway: string, ifaceName: string, fields: JsonRecord): Promise<JsonRecord>;
  vpnKinds(): Promise<string[]>;
  listVpnCommunities(kind: string): Promise<JsonRecord[]>;
  getVpnCommunity(kind: string, name: string): Promise<JsonRecord>;
  addVpnCommunity(kind: string, fields: JsonRecord): Promise<JsonRecord>;
  setVpnCommunity(kind: string, fields: JsonRecord): Promise<JsonRecord>;
  deleteVpnCommunity(kind: string, name: string): Promise<void>;
  publish(): Promise<JsonRecord>;
  discard(): Promise<JsonRecord>;
}

function hasWailsRuntime(): boolean {
  const w = window as unknown as { go?: { service?: { Service?: unknown } } };
  return typeof w.go?.service?.Service !== "undefined";
}

const realService: CpService = {
  login: (req) => Real.Login(req as never),
  status: () => Real.Status() as Promise<SessionInfo>,
  logout: () => Real.Logout(),
  objectKinds: () => Real.ObjectKinds(),
  listObjects: (kind, filter) => Real.ListObjects(kind, filter),
  getObject: (kind, name) => Real.GetObject(kind, name),
  addObject: (kind, fields) => Real.AddObject(kind, fields),
  setObject: (kind, fields) => Real.SetObject(kind, fields),
  deleteObject: (kind, name) => Real.DeleteObject(kind, name),
  listAccessLayers: () => Real.ListAccessLayers(),
  listAccessRulebase: (layer) => Real.ListAccessRulebase(layer),
  listNatRulebase: (pkg) => Real.ListNatRulebase(pkg),
  listPackages: () => Real.ListPackages(),
  listGateways: () => Real.ListGateways(),
  installPolicy: (pkg, targets) => Real.InstallPolicy(pkg, targets),
  verifyPolicy: (pkg) => Real.VerifyPolicy(pkg),
  addAccessRule: (layer, fields) => Real.AddAccessRule(layer, fields),
  setAccessRule: (layer, uid, fields) => Real.SetAccessRule(layer, uid, fields),
  deleteAccessRule: (layer, uid) => Real.DeleteAccessRule(layer, uid),
  addNatRule: (pkg, fields) => Real.AddNatRule(pkg, fields),
  setNatRule: (pkg, uid, fields) => Real.SetNatRule(pkg, uid, fields),
  deleteNatRule: (pkg, uid) => Real.DeleteNatRule(pkg, uid),
  listThreatLayers: () => Real.ListThreatLayers(),
  listThreatRulebase: (layer) => Real.ListThreatRulebase(layer),
  addThreatRule: (layer, fields) => Real.AddThreatRule(layer, fields),
  setThreatRule: (layer, uid, fields) => Real.SetThreatRule(layer, uid, fields),
  deleteThreatRule: (layer, uid) => Real.DeleteThreatRule(layer, uid),
  listThreatProfiles: () => Real.ListThreatProfiles(),
  addThreatProfile: (fields) => Real.AddThreatProfile(fields),
  setThreatProfile: (fields) => Real.SetThreatProfile(fields),
  deleteThreatProfile: (name) => Real.DeleteThreatProfile(name),
  listHttpsLayers: () => Real.ListHttpsLayers(),
  listHttpsRulebase: (layer) => Real.ListHttpsRulebase(layer),
  addHttpsRule: (layer, fields) => Real.AddHttpsRule(layer, fields),
  setHttpsRule: (layer, uid, fields) => Real.SetHttpsRule(layer, uid, fields),
  deleteHttpsRule: (layer, uid) => Real.DeleteHttpsRule(layer, uid),
  listGatewayInterfaces: (gateway) => Real.ListGatewayInterfaces(gateway),
  setGatewayInterface: (gateway, ifaceName, fields) => Real.SetGatewayInterface(gateway, ifaceName, fields),
  vpnKinds: () => Real.VpnKinds(),
  listVpnCommunities: (kind) => Real.ListVpnCommunities(kind),
  getVpnCommunity: (kind, name) => Real.GetVpnCommunity(kind, name),
  addVpnCommunity: (kind, fields) => Real.AddVpnCommunity(kind, fields),
  setVpnCommunity: (kind, fields) => Real.SetVpnCommunity(kind, fields),
  deleteVpnCommunity: (kind, name) => Real.DeleteVpnCommunity(kind, name),
  publish: () => Real.Publish(),
  discard: () => Real.Discard(),
};

/** Real Wails bindings inside the native app / `wails dev`; an in-memory
 * mock everywhere else (plain `vite dev`/`vite preview`) so the UI is
 * iterable and screenshot-able in an ordinary browser. */
export function getService(): CpService {
  return hasWailsRuntime() ? realService : mockService;
}
