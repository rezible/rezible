import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
	input: {path: "/tmp/rezible-spec.yaml"},
	output: {
        path: "src/lib/api/oapi.gen",
        indexFile: false,
    },
	plugins: [
		{name: "@hey-api/client-fetch"},
		{
			name: "@hey-api/typescript",
			enums: "javascript",
		},
		{ 
			name: '@hey-api/sdk',
			asClass: true,
		},
	],
	logs: {
		file: false,
	},
	parser: {
        filters: {
            tags: {
                include: ["documents"],
            },
        },
		transforms: {
			readWrite: false,
		}
	}
});
