import { FormEvent, useState } from "react";

import type { AuthStore } from "../store/auth";
import { useAuthStore } from "../store/auth";

type Props = {
  store?: AuthStore;
};

export function LoginPage({ store = useAuthStore }: Props) {
  const login = store((state) => state.login);
  const loading = store((state) => state.loading);
  const error = store((state) => state.error);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await login(username, password);
  }

  return (
    <main>
      <form aria-label="Admin login" onSubmit={handleSubmit}>
        <h1>Account Admin</h1>
        <label>
          Username
          <input autoComplete="username" value={username} onChange={(event) => setUsername(event.target.value)} />
        </label>
        <label>
          Password
          <input
            autoComplete="current-password"
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
          />
        </label>
        {error ? <p role="alert">{error}</p> : null}
        <button disabled={loading} type="submit">
          Sign in
        </button>
      </form>
    </main>
  );
}
