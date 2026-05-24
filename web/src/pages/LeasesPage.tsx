import { FormEvent, useEffect, useState } from "react";

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
    <main>
      <h1>Leases</h1>
      <form onSubmit={handleSubmit}>
        <label>
          Lease status
          <input value={status} onChange={(event) => setStatus(event.target.value)} />
        </label>
        <button type="submit">Filter leases</button>
      </form>
      <ul>
        {leases.map((lease) => (
          <li key={lease.lease_id}>
            <span>{lease.lease_id}</span>
            <span>{lease.account_id}</span>
            <span>{lease.status}</span>
          </li>
        ))}
      </ul>
    </main>
  );
}
