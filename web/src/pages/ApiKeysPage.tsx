import { FormEvent, useEffect, useState } from "react";

import { OneTimeSecret } from "../components/OneTimeSecret";
import { Alert, AlertDescription } from "../components/ui/alert";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "../components/ui/dialog";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../components/ui/table";
import { apiFetch } from "../lib/api";
import { useI18n } from "../store/settings";

type Status = "active" | "disabled";

type Caller = {
  id: string;
  name: string;
  api_key?: string;
  status: Status;
  description: string;
  created_at: string;
  updated_at: string;
};

type DialogState = { type: "create" } | { type: "edit"; caller: Caller } | { type: "delete"; caller: Caller } | null;
type SecretDialogState = { title: string; value: string } | null;

export function ApiKeysPage() {
  const { t } = useI18n();
  const [callers, setCallers] = useState<Caller[]>([]);
  const [dialog, setDialog] = useState<DialogState>(null);
  const [secretDialog, setSecretDialog] = useState<SecretDialogState>(null);
  const [apiKey, setApiKey] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  async function load() {
    const response = await apiFetch<{ callers: Caller[] }>("/api/v1/api-keys");
    setCallers(response.callers);
  }

  useEffect(() => {
    void load().catch((err: Error) => setError(err.message));
  }, []);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!dialog || dialog.type === "delete") {
      return;
    }
    setSaving(true);
    setError("");
    const form = new FormData(event.currentTarget);
    const payload = {
      name: form.get("name"),
      description: form.get("description"),
      status: form.get("status"),
    };
    try {
      if (dialog.type === "edit") {
        const response = await apiFetch<{ caller: Caller }>(`/api/v1/api-keys/${dialog.caller.id}`, {
          method: "PATCH",
          body: JSON.stringify(payload),
        });
        setCallers((current) => sortCallers(current.map((caller) => (caller.id === response.caller.id ? response.caller : caller))));
      } else {
        const response = await apiFetch<{ caller: Caller; api_key: string }>("/api/v1/api-keys", {
          method: "POST",
          body: JSON.stringify(payload),
        });
        setCallers((current) => sortCallers([...current, response.caller]));
        setApiKey(response.api_key);
      }
      setDialog(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("apiKeys.error"));
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(caller: Caller) {
    setSaving(true);
    setError("");
    try {
      await apiFetch<{ ok: boolean }>(`/api/v1/api-keys/${caller.id}`, { method: "DELETE" });
      setCallers((current) => current.filter((candidate) => candidate.id !== caller.id));
      setDialog(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("apiKeys.error"));
    } finally {
      setSaving(false);
    }
  }

  async function revealAPIKey(caller: Caller) {
    setError("");
    try {
      const response = await apiFetch<{ api_key: string }>(`/api/v1/api-keys/${caller.id}/secret`, { method: "GET" });
      setSecretDialog({ title: t("apiKeys.secretLabel"), value: response.api_key });
    } catch (err) {
      setError(err instanceof Error ? err.message : t("apiKeys.error"));
    }
  }

  return (
    <main className="page">
      <div className="page-header">
        <div>
          <h1 className="page-title">{t("apiKeys.title")}</h1>
          <p className="page-description">{t("apiKeys.description")}</p>
        </div>
        <Button
          onClick={() => {
            setError("");
            setDialog({ type: "create" });
          }}
          type="button"
        >
          {t("apiKeys.create")}
        </Button>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>{t("apiKeys.cardTitle")}</CardTitle>
        </CardHeader>
        <CardContent>
          {apiKey ? <OneTimeSecret label={t("apiKeys.secretLabel")} value={apiKey} /> : null}
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("apiKeys.fieldName")}</TableHead>
                <TableHead>{t("apiKeys.fieldDescription")}</TableHead>
                <TableHead>{t("apiKeys.secretLabel")}</TableHead>
                <TableHead>{t("common.status")}</TableHead>
                <TableHead>{t("common.createdAt")}</TableHead>
                <TableHead>{t("common.updatedAt")}</TableHead>
                <TableHead className="actions-col">{t("modelConfig.actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {callers.map((caller) => (
                <TableRow key={caller.id}>
                  <TableCell className="wide-cell">{caller.name}</TableCell>
                  <TableCell className="wide-cell">{caller.description || "-"}</TableCell>
                  <TableCell>
                    <div className="secret-inline">
                      <code>••••••••</code>
                      <Button aria-label={`${t("common.reveal")} ${t("apiKeys.secretLabel")} ${caller.name}`} onClick={() => void revealAPIKey(caller)} size="sm" type="button" variant="secondary">
                        {t("common.reveal")}
                      </Button>
                    </div>
                  </TableCell>
                  <TableCell>
                    <StatusBadge status={caller.status} />
                  </TableCell>
                  <TableCell className="date-cell">{formatDate(caller.created_at)}</TableCell>
                  <TableCell className="date-cell">{formatDate(caller.updated_at)}</TableCell>
                  <TableCell>
                    <div className="row-actions">
                      <Button
                        aria-label={`${t("common.edit")} ${caller.name}`}
                        onClick={() => {
                          setError("");
                          setDialog({ type: "edit", caller });
                        }}
                        size="sm"
                        type="button"
                        variant="secondary"
                      >
                        {t("common.edit")}
                      </Button>
                      <Button
                        aria-label={`${t("common.delete")} ${caller.name}`}
                        onClick={() => {
                          setError("");
                          setDialog({ type: "delete", caller });
                        }}
                        size="sm"
                        type="button"
                        variant="destructive"
                      >
                        {t("common.delete")}
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          {callers.length === 0 ? <p className="empty-state">{t("apiKeys.empty")}</p> : null}
          {error && !dialog ? (
            <Alert role="alert" variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : null}
        </CardContent>
      </Card>
      <ApiKeyDialog dialog={dialog} error={error} saving={saving} onClose={() => setDialog(null)} onDelete={handleDelete} onSubmit={handleSubmit} />
      <SecretValueDialog dialog={secretDialog} onClose={() => setSecretDialog(null)} />
    </main>
  );
}

function SecretValueDialog({ dialog, onClose }: { dialog: SecretDialogState; onClose: () => void }) {
  const { t } = useI18n();
  return (
    <Dialog open={dialog !== null} onOpenChange={(open) => (!open ? onClose() : undefined)}>
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

function ApiKeyDialog({
  dialog,
  error,
  saving,
  onClose,
  onDelete,
  onSubmit,
}: {
  dialog: DialogState;
  error: string;
  saving: boolean;
  onClose: () => void;
  onDelete: (caller: Caller) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}) {
  const { t } = useI18n();
  const editingCaller = dialog?.type === "edit" ? dialog.caller : null;
  const deletingCaller = dialog?.type === "delete" ? dialog.caller : null;

  return (
    <Dialog open={dialog !== null} onOpenChange={(open) => (!open ? onClose() : undefined)}>
      {dialog?.type === "create" || dialog?.type === "edit" ? (
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingCaller ? t("apiKeys.editTitle") : t("apiKeys.create")}</DialogTitle>
            <DialogDescription>{t("apiKeys.cardDescription")}</DialogDescription>
          </DialogHeader>
          <form className="form-grid" key={editingCaller?.id ?? "create"} onSubmit={onSubmit}>
            <Label className="form-row">
              {t("apiKeys.fieldName")}
              <Input defaultValue={editingCaller?.name ?? ""} name="name" required />
            </Label>
            <Label className="form-row">
              {t("apiKeys.fieldDescription")}
              <Input defaultValue={editingCaller?.description ?? ""} name="description" />
            </Label>
            <Label className="form-row">
              {t("common.status")}
              <StatusSelect defaultValue={editingCaller?.status ?? "active"} />
            </Label>
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
                {editingCaller ? t("apiKeys.save") : t("apiKeys.create")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      ) : null}
      {deletingCaller ? (
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{t("apiKeys.deleteTitle")}</DialogTitle>
            <DialogDescription>
              {t("apiKeys.deleteDescription")} <strong>{deletingCaller.name}</strong>
            </DialogDescription>
          </DialogHeader>
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
            <Button disabled={saving} onClick={() => onDelete(deletingCaller)} type="button" variant="destructive">
              {t("apiKeys.deleteConfirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      ) : null}
    </Dialog>
  );
}

function StatusSelect({ defaultValue }: { defaultValue: Status }) {
  const { t } = useI18n();
  return (
    <select className="ui-select" defaultValue={defaultValue} name="status">
      <option value="active">{t("common.usable")}</option>
      <option value="disabled">{t("common.disabled")}</option>
    </select>
  );
}

function StatusBadge({ status }: { status: Status }) {
  const { t } = useI18n();
  return <Badge variant={status === "active" ? "default" : "secondary"}>{status === "active" ? t("common.usable") : t("common.disabled")}</Badge>;
}

function formatDate(value: string) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}

function sortCallers(callers: Caller[]) {
  return [...callers].sort((left, right) => {
    const dateCompare = new Date(right.created_at).getTime() - new Date(left.created_at).getTime();
    if (dateCompare !== 0) {
      return dateCompare;
    }
    return left.name.localeCompare(right.name);
  });
}
