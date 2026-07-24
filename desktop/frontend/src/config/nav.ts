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
  Rocket,
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
  /** null for the top-level, ungrouped item (Dashboard). */
  title: string | null;
  items: NavItem[];
}

/** Single source of truth for sidebar navigation — consumed by
 * `components/layout/Sidebar.tsx` so the route list isn't duplicated. */
export const navGroups: NavGroup[] = [
  {
    title: null,
    items: [{ label: "Dashboard", to: "/dashboard", icon: LayoutDashboard }],
  },
  {
    title: "Objetos",
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
    title: "Política",
    items: [
      { label: "Access Rules", to: "/access-rules", icon: ShieldCheck },
      { label: "NAT", to: "/nat", icon: Shuffle },
      { label: "Instalar política", to: "/install-policy", icon: Rocket },
    ],
  },
  {
    title: "VPN",
    items: [
      { label: "Star", to: "/vpn/star", icon: Star },
      { label: "Meshed", to: "/vpn/meshed", icon: Share2 },
    ],
  },
  {
    title: "Infraestrutura",
    items: [
      { label: "Gateways", to: "/gateways", icon: Boxes },
      { label: "Packages", to: "/packages", icon: FolderKanban },
    ],
  },
];
