import type { LucideIcon } from "lucide-react";
import {
  LayoutDashboard,
  Server,
  Network,
  Package,
  Plug,
  Radio,
  Ruler,
  Puzzle,
  UserCog,
  ShieldCheck,
  Shuffle,
  ShieldAlert,
  Lock,
  Rocket,
  ScrollText,
  Star,
  Share2,
  Boxes,
  FolderKanban,
} from "lucide-react";

export interface NavItem {
  label: string;
  to: string;
  icon: LucideIcon;
}

export interface NavGroup {
  /** null for top-level, ungrouped items (Dashboard, Gateways & Servers). */
  title: string | null;
  items: NavItem[];
}

/** Single source of truth for sidebar navigation — consumed by
 * `components/layout/Sidebar.tsx` so the route list isn't duplicated.
 *
 * Structure mirrors the real SmartConsole: Access Control + NAT + Threat
 * Prevention + HTTPS Inspection all live under one "Security Policies"
 * concept there (with Install Policy alongside them), not as unrelated
 * top-level items — that's the grouping that was missing before and made
 * the sidebar feel disorganized compared to the real product. */
export const navGroups: NavGroup[] = [
  {
    title: null,
    items: [{ label: "Dashboard", to: "/dashboard", icon: LayoutDashboard }],
  },
  {
    title: null,
    items: [{ label: "Gateways & Servers", to: "/gateways", icon: Boxes }],
  },
  {
    title: "Security Policies",
    items: [
      { label: "Access Control", to: "/access-rules", icon: ShieldCheck },
      { label: "NAT", to: "/nat", icon: Shuffle },
      { label: "Threat Prevention", to: "/threat-prevention", icon: ShieldAlert },
      { label: "HTTPS Inspection", to: "/https-inspection", icon: Lock },
      { label: "Install Policy", to: "/install-policy", icon: Rocket },
    ],
  },
  {
    title: "Monitoring",
    items: [{ label: "Logs & Monitor", to: "/logs", icon: ScrollText }],
  },
  {
    title: "VPN Communities",
    items: [
      { label: "Star", to: "/vpn/star", icon: Star },
      { label: "Meshed", to: "/vpn/meshed", icon: Share2 },
    ],
  },
  {
    title: "Object Explorer",
    items: [
      { label: "Hosts", to: "/objects/host", icon: Server },
      { label: "Networks", to: "/objects/network", icon: Network },
      { label: "Groups", to: "/objects/group", icon: Package },
      { label: "Services TCP", to: "/objects/service-tcp", icon: Plug },
      { label: "Services UDP", to: "/objects/service-udp", icon: Radio },
      { label: "Address Ranges", to: "/objects/address-range", icon: Ruler },
      { label: "Service Groups", to: "/objects/service-group", icon: Puzzle },
      { label: "Access Roles", to: "/objects/access-role", icon: UserCog },
    ],
  },
  {
    title: "Manage",
    items: [{ label: "Policy Packages", to: "/packages", icon: FolderKanban }],
  },
];
