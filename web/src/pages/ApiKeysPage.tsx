import { FormEvent, useState } from "react";

import { OneTimeSecret } from "../components/OneTimeSecret";
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
    <main>
      <h1>API Keys</h1>
      <form onSubmit={handleSubmit}>
        <label>
          Name
          <input name="name" />
        </label>
        <label>
          Description
          <input name="description" />
        </label>
        <button type="submit">Create API key</button>
      </form>
      {apiKey ? <OneTimeSecret label="API key" value={apiKey} /> : null}
    </main>
  );
}
