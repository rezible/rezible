import { defineConfig, type UserConfig } from "@hey-api/openapi-ts";

export const createConfig = (input: UserConfig["input"], output: UserConfig["output"]) => {
	return defineConfig({
		input, output,
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
	})
};
