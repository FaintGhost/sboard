import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../openapi.yaml",
  output: "src/lib/api/gen",
  plugins: [
    "@hey-api/typescript",
    {
      name: "@hey-api/sdk",
    },
    "@hey-api/client-fetch",
    {
      name: "zod",
    },
  ],
});
