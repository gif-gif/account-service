import { FormEvent, useState } from "react";

import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader } from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import type { AuthStore } from "../store/auth";
import { useAuthStore } from "../store/auth";
import { useI18n } from "../store/settings";

type Props = {
  store?: AuthStore;
};

export function LoginPage({ store = useAuthStore }: Props) {
  const { t } = useI18n();
  const login = store((state) => state.login);
  const loading = store((state) => state.loading);
  const error = store((state) => state.error);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await login(username, password);
  }

  return (
    <main className="login-page">
      <Card className="login-card">
        <CardHeader>
          <div className="login-mark">AA</div>
          <h1 className="ui-card__title">{t("app.brand")}</h1>
          <CardDescription>{t("login.description")}</CardDescription>
        </CardHeader>
        <CardContent>
          <form aria-label={t("login.form")} className="login-form" onSubmit={handleSubmit}>
            <Label className="form-row">
              {t("login.username")}
              <Input autoComplete="username" value={username} onChange={(event) => setUsername(event.target.value)} />
            </Label>
            <Label className="form-row">
              {t("login.password")}
              <Input
                autoComplete="current-password"
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
              />
            </Label>
            {error ? (
              <Alert role="alert" variant="destructive">
                <AlertTitle>{t("login.errorTitle")}</AlertTitle>
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            ) : null}
            <Button disabled={loading} type="submit">
              {loading ? t("login.submitting") : t("login.submit")}
            </Button>
          </form>
        </CardContent>
      </Card>
    </main>
  );
}
