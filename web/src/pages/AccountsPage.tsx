import { type FormEvent, type ReactNode, useEffect, useMemo, useState } from "react";
import { Eye, Pencil, Plus, Search, Trash2 } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "../components/ui/dialog";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../components/ui/table";
import { apiFetch } from "../lib/api";
import type { Account, AccountsStore } from "../store/accounts";
import { accountTypes } from "../store/accounts";
import { useAccountsStore } from "../store/accounts";
import { useI18n, type TranslationKey } from "../store/settings";

type Props = {
  store?: AccountsStore;
};

type DialogState =
  | { type: "create" }
  | { type: "view"; account: Account }
  | { type: "edit"; account: Account }
  | { type: "delete"; account: Account }
  | null;

type SecretDialogState = { title: string; value: string } | null;

const accountStatuses = ["active", "disabled", "exhausted", "login_failed", "token_expired", "region_blocked", "error"] as const;

export function AccountsPage({ store = useAccountsStore }: Props) {
  const { t } = useI18n();
  const accounts = store((state) => state.accounts);
  const filters = store((state) => state.filters);
  const error = store((state) => state.error);
  const loading = store((state) => state.loading);
  const load = store((state) => state.load);
  const setFilter = store((state) => state.setFilter);
  const [dialog, setDialog] = useState<DialogState>(null);
  const [secretDialog, setSecretDialog] = useState<SecretDialogState>(null);
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
          <h1 className="page-title">{t("accounts.title")}</h1>
          <p className="page-description">{t("accounts.description")}</p>
        </div>
        <Button onClick={() => setDialog({ type: "create" })} type="button">
          <Plus />
          {t("accounts.add")}
        </Button>
      </div>
      <div className="metric-grid">
        <MetricCard label={t("accounts.activeAccounts")} value={activeAccounts.toString()} />
        <MetricCard label={t("accounts.quotaRemaining")} value={totalQuota.toString()} />
        <MetricCard label={t("accounts.regions")} value={regions.toString()} />
        <MetricCard label={t("accounts.errorStates")} value={errorStates.toString()} />
      </div>
      <div className="content-stack">
        <Card className="filter-panel" role="group" aria-label={t("accounts.filterTitle")}>
          <CardHeader className="filter-panel__header">
            <CardTitle>{t("accounts.filterTitle")}</CardTitle>
            <CardDescription>{t("accounts.description")}</CardDescription>
          </CardHeader>
          <CardContent>
            <form className="account-filter-grid" onSubmit={handleSubmit}>
              <Label className="toolbar-field">
                {t("accounts.region")}
                <Input value={filters.region} onChange={(event) => setFilter("region", event.target.value)} />
              </Label>
              <Label className="toolbar-field">
                {t("accounts.type")}
                <select className="ui-select" value={filters.accountType} onChange={(event) => setFilter("accountType", event.target.value)}>
                  <option value="">{t("accounts.typeAll")}</option>
                  {accountTypes.map((accountType) => (
                    <option key={accountType} value={accountType}>
                      {accountType}
                    </option>
                  ))}
                </select>
              </Label>
              <Label className="toolbar-field">
                {t("accounts.status")}
                <select className="ui-select" value={filters.status} onChange={(event) => setFilter("status", event.target.value)}>
                  <option value="">{t("accounts.statusAll")}</option>
                  {accountStatuses.map((status) => (
                    <option key={status} value={status}>
                      {status}
                    </option>
                  ))}
                </select>
              </Label>
              <Label className="toolbar-field toolbar-field--compact">
                {t("accounts.minQuota")}
                <Input
                  min={0}
                  type="number"
                  value={filters.minQuotaRemaining}
                  onChange={(event) => setFilter("minQuotaRemaining", Number(event.target.value || 0))}
                />
              </Label>
              <Label className="tag-filter">
                {t("accounts.tags")}
                <Input placeholder={t("accounts.tagsPlaceholder")} value={filters.tags} onChange={(event) => setFilter("tags", event.target.value)} />
              </Label>
              <Button className="filter-submit" type="submit" variant="secondary">
                <Search />
                {t("accounts.applyFilters")}
              </Button>
            </form>
          </CardContent>
        </Card>
        <Card className="admin-panel">
          <CardHeader className="admin-card-header">
            <CardTitle>{t("accounts.inventory")}</CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? <p className="empty-state">{t("accounts.loading")}</p> : null}
            {error ? (
              <Alert role="alert" variant="destructive">
                <AlertTitle>{t("accounts.errorTitle")}</AlertTitle>
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            ) : null}
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("accounts.id")}</TableHead>
                  <TableHead>{t("accounts.username")}</TableHead>
                  <TableHead>{t("accounts.password")}</TableHead>
                  <TableHead>{t("accounts.loginUrl")}</TableHead>
                  <TableHead>{t("accounts.accessToken")}</TableHead>
                  <TableHead>{t("accounts.refreshToken")}</TableHead>
                  <TableHead>{t("accounts.region")}</TableHead>
                  <TableHead>{t("accounts.accountType")}</TableHead>
                  <TableHead>{t("accounts.status")}</TableHead>
                  <TableHead>{t("accounts.quotaTotal")}</TableHead>
                  <TableHead>{t("accounts.quotaUsed")}</TableHead>
                  <TableHead>{t("accounts.quotaRemaining")}</TableHead>
                  <TableHead>{t("accounts.maxLeases")}</TableHead>
                  <TableHead>{t("accounts.tags")}</TableHead>
                  <TableHead>{t("accounts.notes")}</TableHead>
                  <TableHead>{t("accounts.createdAt")}</TableHead>
                  <TableHead>{t("accounts.updatedAt")}</TableHead>
                  <TableHead className="actions-col">{t("common.edit")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {accounts.map((account) => (
                  <TableRow key={account.id}>
                    <TableCell>
                      <SecretButton label={t("accounts.id")} value={account.id} account={account} onReveal={(title, value) => setSecretDialog({ title, value })} t={t} />
                    </TableCell>
                    <TableCell className="account-name">{account.username}</TableCell>
                    <TableCell>
                      <SecretButton label={t("accounts.password")} value={account.password} account={account} onReveal={(title, value) => setSecretDialog({ title, value })} t={t} />
                    </TableCell>
                    <TableCell className="wide-cell">{account.login_url || t("accounts.noLoginUrl")}</TableCell>
                    <TableCell>
                      <SecretButton label={t("accounts.accessToken")} value={account.access_token} account={account} onReveal={(title, value) => setSecretDialog({ title, value })} t={t} />
                    </TableCell>
                    <TableCell>
                      <SecretButton label={t("accounts.refreshToken")} value={account.refresh_token} account={account} onReveal={(title, value) => setSecretDialog({ title, value })} t={t} />
                    </TableCell>
                    <TableCell>{account.region || "-"}</TableCell>
                    <TableCell>{account.account_type || "-"}</TableCell>
                    <TableCell>
                      <StatusBadge status={account.status} />
                    </TableCell>
                    <TableCell className="numeric-cell">{account.quota_total ?? 0}</TableCell>
                    <TableCell className="numeric-cell">{account.quota_used ?? 0}</TableCell>
                    <TableCell className="numeric-cell">{account.quota_remaining}</TableCell>
                    <TableCell className="numeric-cell">{account.max_concurrent_leases ?? 1}</TableCell>
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
                    <TableCell className="wide-cell">{account.notes || "-"}</TableCell>
                    <TableCell className="date-cell">{formatDate(account.created_at)}</TableCell>
                    <TableCell className="date-cell">{formatDate(account.updated_at)}</TableCell>
                    <TableCell>
                      <div className="row-actions">
                        <Button
                          aria-label={`${t("common.view")} ${account.username}`}
                          onClick={() => setDialog({ type: "view", account })}
                          size="icon-sm"
                          title={t("common.view")}
                          type="button"
                          variant="ghost"
                        >
                          <Eye />
                        </Button>
                        <Button
                          aria-label={`${t("common.edit")} ${account.username}`}
                          onClick={() => setDialog({ type: "edit", account })}
                          size="icon-sm"
                          title={t("common.edit")}
                          type="button"
                          variant="ghost"
                        >
                          <Pencil />
                        </Button>
                        <Button
                          aria-label={`${t("common.delete")} ${account.username}`}
                          onClick={() => setDialog({ type: "delete", account })}
                          size="icon-sm"
                          title={t("common.delete")}
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
            {!loading && accounts.length === 0 && !error ? <p className="empty-state">{t("accounts.empty")}</p> : null}
          </CardContent>
        </Card>
      </div>
      <AccountDialogs
        actionError={actionError}
        dialog={dialog}
        saving={saving}
        secretOpen={secretDialog !== null}
        onClose={() => setDialog(null)}
        onDelete={deleteAccount}
        onReveal={(title, value) => setSecretDialog({ title, value })}
        onSubmit={submitAccount}
        t={t}
      />
      <SecretValueDialog dialog={secretDialog} onClose={() => setSecretDialog(null)} t={t} />
    </main>
  );
}

function AccountDialogs({
  actionError,
  dialog,
  onClose,
  onDelete,
  onReveal,
  onSubmit,
  saving,
  secretOpen,
  t,
}: {
  actionError: string | null;
  dialog: DialogState;
  onClose: () => void;
  onDelete: (account: Account) => Promise<void>;
  onReveal: (title: string, value: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>, account?: Account) => Promise<void>;
  saving: boolean;
  secretOpen: boolean;
  t: (key: TranslationKey) => string;
}) {
  return (
    <Dialog open={dialog !== null} onOpenChange={(open) => (!open ? onClose() : undefined)}>
      {dialog?.type === "create" ? (
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("accounts.add")}</DialogTitle>
            <DialogDescription>{t("accounts.createDescription")}</DialogDescription>
          </DialogHeader>
          <AccountForm error={actionError} saving={saving} submitLabel={t("common.create")} onSubmit={(event) => void onSubmit(event)} t={t} />
        </DialogContent>
      ) : null}
      {dialog?.type === "edit" ? (
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("accounts.editTitle")}</DialogTitle>
            <DialogDescription>{t("accounts.editDescription")}</DialogDescription>
          </DialogHeader>
          <AccountForm
            account={dialog.account}
            error={actionError}
            saving={saving}
            submitLabel={t("common.save")}
            onSubmit={(event) => void onSubmit(event, dialog.account)}
            t={t}
          />
        </DialogContent>
      ) : null}
      {dialog?.type === "view" ? (
        <DialogContent className="max-w-2xl" onInteractOutside={(event) => (secretOpen ? event.preventDefault() : undefined)}>
          <DialogHeader>
            <DialogTitle>{t("accounts.detailsTitle")}</DialogTitle>
            <DialogDescription>{dialog.account.username}</DialogDescription>
          </DialogHeader>
          <AccountDetails account={dialog.account} onReveal={onReveal} t={t} />
          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="secondary">
                {t("common.close")}
              </Button>
            </DialogClose>
          </DialogFooter>
        </DialogContent>
      ) : null}
      {dialog?.type === "delete" ? (
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{t("accounts.deleteTitle")}</DialogTitle>
            <DialogDescription>
              {t("accounts.deleteDescriptionPrefix")} {dialog.account.username}
              {t("accounts.deleteDescriptionSuffix")}
            </DialogDescription>
          </DialogHeader>
          {actionError ? (
            <Alert role="alert" variant="destructive">
              <AlertDescription>{actionError}</AlertDescription>
            </Alert>
          ) : null}
          <DialogFooter>
            <DialogClose asChild>
              <Button disabled={saving} type="button" variant="secondary">
                {t("common.cancel")}
              </Button>
            </DialogClose>
            <Button disabled={saving} onClick={() => void onDelete(dialog.account)} type="button" variant="destructive">
              {t("accounts.deleteConfirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      ) : null}
    </Dialog>
  );
}

function SecretValueDialog({ dialog, onClose, t }: { dialog: SecretDialogState; onClose: () => void; t: (key: TranslationKey) => string }) {
  return (
    <Dialog modal={false} open={dialog !== null} onOpenChange={(open) => (!open ? onClose() : undefined)}>
      {dialog ? (
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{dialog.title}</DialogTitle>
          </DialogHeader>
          <code className="secret-value">{dialog.value || "-"}</code>
          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="secondary">
                {t("common.close")}
              </Button>
            </DialogClose>
          </DialogFooter>
        </DialogContent>
      ) : null}
    </Dialog>
  );
}

function AccountDetails({
  account,
  onReveal,
  t,
}: {
  account: Account;
  onReveal: (title: string, value: string) => void;
  t: (key: TranslationKey) => string;
}) {
  return (
    <div className="detail-grid">
      <DetailItem label={t("accounts.id")} value={<code className="metadata-code">{account.id}</code>} />
      <DetailItem label={t("accounts.username")} value={account.username || "-"} />
      <DetailItem
        label={t("accounts.password")}
        value={<SecretButton label={t("accounts.password")} value={account.password} account={account} onReveal={onReveal} t={t} />}
      />
      <DetailItem label={t("accounts.loginUrl")} value={account.login_url || "-"} />
      <DetailItem
        label={t("accounts.accessToken")}
        value={<SecretButton label={t("accounts.accessToken")} value={account.access_token} account={account} onReveal={onReveal} t={t} />}
      />
      <DetailItem
        label={t("accounts.refreshToken")}
        value={<SecretButton label={t("accounts.refreshToken")} value={account.refresh_token} account={account} onReveal={onReveal} t={t} />}
      />
      <DetailItem label={t("accounts.region")} value={account.region || "-"} />
      <DetailItem label={t("accounts.accountType")} value={account.account_type || "-"} />
      <DetailItem label={t("accounts.status")} value={<StatusBadge status={account.status} />} />
      <DetailItem label={t("accounts.quotaTotal")} value={account.quota_total ?? 0} />
      <DetailItem label={t("accounts.quotaUsed")} value={account.quota_used ?? 0} />
      <DetailItem label={t("accounts.quotaRemaining")} value={account.quota_remaining ?? 0} />
      <DetailItem label={t("accounts.maxLeases")} value={account.max_concurrent_leases ?? 1} />
      <DetailItem
        label={t("accounts.tags")}
        value={
          (account.tags ?? []).length > 0 ? (
            <div className="tag-list">
              {(account.tags ?? []).map((tag) => (
                <Badge key={tag} variant="secondary">
                  {tag}
                </Badge>
              ))}
            </div>
          ) : (
            "-"
          )
        }
      />
      <DetailItem label={t("accounts.notes")} value={account.notes || "-"} />
      <DetailItem label={t("accounts.createdAt")} value={formatDate(account.created_at)} />
      <DetailItem label={t("accounts.updatedAt")} value={formatDate(account.updated_at)} />
    </div>
  );
}

function SecretButton({
  account,
  label,
  onReveal,
  t,
  value,
}: {
  account: Account;
  label: string;
  onReveal: (title: string, value: string) => void;
  t: (key: TranslationKey) => string;
  value?: string;
}) {
  return (
    <Button
      aria-label={`${t("common.reveal")} ${label} ${account.username}`}
      onClick={() => onReveal(label, value ?? "")}
      size="sm"
      type="button"
      variant="secondary"
    >
      <Eye />
      ******
    </Button>
  );
}

function AccountForm({
  account,
  error,
  onSubmit,
  saving,
  submitLabel,
  t,
}: {
  account?: Account;
  error: string | null;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  saving: boolean;
  submitLabel: string;
  t: (key: TranslationKey) => string;
}) {
  return (
    <form className="account-form" onSubmit={onSubmit}>
      <div className="form-grid form-grid--two">
        <Label className="form-row">
          {t("accounts.username")}
          <Input defaultValue={account?.username ?? ""} name="username" required />
        </Label>
        <Label className="form-row">
          {t("accounts.password")}
          <Input defaultValue={account?.password ?? ""} name="password" type="password" />
        </Label>
        <Label className="form-row">
          {t("accounts.loginUrl")}
          <Input defaultValue={account?.login_url ?? ""} name="login_url" />
        </Label>
        <Label className="form-row">
          {t("accounts.region")}
          <Input defaultValue={account?.region ?? ""} name="region" />
        </Label>
        <Label className="form-row">
          {t("accounts.accountType")}
          <select className="ui-select" defaultValue={account?.account_type ?? accountTypes[0]} name="account_type">
            {accountTypes.map((accountType) => (
              <option key={accountType} value={accountType}>
                {accountType}
              </option>
            ))}
          </select>
        </Label>
        <Label className="form-row">
          {t("accounts.status")}
          <select className="ui-select" defaultValue={account?.status ?? "active"} name="status">
            {accountStatuses.map((status) => (
              <option key={status} value={status}>
                {status}
              </option>
            ))}
          </select>
        </Label>
        <Label className="form-row">
          {t("accounts.quotaTotal")}
          <Input defaultValue={account?.quota_total ?? 0} min={0} name="quota_total" type="number" />
        </Label>
        <Label className="form-row">
          {t("accounts.quotaUsed")}
          <Input defaultValue={account?.quota_used ?? 0} min={0} name="quota_used" type="number" />
        </Label>
        <Label className="form-row">
          {t("accounts.quotaRemaining")}
          <Input defaultValue={account?.quota_remaining ?? 0} min={0} name="quota_remaining" type="number" />
        </Label>
        <Label className="form-row">
          {t("accounts.maxLeases")}
          <Input defaultValue={account?.max_concurrent_leases ?? 1} min={1} name="max_concurrent_leases" type="number" />
        </Label>
        <Label className="form-row form-row--wide">
          {t("accounts.accessToken")}
          <Input defaultValue={account?.access_token ?? ""} name="access_token" />
        </Label>
        <Label className="form-row form-row--wide">
          {t("accounts.refreshToken")}
          <Input defaultValue={account?.refresh_token ?? ""} name="refresh_token" />
        </Label>
        <Label className="form-row form-row--wide">
          {t("accounts.tags")}
          <Input defaultValue={(account?.tags ?? []).join(", ")} name="tags" placeholder={t("accounts.tagsPlaceholder")} />
        </Label>
        <Label className="form-row form-row--wide">
          {t("accounts.notes")}
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
            {t("common.cancel")}
          </Button>
        </DialogClose>
        <Button disabled={saving} type="submit">
          {submitLabel}
        </Button>
      </DialogFooter>
    </form>
  );
}

function DetailItem({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="detail-item">
      <span>{label}</span>
      <div className="detail-value">{value}</div>
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

function formatDate(value?: string) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
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
