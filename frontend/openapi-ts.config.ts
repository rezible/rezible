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
	plugins: [
		{
			name: "@hey-api/typescript",
			enums: "javascript",
		},
		{
			name: "@tanstack/svelte-query",
		},
	],
});
