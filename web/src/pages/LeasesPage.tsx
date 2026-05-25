import { FormEvent, useEffect, useState } from "react";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../components/ui/table";
import { apiFetch } from "../lib/api";
import { useI18n } from "../store/settings";

type Lease = {
  lease_id: string;
  account_id: string;
  status: string;
};

export function LeasesPage() {
  const { t } = useI18n();
  const [status, setStatus] = useState("");
  const [leases, setLeases] = useState<Lease[]>([]);

  async function load(nextStatus = status) {
    const query = nextStatus ? `?status=${encodeURIComponent(nextStatus)}` : "";
    const response = await apiFetch<{ leases: Lease[] }>(`/api/v1/leases${query}`);
    setLeases(response.leases);
  }

  useEffect(() => {
    void load("");
  }, []);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void load(status);
  }

  return (
    <main className="page">
      <div className="page-header">
        <div>
          <h1 className="page-title">{t("leases.title")}</h1>
          <p className="page-description">{t("leases.description")}</p>
        </div>
      </div>
      <div className="content-stack">
        <Card>
          <CardHeader>
            <CardTitle>{t("leases.filters")}</CardTitle>
            <CardDescription>{t("leases.cardDescription")}</CardDescription>
          </CardHeader>
          <CardContent>
            <form className="filter-grid" onSubmit={handleSubmit}>
              <Label className="form-row">
                {t("leases.status")}
                <Input value={status} onChange={(event) => setStatus(event.target.value)} />
              </Label>
              <Button type="submit">{t("leases.filter")}</Button>
            </form>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>{t("leases.activity")}</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("leases.leaseId")}</TableHead>
                  <TableHead>{t("leases.accountId")}</TableHead>
                  <TableHead>{t("common.status")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {leases.map((lease) => (
                  <TableRow key={lease.lease_id}>
                    <TableCell>{lease.lease_id}</TableCell>
                    <TableCell>{lease.account_id}</TableCell>
                    <TableCell>
                      <Badge className={leaseStatusClassName(lease.status)} variant="secondary">
                        {lease.status}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            {leases.length === 0 ? <p className="empty-state">{t("leases.empty")}</p> : null}
          </CardContent>
        </Card>
      </div>
    </main>
  );
}

function leaseStatusClassName(status: string) {
  if (status === "active") {
    return "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300";
  }
  if (status === "expired") {
    return "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300";
  }
  return undefined;
}
