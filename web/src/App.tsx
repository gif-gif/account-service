import { AdminShell } from "./components/AdminShell";
import { LoginPage } from "./pages/LoginPage";
import type { AuthStore } from "./store/auth";
import { useAuthStore } from "./store/auth";

type Props = {
  authStore?: AuthStore;
};

export function App({ authStore = useAuthStore }: Props) {
  const user = authStore((state) => state.user);

  if (!user) {
    return <LoginPage store={authStore} />;
  }

  return <AdminShell authStore={authStore} />;
}
