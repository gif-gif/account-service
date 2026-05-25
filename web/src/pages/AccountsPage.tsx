import { FormEvent, useEffect, useMemo, useState } from "react";
import { Eye, Pencil, Plus, Search, Trash2 } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "../components/ui/dialog";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../components/ui/table";
import { apiFetch } from "../lib/api";
import type { Account, AccountsStore } from "../store/accounts";
import { useAccountsStore } from "../store/accounts";

type Props = {
  store?: AccountsStore;
};

type DialogState =
  | { type: "create" }
  | { type: "view"; account: Account }
  | { type: "edit"; account: Account }
  | { type: "delete"; account: Account }
  | null;

export function AccountsPage({ store = useAccountsStore }: Props) {
  const accounts = store((state) => state.accounts);
  const filters = store((state) => state.filters);
  const error = store((state) => state.error);
  const loading = store((state) => state.loading);
  const load = store((state) => state.load);
  const setFilter = store((state) => state.setFilter);
  const [dialog, setDialog] = useState<DialogState>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    void load();
  }, [load]);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void load();
  }

  const activeAccounts = accounts.filter((account) => account.status === "active").length;
  const totalQuota = accounts.reduce((total, account) => total + account.quota_remaining, 0);
  const errorStates = accounts.filter((account) => !["active", "available"].includes(account.status)).length;
  const regions = useMemo(() => new Set(accounts.map((account) => account.region).filter(Boolean)).size, [accounts]);

  async function submitAccount(event: FormEvent<HTMLFormElement>, account?: Account) {
    event.preventDefault();
    setSaving(true);
    setActionError(null);
    const payload = accountPayload(new FormData(event.currentTarget));
    try {
      if (account) {
        await apiFetch<{ account: Account }>(`/api/v1/accounts/${account.id}`, {
          method: "PATCH",
          body: JSON.stringify(payload),
        });
      } else {
        await apiFetch<{ account: Account }>("/api/v1/accounts", {
          method: "POST",
          body: JSON.stringify(payload),
        });
      }
      setDialog(null);
      await load();
    } catch (error) {
      setActionError(error instanceof Error ? error.message : "Failed to save account");
    } finally {
      setSaving(false);
    }
  }

  async function deleteAccount(account: Account) {
    setSaving(true);
    setActionError(null);
    try {
      await apiFetch<{ ok: boolean }>(`/api/v1/accounts/${account.id}`, { method: "DELETE" });
      setDialog(null);
      await load();
    } catch (error) {
      setActionError(error instanceof Error ? error.message : "Failed to delete account");
    } finally {
      setSaving(false);
    }
  }

  return (
    <main className="page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Accounts</h1>
          <p className="page-description">Manage account inventory, quota, credentials, and operational state.</p>
        </div>
        <Button onClick={() => setDialog({ type: "create" })} type="button">
          <Plus />
          New account
        </Button>
      </div>
      <div className="metric-grid">
        <MetricCard label="Active accounts" value={activeAccounts.toString()} />
        <MetricCard label="Total quota" value={totalQuota.toString()} />
        <MetricCard label="Regions" value={regions.toString()} />
        <MetricCard label="Error states" value={errorStates.toString()} />
      </div>
      <div className="content-stack">
        <Card className="admin-panel">
          <CardHeader className="admin-card-header">
            <CardTitle>Account inventory</CardTitle>
            <CardAction>
              <form className="admin-toolbar" onSubmit={handleSubmit}>
                <Label className="toolbar-field">
                  Region
                  <Input value={filters.region} onChange={(event) => setFilter("region", event.target.value)} />
                </Label>
                <Label className="toolbar-field">
                  Type
                  <Input value={filters.accountType} onChange={(event) => setFilter("accountType", event.target.value)} />
                </Label>
                <Label className="toolbar-field">
                  Status
                  <Input value={filters.status} onChange={(event) => setFilter("status", event.target.value)} />
                </Label>
                <Label className="toolbar-field toolbar-field--compact">
                  Min quota
                  <Input
                    min={0}
                    type="number"
                    value={filters.minQuotaRemaining}
                    onChange={(event) => setFilter("minQuotaRemaining", Number(event.target.value || 0))}
                  />
                </Label>
                <Button type="submit" variant="secondary">
                  <Search />
                  Apply filters
                </Button>
              </form>
            </CardAction>
          </CardHeader>
          <CardContent>
            <Label className="tag-filter">
              Tags
              <Input placeholder="openai, paid" value={filters.tags} onChange={(event) => setFilter("tags", event.target.value)} />
            </Label>
            {loading ? <p className="empty-state">Loading accounts</p> : null}
            {error ? (
              <Alert role="alert" variant="destructive">
                <AlertTitle>Accounts unavailable</AlertTitle>
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            ) : null}
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Account</TableHead>
                  <TableHead>Region</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Quota</TableHead>
                  <TableHead>Tags</TableHead>
                  <TableHead className="actions-col">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {accounts.map((account) => (
                  <TableRow key={account.id}>
                    <TableCell>
                      <div className="account-cell">
                        <span className="account-name">{account.username}</span>
                        <span className="account-subtext">{account.login_url || "No login URL"}</span>
                      </div>
                    </TableCell>
                    <TableCell>{account.region || "-"}</TableCell>
                    <TableCell>{account.account_type || "-"}</TableCell>
                    <TableCell>
                      <StatusBadge status={account.status} />
                    </TableCell>
                    <TableCell className="numeric-cell">{account.quota_remaining}</TableCell>
                    <TableCell>
                      <div className="tag-list">
                        {(account.tags ?? []).slice(0, 2).map((tag) => (
                          <Badge key={tag} variant="secondary">
                            {tag}
                          </Badge>
                        ))}
                        {(account.tags ?? []).length === 0 ? <span className="muted-text">-</span> : null}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="row-actions">
                        <Button aria-label={`View ${account.username}`} onClick={() => setDialog({ type: "view", account })} size="icon-sm" type="button" variant="ghost">
                          <Eye />
                        </Button>
                        <Button aria-label={`Edit ${account.username}`} onClick={() => setDialog({ type: "edit", account })} size="icon-sm" type="button" variant="ghost">
                          <Pencil />
                        </Button>
                        <Button
                          aria-label={`Delete ${account.username}`}
                          onClick={() => setDialog({ type: "delete", account })}
                          size="icon-sm"
                          type="button"
                          variant="ghost"
                        >
                          <Trash2 />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            {!loading && accounts.length === 0 && !error ? <p className="empty-state">No accounts</p> : null}
          </CardContent>
        </Card>
      </div>
      <AccountDialogs dialog={dialog} actionError={actionError} saving={saving} onClose={() => setDialog(null)} onDelete={deleteAccount} onSubmit={submitAccount} />
    </main>
  );
}

function AccountDialogs({
  actionError,
  dialog,
  onClose,
  onDelete,
  onSubmit,
  saving,
}: {
  actionError: string | null;
  dialog: DialogState;
  onClose: () => void;
  onDelete: (account: Account) => Promise<void>;
  onSubmit: (event: FormEvent<HTMLFormElement>, account?: Account) => Promise<void>;
  saving: boolean;
}) {
  return (
    <Dialog open={dialog !== null} onOpenChange={(open) => (!open ? onClose() : undefined)}>
      {dialog?.type === "create" ? (
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New account</DialogTitle>
            <DialogDescription>Create an account record for allocation and quota tracking.</DialogDescription>
          </DialogHeader>
          <AccountForm error={actionError} saving={saving} submitLabel="Create account" onSubmit={(event) => void onSubmit(event)} />
        </DialogContent>
      ) : null}
      {dialog?.type === "edit" ? (
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit account</DialogTitle>
            <DialogDescription>Update credentials, status, quota, or routing metadata.</DialogDescription>
          </DialogHeader>
          <AccountForm account={dialog.account} error={actionError} saving={saving} submitLabel="Save changes" onSubmit={(event) => void onSubmit(event, dialog.account)} />
        </DialogContent>
      ) : null}
      {dialog?.type === "view" ? (
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Account details</DialogTitle>
            <DialogDescription>{dialog.account.username}</DialogDescription>
          </DialogHeader>
          <div className="detail-grid">
            <DetailItem label="Region" value={dialog.account.region} />
            <DetailItem label="Type" value={dialog.account.account_type} />
            <DetailItem label="Status" value={dialog.account.status} />
            <DetailItem label="Quota remaining" value={dialog.account.quota_remaining.toString()} />
            <DetailItem label="Login URL" value={dialog.account.login_url || "-"} />
            <DetailItem label="Notes" value={dialog.account.notes || "-"} />
          </div>
          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="secondary">
                Close
              </Button>
            </DialogClose>
          </DialogFooter>
        </DialogContent>
      ) : null}
      {dialog?.type === "delete" ? (
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Delete account</DialogTitle>
            <DialogDescription>Delete {dialog.account.username}. This action cannot be undone.</DialogDescription>
          </DialogHeader>
          {actionError ? (
            <Alert role="alert" variant="destructive">
              <AlertDescription>{actionError}</AlertDescription>
            </Alert>
          ) : null}
          <DialogFooter>
            <DialogClose asChild>
              <Button disabled={saving} type="button" variant="secondary">
                Cancel
              </Button>
            </DialogClose>
            <Button disabled={saving} onClick={() => void onDelete(dialog.account)} type="button" variant="destructive">
              Delete account
            </Button>
          </DialogFooter>
        </DialogContent>
      ) : null}
    </Dialog>
  );
}

function AccountForm({
  account,
  error,
  onSubmit,
  saving,
  submitLabel,
}: {
  account?: Account;
  error: string | null;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  saving: boolean;
  submitLabel: string;
}) {
  return (
    <form className="account-form" onSubmit={onSubmit}>
      <div className="form-grid form-grid--two">
        <Label className="form-row">
          Username
          <Input defaultValue={account?.username ?? ""} name="username" required />
        </Label>
        <Label className="form-row">
          Password
          <Input defaultValue={account?.password ?? ""} name="password" type="password" />
        </Label>
        <Label className="form-row">
          Login URL
          <Input defaultValue={account?.login_url ?? ""} name="login_url" />
        </Label>
        <Label className="form-row">
          Region
          <Input defaultValue={account?.region ?? ""} name="region" />
        </Label>
        <Label className="form-row">
          Account type
          <Input defaultValue={account?.account_type ?? ""} name="account_type" />
        </Label>
        <Label className="form-row">
          Status
          <Input defaultValue={account?.status ?? "active"} name="status" />
        </Label>
        <Label className="form-row">
          Quota total
          <Input defaultValue={account?.quota_total ?? 0} min={0} name="quota_total" type="number" />
        </Label>
        <Label className="form-row">
          Quota used
          <Input defaultValue={account?.quota_used ?? 0} min={0} name="quota_used" type="number" />
        </Label>
        <Label className="form-row">
          Quota remaining
          <Input defaultValue={account?.quota_remaining ?? 0} min={0} name="quota_remaining" type="number" />
        </Label>
        <Label className="form-row">
          Max leases
          <Input defaultValue={account?.max_concurrent_leases ?? 1} min={1} name="max_concurrent_leases" type="number" />
        </Label>
        <Label className="form-row form-row--wide">
          Provider access token
          <Input defaultValue={account?.access_token ?? ""} name="access_token" />
        </Label>
        <Label className="form-row form-row--wide">
          Provider refresh token
          <Input defaultValue={account?.refresh_token ?? ""} name="refresh_token" />
        </Label>
        <Label className="form-row form-row--wide">
          Tags
          <Input defaultValue={(account?.tags ?? []).join(", ")} name="tags" placeholder="openai, paid" />
        </Label>
        <Label className="form-row form-row--wide">
          Notes
          <Input defaultValue={account?.notes ?? ""} name="notes" />
        </Label>
      </div>
      {error ? (
        <Alert role="alert" variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <DialogFooter>
        <DialogClose asChild>
          <Button disabled={saving} type="button" variant="secondary">
            Cancel
          </Button>
        </DialogClose>
        <Button disabled={saving} type="submit">
          {submitLabel}
        </Button>
      </DialogFooter>
    </form>
  );
}

function DetailItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="detail-item">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  if (status === "active") {
    return <Badge className="bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300">active</Badge>;
  }
  if (status === "error" || status === "token_expired" || status === "login_failed") {
    return <Badge variant="destructive">{status}</Badge>;
  }
  return <Badge variant="secondary">{status || "-"}</Badge>;
}

function accountPayload(form: FormData) {
  return {
    username: stringValue(form, "username"),
    password: stringValue(form, "password"),
    login_url: stringValue(form, "login_url"),
    access_token: stringValue(form, "access_token"),
    refresh_token: stringValue(form, "refresh_token"),
    region: stringValue(form, "region"),
    account_type: stringValue(form, "account_type"),
    status: stringValue(form, "status") || "active",
    quota_total: numberValue(form, "quota_total"),
    quota_used: numberValue(form, "quota_used"),
    quota_remaining: numberValue(form, "quota_remaining"),
    max_concurrent_leases: numberValue(form, "max_concurrent_leases") || 1,
    tags: stringValue(form, "tags")
      .split(",")
      .map((tag) => tag.trim())
      .filter(Boolean),
    notes: stringValue(form, "notes"),
  };
}

function stringValue(form: FormData, key: string) {
  return String(form.get(key) ?? "").trim();
}

function numberValue(form: FormData, key: string) {
  return Number(form.get(key) || 0);
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <Card>
      <CardContent>
        <p className="metric-label">{label}</p>
        <p className="metric-value">{value}</p>
      </CardContent>
    </Card>
  );
}
