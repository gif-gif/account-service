import { AccountDetailPage } from "./pages/AccountDetailPage";
import { AccountsPage } from "./pages/AccountsPage";
import { LoginPage } from "./pages/LoginPage";

export function App() {
  return (
    <>
      <LoginPage />
      <AccountsPage />
      <AccountDetailPage />
    </>
  );
}
