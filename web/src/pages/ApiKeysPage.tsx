import { FormEvent, useState } from "react";

import { OneTimeSecret } from "../components/OneTimeSecret";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { apiFetch } from "../lib/api";

export function ApiKeysPage() {
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
          <h1 className="page-title">API Keys</h1>
          <p className="page-description">Create caller credentials for service-to-service access.</p>
        </div>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Caller credentials</CardTitle>
          <CardDescription>The plaintext key is shown once after creation.</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="form-grid" onSubmit={handleSubmit}>
            <Label>
              Name
              <Input name="name" />
            </Label>
            <Label>
              Description
              <Input name="description" />
            </Label>
            <Button type="submit">Create API key</Button>
          </form>
          {apiKey ? <OneTimeSecret label="API key" value={apiKey} /> : null}
        </CardContent>
      </Card>
    </main>
  );
}
