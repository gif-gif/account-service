import { FormEvent, useEffect } from "react";

import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../components/ui/table";
import type { AccountsStore } from "../store/accounts";
import { useAccountsStore } from "../store/accounts";

type Props = {
  store?: AccountsStore;
};

export function AccountsPage({ store = useAccountsStore }: Props) {
  const accounts = store((state) => state.accounts);
  const filters = store((state) => state.filters);
  const error = store((state) => state.error);
  const loading = store((state) => state.loading);
  const load = store((state) => state.load);
  const setFilter = store((state) => state.setFilter);

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

  return (
    <main className="page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Accounts</h1>
          <p className="page-description">Query account capacity and operational state.</p>
        </div>
      </div>
      <div className="metric-grid">
        <MetricCard label="Active accounts" value={activeAccounts.toString()} />
        <MetricCard label="Total quota" value={totalQuota.toString()} />
        <MetricCard label="Active leases" value="-" />
        <MetricCard label="Error states" value={errorStates.toString()} />
      </div>
      <div className="content-stack">
        <Card>
          <CardHeader>
            <CardTitle>Filters</CardTitle>
          </CardHeader>
          <CardContent>
            <form className="filter-grid" onSubmit={handleSubmit}>
              <Label className="form-row">
                Region
                <Input value={filters.region} onChange={(event) => setFilter("region", event.target.value)} />
              </Label>
              <Label className="form-row">
                Account type
                <Input value={filters.accountType} onChange={(event) => setFilter("accountType", event.target.value)} />
              </Label>
              <Label className="form-row">
                Status
                <Input value={filters.status} onChange={(event) => setFilter("status", event.target.value)} />
              </Label>
              <Label className="form-row">
                Tags
                <Input value={filters.tags} onChange={(event) => setFilter("tags", event.target.value)} />
              </Label>
              <Label className="form-row">
                Minimum quota
                <Input
                  min={0}
                  type="number"
                  value={filters.minQuotaRemaining}
                  onChange={(event) => setFilter("minQuotaRemaining", Number(event.target.value || 0))}
                />
              </Label>
              <Button type="submit">Apply filters</Button>
            </form>
          </CardContent>
        </Card>
        {loading ? <p className="empty-state">Loading accounts</p> : null}
        {error ? (
          <Alert role="alert" variant="destructive">
            <AlertTitle>Accounts unavailable</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : null}
        <Card>
          <CardHeader>
            <CardTitle>Account inventory</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Username</TableHead>
                  <TableHead>Region</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Quota</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {accounts.map((account) => (
                  <TableRow key={account.id}>
                    <TableCell>{account.username}</TableCell>
                    <TableCell>{account.region}</TableCell>
                    <TableCell>{account.account_type}</TableCell>
                    <TableCell>
                      <Badge
                        className={account.status === "active" ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300" : undefined}
                        variant="secondary"
                      >
                        {account.status}
                      </Badge>
                    </TableCell>
                    <TableCell>{account.quota_remaining}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            {!loading && accounts.length === 0 && !error ? <p className="empty-state">No accounts</p> : null}
          </CardContent>
        </Card>
      </div>
    </main>
  );
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
