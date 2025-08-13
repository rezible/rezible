import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
	input: {
		path: "/tmp/rezible-spec.yaml",
	},
	output: {
		path: "src/lib/api/oapi.gen",
	},
	plugins: [
		{
			name: '@hey-api/client-fetch',
		},
		{
			name: "@hey-api/typescript",
			enums: "javascript",
		},
		{
			name: "@tanstack/svelte-query",
		},
	],
	logs: {
		file: false,
	},
	parser: {
		transforms: {
			readWrite: false,
		}
	}
});
