import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const requiredFiles = [
  ".env.localdev",
  ".env.development",
  ".env.test",
  ".env.production",
];

for (const file of requiredFiles) {
  const path = resolve(file);
  const contents = readFileSync(path, "utf8");

  if (!/^VITE_APP_ENV=.+$/m.test(contents)) {
    throw new Error(`${file} must define VITE_APP_ENV`);
  }
  if (!/^VITE_API_BASE_URL=.+$/m.test(contents)) {
    throw new Error(`${file} must define VITE_API_BASE_URL`);
  }
}

console.log("environment files are valid");
