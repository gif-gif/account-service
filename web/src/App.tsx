import { useEffect } from "react";

import { AdminShell } from "./components/AdminShell";
import { getAuthTokens } from "./lib/authTokens";
import { LoginPage } from "./pages/LoginPage";
import type { AuthStore } from "./store/auth";
import { useAuthStore } from "./store/auth";

type Props = {
  authStore?: AuthStore;
};

export function App({ authStore = useAuthStore }: Props) {
  const user = authStore((state) => state.user);
  const restore = authStore((state) => state.restore);

  useEffect(() => {
    if (!user && getAuthTokens()) {
      void restore();
    }
  }, [restore, user]);

  if (!user) {
    return <LoginPage store={authStore} />;
  }

  return <AdminShell authStore={authStore} />;
}
