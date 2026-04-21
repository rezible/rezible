import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
    input: {path: "../../apps/backend/openapi/v1/openapi.yaml"},
    output: {path: "./src/oapi.gen"},
    plugins: [
        {name: '@hey-api/client-fetch'},
        {name: "@hey-api/typescript", enums: "javascript"},
        {name: "@tanstack/svelte-query"},
    ],
    logs: { file: false },
    parser: {
        transforms: { 
            readWrite: false,
        }
    }
});