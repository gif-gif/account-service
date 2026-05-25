export function apiBaseURL() {
  return import.meta.env.VITE_API_BASE_URL ?? "";
}

export function appEnvironment() {
  return import.meta.env.VITE_APP_ENV ?? import.meta.env.MODE;
}
