import { useEffect, useState } from "react";

import { apiFetch } from "../lib/api";

type AuditLog = {
  id: string;
  actor_type: string;
  action: string;
  request_id: string;
  metadata: Record<string, unknown>;
};

export function AuditLogsPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);

  useEffect(() => {
    void apiFetch<{ audit_logs: AuditLog[] }>("/api/v1/audit-logs").then((response) => setLogs(response.audit_logs));
  }, []);

  return (
    <main>
      <h1>Audit logs</h1>
      <ul>
        {logs.map((log) => (
          <li key={log.id}>
            <span>{log.actor_type}</span>
            <span>{log.action}</span>
            <span>{log.request_id}</span>
            <code>{Object.values(log.metadata).join(" ")}</code>
          </li>
        ))}
      </ul>
    </main>
  );
}
