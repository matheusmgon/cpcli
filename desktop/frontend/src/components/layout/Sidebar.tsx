import { ShieldCheck } from "lucide-react";
import { NavLink } from "react-router-dom";

import { navGroups } from "@/config/nav";
import { cn } from "@/lib/utils";

/** Fixed deep-navy brand rail — intentionally constant across light/dark
 * themes (see index.css), grouped nav sourced from config/nav.ts. */
export function Sidebar() {
  return (
    <aside className="flex h-full w-64 shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-text">
      <div className="flex h-14 items-center gap-2 border-b border-sidebar-border px-5">
        <ShieldCheck className="size-5 text-accent" />
        <span className="text-sm font-semibold tracking-wide">CheckPoint Console</span>
      </div>

      <nav className="flex-1 overflow-y-auto px-3 py-4">
        {navGroups.map((group, idx) => (
          <div key={group.title ?? `top-${idx}`} className={cn(idx > 0 && "mt-5")}>
            {group.title && (
              <p className="mb-1.5 px-2 text-[11px] font-semibold uppercase tracking-wider text-sidebar-muted">
                {group.title}
              </p>
            )}
            <ul className="flex flex-col gap-0.5">
              {group.items.map((item) => (
                <li key={item.to}>
                  <NavLink
                    to={item.to}
                    className={({ isActive }) =>
                      cn(
                        "flex items-center gap-2.5 rounded-md border-l-2 border-transparent px-2.5 py-1.5 text-sm " +
                          "font-medium text-sidebar-text/85 transition-colors hover:bg-sidebar-hover hover:text-sidebar-text",
                        isActive && "border-accent bg-sidebar-active text-sidebar-text",
                      )
                    }
                  >
                    <item.icon className="size-4 shrink-0" />
                    <span className="truncate">{item.label}</span>
                  </NavLink>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </nav>
    </aside>
  );
}
