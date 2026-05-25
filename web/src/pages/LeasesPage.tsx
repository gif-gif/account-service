import { FormEvent, useEffect, useState } from "react";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../components/ui/table";
import { apiFetch } from "../lib/api";

type Lease = {
  lease_id: string;
  account_id: string;
  status: string;
};

export function LeasesPage() {
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
          <h1 className="page-title">Leases</h1>
          <p className="page-description">Inspect active and historical account lease assignments.</p>
        </div>
      </div>
      <div className="content-stack">
        <Card>
          <CardHeader>
            <CardTitle>Lease filters</CardTitle>
            <CardDescription>Filter by active, released, or expired status.</CardDescription>
          </CardHeader>
          <CardContent>
            <form className="filter-grid" onSubmit={handleSubmit}>
              <Label className="form-row">
                Lease status
                <Input value={status} onChange={(event) => setStatus(event.target.value)} />
              </Label>
              <Button type="submit">Filter leases</Button>
            </form>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Lease activity</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Lease ID</TableHead>
                  <TableHead>Account ID</TableHead>
                  <TableHead>Status</TableHead>
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
            {leases.length === 0 ? <p className="empty-state">No leases</p> : null}
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
