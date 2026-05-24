import { FormEvent, useState } from "react";

import { apiFetch } from "../lib/api";

export function AccountDetailPage() {
  const [savedID, setSavedID] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const response = await apiFetch<{ account: { id: string } }>("/api/v1/accounts", {
      method: "POST",
      body: JSON.stringify({
        username: form.get("username"),
        password: form.get("password"),
        login_url: form.get("login_url"),
        access_token: form.get("access_token"),
        refresh_token: form.get("refresh_token"),
        region: form.get("region"),
        account_type: form.get("account_type"),
        status: "active",
        quota_remaining: Number(form.get("quota_remaining") || 0),
        max_concurrent_leases: 1,
        tags: [],
      }),
    });
    setSavedID(response.account.id);
  }

  return (
    <main>
      <h1>Account detail</h1>
      <form onSubmit={handleSubmit}>
        <label>
          Username
          <input name="username" />
        </label>
        <label>
          Password
          <input name="password" type="password" />
        </label>
        <label>
          Login URL
          <input name="login_url" />
        </label>
        <label>
          Access token
          <input name="access_token" />
        </label>
        <label>
          Refresh token
          <input name="refresh_token" />
        </label>
        <label>
          Region
          <input name="region" />
        </label>
        <label>
          Account type
          <input name="account_type" />
        </label>
        <label>
          Quota remaining
          <input name="quota_remaining" type="number" />
        </label>
        <button type="submit">Save account</button>
      </form>
      {savedID ? <p>Saved {savedID}</p> : null}
    </main>
  );
}
