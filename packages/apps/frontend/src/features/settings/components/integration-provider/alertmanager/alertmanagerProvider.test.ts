import { describe, expect, mock, test } from "bun:test";

import { AlertmanagerProviderController } from "./alertmanagerProvider.svelte.ts";

describe("AlertmanagerProviderController", () => {
	test("installing sends no configuration", async () => {
		const install = mock(async (_name: string, _config: Record<string, unknown>) => true);
		const controller = new AlertmanagerProviderController({ install });

		await controller.install();

		expect(install).toHaveBeenCalledWith("alertmanager", {});
	});
});
