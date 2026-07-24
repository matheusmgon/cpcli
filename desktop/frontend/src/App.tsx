import { Navigate, Route, Routes } from "react-router-dom";

import { AppShell } from "@/components/layout/AppShell";
import { AccessRulesPage } from "@/pages/AccessRulesPage";
import { DashboardPage } from "@/pages/DashboardPage";
import { GatewaysPage } from "@/pages/GatewaysPage";
import { InstallPolicyPage } from "@/pages/InstallPolicyPage";
import { LoginPage } from "@/pages/LoginPage";
import { NatPage } from "@/pages/NatPage";
import { ObjectsPage } from "@/pages/ObjectsPage";
import { PackagesPage } from "@/pages/PackagesPage";
import { VpnPage } from "@/pages/VpnPage";

/** Route table — /login lives outside AppShell; every other route is a
 * child of the AppShell layout route (sidebar + topbar + session guard). */
export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />

      <Route element={<AppShell />}>
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/objects/:kind" element={<ObjectsPage />} />
        <Route path="/access-rules" element={<AccessRulesPage />} />
        <Route path="/nat" element={<NatPage />} />
        <Route path="/vpn/:kind" element={<VpnPage />} />
        <Route path="/install-policy" element={<InstallPolicyPage />} />
        <Route path="/gateways" element={<GatewaysPage />} />
        <Route path="/packages" element={<PackagesPage />} />
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Route>
    </Routes>
  );
}
