import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
	input: '/tmp/rezible-spec.yaml',
	output: {
		path: 'src/lib/api/oapi.gen',
	},
	client: '@hey-api/client-fetch',
	types: {
		enums: 'javascript',
		export: true,
		dates: false,
	},
	services: {
		asClass: false, 
	},
	schemas: false,
	plugins: ['@tanstack/svelte-query'], 
});