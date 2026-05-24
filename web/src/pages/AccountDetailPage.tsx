import { FormEvent, useState } from "react";

import { Alert, AlertDescription } from "../components/ui/alert";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
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
    <main className="page">
      <Card>
        <CardHeader>
          <CardTitle>Account detail</CardTitle>
          <CardDescription>Create an account record. Sensitive fields stay inside the form until submitted.</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="form-grid" onSubmit={handleSubmit}>
            <Label>
              Username
              <Input name="username" />
            </Label>
            <Label>
              Password
              <Input name="password" type="password" />
            </Label>
            <Label>
              Login URL
              <Input name="login_url" />
            </Label>
            <Label>
              Access token
              <Input name="access_token" />
            </Label>
            <Label>
              Refresh token
              <Input name="refresh_token" />
            </Label>
            <Label>
              Region
              <Input name="region" />
            </Label>
            <Label>
              Account type
              <Input name="account_type" />
            </Label>
            <Label>
              Quota remaining
              <Input name="quota_remaining" type="number" />
            </Label>
            <Button type="submit">Save account</Button>
          </form>
          {savedID ? (
            <Alert>
              <AlertDescription>Saved {savedID}</AlertDescription>
            </Alert>
          ) : null}
        </CardContent>
      </Card>
    </main>
  );
}
