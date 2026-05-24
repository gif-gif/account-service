import { AccountDetailPage } from "./pages/AccountDetailPage";
import { AccountsPage } from "./pages/AccountsPage";
import { ApiKeysPage } from "./pages/ApiKeysPage";
import { AuditLogsPage } from "./pages/AuditLogsPage";
import { LeasesPage } from "./pages/LeasesPage";
import { LoginPage } from "./pages/LoginPage";

export function App() {
  return (
    <>
      <LoginPage />
      <AccountsPage />
      <AccountDetailPage />
      <LeasesPage />
      <ApiKeysPage />
      <AuditLogsPage />
    </>
  );
}
