import { FormEvent, useEffect, useState } from "react";

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

type ModelConfigKind = "fallback_model" | "hidden_model" | "model_alias" | "hidden_from_list";

type ModelConfigItem = {
  id: string;
  kind: ModelConfigKind;
  key: string;
  value: string;
  display_order: number;
  created_at: string;
  updated_at: string;
};

const modelConfigKinds: ModelConfigKind[] = ["fallback_model", "hidden_model", "model_alias", "hidden_from_list"];

type DialogState = { type: "create" } | { type: "edit"; item: ModelConfigItem } | { type: "delete"; item: ModelConfigItem } | null;

export function ModelConfigPage() {
  const { t } = useI18n();
  const [items, setItems] = useState<ModelConfigItem[]>([]);
  const [dialog, setDialog] = useState<DialogState>(null);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  async function load() {
    const response = await apiFetch<{ items: ModelConfigItem[] }>("/api/v1/model-config/items");
    setItems(response.items);
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
      kind: form.get("kind"),
      key: form.get("key"),
      value: form.get("value"),
      display_order: Number(form.get("display_order") || 0),
    };
    try {
      if (dialog.type === "edit") {
        const response = await apiFetch<{ item: ModelConfigItem }>(`/api/v1/model-config/items/${dialog.item.id}`, {
          method: "PATCH",
          body: JSON.stringify(payload),
        });
        setItems((current) => sortItems(current.map((item) => (item.id === response.item.id ? response.item : item))));
      } else {
        const response = await apiFetch<{ item: ModelConfigItem }>("/api/v1/model-config/items", {
          method: "POST",
          body: JSON.stringify(payload),
        });
        setItems((current) => sortItems([...current, response.item]));
      }
      setDialog(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("modelConfig.error"));
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(item: ModelConfigItem) {
    setSaving(true);
    setError("");
    try {
      await apiFetch<{ ok: boolean }>(`/api/v1/model-config/items/${item.id}`, { method: "DELETE" });
      setItems((current) => current.filter((candidate) => candidate.id !== item.id));
      setDialog(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t("modelConfig.error"));
    } finally {
      setSaving(false);
    }
  }

  return (
    <main className="page">
      <div className="page-header">
        <div>
          <h1 className="page-title">{t("modelConfig.title")}</h1>
          <p className="page-description">{t("modelConfig.description")}</p>
        </div>
        <Button
          onClick={() => {
            setError("");
            setDialog({ type: "create" });
          }}
          type="button"
        >
          {t("modelConfig.create")}
        </Button>
      </div>
      <div className="content-stack">
        <Card>
          <CardHeader>
            <CardTitle>{t("modelConfig.listTitle")}</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("modelConfig.kind")}</TableHead>
                  <TableHead>{t("modelConfig.key")}</TableHead>
                  <TableHead>{t("modelConfig.value")}</TableHead>
                  <TableHead>{t("modelConfig.order")}</TableHead>
                  <TableHead className="actions-col">{t("modelConfig.actions")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell>
                      <Badge variant="secondary">{item.kind}</Badge>
                    </TableCell>
                    <TableCell className="wide-cell">{item.key}</TableCell>
                    <TableCell className="wide-cell">{item.value || "-"}</TableCell>
                    <TableCell>{item.display_order}</TableCell>
                    <TableCell>
                      <div className="row-actions">
                        <Button
                          aria-label={`${t("common.edit")} ${item.key}`}
                          onClick={() => {
                            setError("");
                            setDialog({ type: "edit", item });
                          }}
                          size="sm"
                          type="button"
                          variant="secondary"
                        >
                          {t("common.edit")}
                        </Button>
                        <Button
                          aria-label={`${t("common.delete")} ${item.key}`}
                          onClick={() => {
                            setError("");
                            setDialog({ type: "delete", item });
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
            {items.length === 0 ? <p className="empty-state">{t("modelConfig.empty")}</p> : null}
          </CardContent>
        </Card>
      </div>
      <ModelConfigDialog dialog={dialog} error={error} saving={saving} onClose={() => setDialog(null)} onDelete={handleDelete} onSubmit={handleSubmit} />
    </main>
  );
}

function ModelConfigDialog({
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
  onDelete: (item: ModelConfigItem) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}) {
  const { t } = useI18n();
  const editingItem = dialog?.type === "edit" ? dialog.item : null;
  const deletingItem = dialog?.type === "delete" ? dialog.item : null;

  return (
    <Dialog open={dialog !== null} onOpenChange={(open) => (!open ? onClose() : undefined)}>
      {dialog?.type === "create" || dialog?.type === "edit" ? (
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editingItem ? t("modelConfig.editTitle") : t("modelConfig.createTitle")}</DialogTitle>
            <DialogDescription>{t("modelConfig.formDescription")}</DialogDescription>
          </DialogHeader>
          <form className="form-grid form-grid--two" key={editingItem?.id ?? "create"} onSubmit={onSubmit}>
            <Label className="form-row">
              {t("modelConfig.kind")}
              <select className="ui-select" defaultValue={editingItem?.kind ?? "fallback_model"} name="kind">
                {modelConfigKinds.map((kind) => (
                  <option key={kind} value={kind}>
                    {kind}
                  </option>
                ))}
              </select>
            </Label>
            <Label className="form-row">
              {t("modelConfig.key")}
              <Input defaultValue={editingItem?.key ?? ""} name="key" required />
            </Label>
            <Label className="form-row">
              {t("modelConfig.value")}
              <Input defaultValue={editingItem?.value ?? ""} name="value" />
            </Label>
            <Label className="form-row">
              {t("modelConfig.order")}
              <Input defaultValue={editingItem?.display_order ?? 0} min={0} name="display_order" type="number" />
            </Label>
            {error ? (
              <Alert className="form-row--wide" role="alert" variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            ) : null}
            <DialogFooter className="form-row--wide">
              <DialogClose asChild>
                <Button disabled={saving} type="button" variant="secondary">
                  {t("common.cancel")}
                </Button>
              </DialogClose>
              <Button disabled={saving} type="submit">
                {editingItem ? t("modelConfig.save") : t("modelConfig.create")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      ) : null}
      {deletingItem ? (
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{t("modelConfig.deleteTitle")}</DialogTitle>
            <DialogDescription>
              {t("modelConfig.deleteDescription")} <strong>{deletingItem.key}</strong>
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
            <Button disabled={saving} onClick={() => onDelete(deletingItem)} type="button" variant="destructive">
              {t("modelConfig.deleteConfirm")}
            </Button>
          </DialogFooter>
        </DialogContent>
      ) : null}
    </Dialog>
  );
}

function sortItems(items: ModelConfigItem[]) {
  return [...items].sort((left, right) => {
    if (left.kind !== right.kind) {
      return left.kind.localeCompare(right.kind);
    }
    if (left.display_order !== right.display_order) {
      return left.display_order - right.display_order;
    }
    return left.key.localeCompare(right.key);
  });
}
