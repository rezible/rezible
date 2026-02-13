import { defineConfig, type UserConfig } from "@hey-api/openapi-ts";

export const createConfig = (input: UserConfig["input"], output: UserConfig["output"]) => {
	return defineConfig({
		input,
		output,
		plugins: [
			{name: "@hey-api/client-fetch"},
			{name: "@hey-api/typescript", enums: "javascript"},
			{name: '@hey-api/sdk', asClass: true},
		],
		logs: { file: false },
		parser: {
			filters: {
				tags: { include: ["documents"] },
			},
			transforms: {
				readWrite: false,
			}
		}
	});
}
