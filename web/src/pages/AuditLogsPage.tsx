import { useEffect, useState } from "react";

import { Badge } from "../components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "../components/ui/table";
import { apiFetch } from "../lib/api";
import { useI18n } from "../store/settings";

type AuditLog = {
  id: string;
  actor_type: string;
  action: string;
  request_id: string;
  metadata: Record<string, unknown>;
};

export function AuditLogsPage() {
  const { t } = useI18n();
  const [logs, setLogs] = useState<AuditLog[]>([]);

  useEffect(() => {
    void apiFetch<{ audit_logs: AuditLog[] }>("/api/v1/audit-logs").then((response) => setLogs(response.audit_logs));
  }, []);

  return (
    <main className="page">
      <div className="page-header">
        <div>
          <h1 className="page-title">{t("audit.title")}</h1>
          <p className="page-description">{t("audit.description")}</p>
        </div>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>{t("audit.cardTitle")}</CardTitle>
          <CardDescription>{t("audit.cardDescription")}</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("audit.actor")}</TableHead>
                <TableHead>{t("audit.action")}</TableHead>
                <TableHead>{t("audit.requestId")}</TableHead>
                <TableHead>{t("audit.metadata")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {logs.map((log) => (
                <TableRow key={log.id}>
                  <TableCell>
                    <Badge variant="secondary">{log.actor_type}</Badge>
                  </TableCell>
                  <TableCell>{log.action}</TableCell>
                  <TableCell>{log.request_id}</TableCell>
                  <TableCell>
                    <code className="metadata-code">{Object.values(log.metadata).join(" ")}</code>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          {logs.length === 0 ? <p className="empty-state">{t("audit.empty")}</p> : null}
        </CardContent>
      </Card>
    </main>
  );
}
