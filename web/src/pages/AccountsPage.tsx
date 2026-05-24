import { FormEvent, useEffect } from "react";

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

  return (
    <main>
      <h1>Accounts</h1>
      <form onSubmit={handleSubmit}>
        <label>
          Region
          <input value={filters.region} onChange={(event) => setFilter("region", event.target.value)} />
        </label>
        <label>
          Account type
          <input value={filters.accountType} onChange={(event) => setFilter("accountType", event.target.value)} />
        </label>
        <label>
          Status
          <input value={filters.status} onChange={(event) => setFilter("status", event.target.value)} />
        </label>
        <label>
          Tags
          <input value={filters.tags} onChange={(event) => setFilter("tags", event.target.value)} />
        </label>
        <button type="submit">Apply filters</button>
      </form>
      {loading ? <p>Loading accounts</p> : null}
      {error ? <p role="alert">{error}</p> : null}
      <table>
        <thead>
          <tr>
            <th>Username</th>
            <th>Region</th>
            <th>Type</th>
            <th>Status</th>
            <th>Quota</th>
          </tr>
        </thead>
        <tbody>
          {accounts.map((account) => (
            <tr key={account.id}>
              <td>{account.username}</td>
              <td>{account.region}</td>
              <td>{account.account_type}</td>
              <td>{account.status}</td>
              <td>{account.quota_remaining}</td>
            </tr>
          ))}
        </tbody>
      </table>
      {!loading && accounts.length === 0 && !error ? <p>No accounts</p> : null}
    </main>
  );
}
