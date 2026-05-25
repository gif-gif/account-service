import { FormEvent, useState } from "react";

import { OneTimeSecret } from "../components/OneTimeSecret";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { apiFetch } from "../lib/api";
import { useI18n } from "../store/settings";

export function ApiKeysPage() {
  const { t } = useI18n();
  const [apiKey, setApiKey] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const response = await apiFetch<{ api_key: string }>("/api/v1/api-keys", {
      method: "POST",
      body: JSON.stringify({
        name: form.get("name"),
        description: form.get("description"),
      }),
    });
    setApiKey(response.api_key);
  }

  return (
    <main className="page">
      <div className="page-header">
        <div>
          <h1 className="page-title">{t("apiKeys.title")}</h1>
          <p className="page-description">{t("apiKeys.description")}</p>
        </div>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>{t("apiKeys.cardTitle")}</CardTitle>
          <CardDescription>{t("apiKeys.cardDescription")}</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="form-grid" onSubmit={handleSubmit}>
            <Label className="form-row">
              {t("apiKeys.fieldName")}
              <Input name="name" />
            </Label>
            <Label className="form-row">
              {t("apiKeys.fieldDescription")}
              <Input name="description" />
            </Label>
            <Button type="submit">{t("apiKeys.create")}</Button>
          </form>
          {apiKey ? <OneTimeSecret label={t("apiKeys.secretLabel")} value={apiKey} /> : null}
        </CardContent>
      </Card>
    </main>
  );
}
