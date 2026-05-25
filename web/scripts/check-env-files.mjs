import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const environments = [
  { appEnv: "local", file: ".env.localdev", mode: "localdev" },
  { appEnv: "development", file: ".env.development", mode: "development" },
  { appEnv: "test", file: ".env.test", mode: "test" },
  { appEnv: "production", file: ".env.production", mode: "production" },
];

const readme = readFileSync(resolve("README.md"), "utf8");

for (const { appEnv, file, mode } of environments) {
  const path = resolve(file);
  const contents = readFileSync(path, "utf8");
  const env = parseEnv(contents);

  if (env.VITE_APP_ENV !== appEnv) {
    throw new Error(`${file} must define VITE_APP_ENV=${appEnv}`);
  }
  if (!env.VITE_API_BASE_URL) {
    throw new Error(`${file} must define VITE_API_BASE_URL`);
  }
  if (!readme.includes(`\`${mode}\``) || !readme.includes(`\`${file}\``) || !readme.includes(env.VITE_API_BASE_URL)) {
    throw new Error(`README.md must document ${file} with mode ${mode} and API base URL ${env.VITE_API_BASE_URL}`);
  }
}

console.log("environment files are valid");

function parseEnv(contents) {
  return Object.fromEntries(
    contents
      .split(/\r?\n/)
      .map((line) => line.trim())
      .filter((line) => line && !line.startsWith("#"))
      .map((line) => {
        const separator = line.indexOf("=");
        return [line.slice(0, separator), line.slice(separator + 1)];
      }),
  );
}
