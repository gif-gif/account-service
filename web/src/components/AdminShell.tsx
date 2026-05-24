import { useState } from "react";

import { AccountDetailPage } from "../pages/AccountDetailPage";
import { AccountsPage } from "../pages/AccountsPage";
import { ApiKeysPage } from "../pages/ApiKeysPage";
import { AuditLogsPage } from "../pages/AuditLogsPage";
import { LeasesPage } from "../pages/LeasesPage";
import type { AuthStore } from "../store/auth";
import { useAuthStore } from "../store/auth";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";

type SectionID = "overview" | "accounts" | "leases" | "api-keys" | "audit-logs";

type Props = {
  authStore?: AuthStore;
};

const sections: Array<{ id: SectionID; label: string }> = [
  { id: "overview", label: "Overview" },
  { id: "accounts", label: "Accounts" },
  { id: "leases", label: "Leases" },
  { id: "api-keys", label: "API Keys" },
  { id: "audit-logs", label: "Audit Logs" },
];

export function AdminShell({ authStore = useAuthStore }: Props) {
  const [activeSection, setActiveSection] = useState<SectionID>("overview");
  const user = authStore((state) => state.user);
  const logout = authStore((state) => state.logout);
  const loading = authStore((state) => state.loading);

  return (
    <div className="shell">
      <aside className="shell-sidebar">
        <div className="shell-brand">
          <div className="shell-brand__mark">AA</div>
          <div>
            <p className="shell-brand__name">Account Admin</p>
            <p className="shell-brand__meta">Operations console</p>
          </div>
        </div>
        <nav aria-label="Admin sections" className="shell-nav">
          {sections.map((section) => (
            <Button
              aria-current={activeSection === section.id ? "page" : undefined}
              className="shell-nav__item"
              key={section.id}
              onClick={() => setActiveSection(section.id)}
              type="button"
              variant="ghost"
            >
              {section.label}
            </Button>
          ))}
        </nav>
      </aside>
      <div className="shell-main">
        <header className="shell-topbar">
          <Badge variant="success">Service online</Badge>
          <div className="shell-user">
            <span>{user?.username}</span>
            <Button disabled={loading} onClick={() => void logout()} size="sm" type="button" variant="secondary">
              Sign out
            </Button>
          </div>
        </header>
        {activeSection === "overview" ? <OverviewPage /> : null}
        {activeSection === "accounts" ? (
          <div className="two-column-grid page">
            <AccountsPage />
            <AccountDetailPage />
          </div>
        ) : null}
        {activeSection === "leases" ? <LeasesPage /> : null}
        {activeSection === "api-keys" ? <ApiKeysPage /> : null}
        {activeSection === "audit-logs" ? <AuditLogsPage /> : null}
      </div>
    </div>
  );
}

function OverviewPage() {
  return (
    <main className="page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Operations overview</h1>
          <p className="page-description">Monitor account capacity, leases, callers, and audit activity from one console.</p>
        </div>
      </div>
      <div className="metric-grid">
        <MetricCard label="Primary workspace" value="Accounts" />
        <MetricCard label="Security posture" value="Session auth" />
        <MetricCard label="Lease flow" value="Managed" />
        <MetricCard label="Audit trail" value="Tracked" />
      </div>
    </main>
  );
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{label}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="metric-value">{value}</p>
      </CardContent>
    </Card>
  );
}
