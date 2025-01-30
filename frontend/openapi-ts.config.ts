import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: {
    path: "/tmp/rezible-spec.yaml",
  },
  output: {
    path: "src/lib/api/oapi.gen",
  },
  client: {
    name: "@hey-api/client-fetch",
  },
  // types: {
  // 	enums: 'javascript',
  // 	export: true,
  // 	dates: false,
  // },
  // services: {
  // 	asClass: false,
  // },
  // schemas: false,
  plugins: [
    {
      name: "@hey-api/typescript",
      enums: "javascript",
    },
    {
      name: "@hey-api/transformers",
      dates: true,
    },
    {
      name: "@tanstack/svelte-query",
    },
  ],
});
