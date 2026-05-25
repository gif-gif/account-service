import { useState } from "react";
import { BarChart3, ClipboardList, FileClock, KeyRound, Languages, LayoutDashboard, LogOut, Settings, UsersRound } from "lucide-react";

import { AccountsPage } from "../pages/AccountsPage";
import { ApiKeysPage } from "../pages/ApiKeysPage";
import { AuditLogsPage } from "../pages/AuditLogsPage";
import { LeasesPage } from "../pages/LeasesPage";
import type { AuthStore } from "../store/auth";
import { useAuthStore } from "../store/auth";
import { useI18n, type TranslationKey } from "../store/settings";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";

type SectionID = "overview" | "accounts" | "leases" | "api-keys" | "audit-logs";

type Props = {
  authStore?: AuthStore;
};

const sections: Array<{ id: SectionID; labelKey: TranslationKey; icon: typeof LayoutDashboard }> = [
  { id: "overview", labelKey: "nav.overview", icon: LayoutDashboard },
  { id: "accounts", labelKey: "nav.accounts", icon: UsersRound },
  { id: "leases", labelKey: "nav.leases", icon: ClipboardList },
  { id: "api-keys", labelKey: "nav.apiKeys", icon: KeyRound },
  { id: "audit-logs", labelKey: "nav.auditLogs", icon: FileClock },
];

export function AdminShell({ authStore = useAuthStore }: Props) {
  const { t, toggleLanguage } = useI18n();
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
            <p className="shell-brand__name">{t("app.brand")}</p>
            <p className="shell-brand__meta">{t("app.meta")}</p>
          </div>
        </div>
        <nav aria-label={t("nav.label")} className="shell-nav">
          {sections.map((section) => {
            const Icon = section.icon;
            return (
              <Button
                aria-current={activeSection === section.id ? "page" : undefined}
                className="shell-nav__item"
                key={section.id}
                onClick={() => setActiveSection(section.id)}
                type="button"
                variant="ghost"
              >
                <Icon />
                {t(section.labelKey)}
              </Button>
            );
          })}
        </nav>
      </aside>
      <div className="shell-main">
        <header className="shell-topbar">
          <div className="shell-topbar__left">
            <Badge className="bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300" variant="secondary">
              {t("common.serviceOnline")}
            </Badge>
            <span className="topbar-section">
              <BarChart3 />
              {t(sections.find((section) => section.id === activeSection)?.labelKey ?? "nav.overview")}
            </span>
          </div>
          <div className="shell-user">
            <span className="settings-label">
              <Settings />
              {t("common.settings")}
            </span>
            <Button onClick={toggleLanguage} size="sm" type="button" variant="secondary">
              <Languages />
              Language: {t("language.current")}
            </Button>
            <span>{user?.username}</span>
            <Button disabled={loading} onClick={() => void logout()} size="sm" type="button" variant="secondary">
              <LogOut />
              {t("common.signOut")}
            </Button>
          </div>
        </header>
        {activeSection === "overview" ? <OverviewPage /> : null}
        {activeSection === "accounts" ? <AccountsPage /> : null}
        {activeSection === "leases" ? <LeasesPage /> : null}
        {activeSection === "api-keys" ? <ApiKeysPage /> : null}
        {activeSection === "audit-logs" ? <AuditLogsPage /> : null}
      </div>
    </div>
  );
}

function OverviewPage() {
  const { t } = useI18n();

  return (
    <main className="page">
      <div className="page-header">
        <div>
          <h1 className="page-title">{t("overview.title")}</h1>
          <p className="page-description">{t("overview.description")}</p>
        </div>
      </div>
      <div className="metric-grid">
        <MetricCard label={t("overview.workspaceLabel")} value={t("overview.workspaceValue")} />
        <MetricCard label={t("overview.securityLabel")} value={t("overview.securityValue")} />
        <MetricCard label={t("overview.leaseLabel")} value={t("overview.leaseValue")} />
        <MetricCard label={t("overview.auditLabel")} value={t("overview.auditValue")} />
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
