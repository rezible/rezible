import { createMutation } from "@tanstack/svelte-query";

import { checkIntegrationHealthMutation, type ErrorModel, type IntegrationInstallation } from "$lib/api";

export type GrafanaHealthResult =
	{ kind: "ok" } | { kind: "failed"; message: string } | { kind: "error"; error: ErrorModel };

type CheckedResult = {
	// The settings the check ran with. A result for other settings no longer describes the connection.
	settings: string;
	result: GrafanaHealthResult;
};

const settingsSnapshot = (installation: IntegrationInstallation) => {
	return JSON.stringify(installation.attributes.userSettings ?? {});
};

// Checks whether Rezible can read each configured Grafana data source with the saved settings and token.
// Grafana allows one installation, so there is one result.
export class GrafanaHealthCheck {
	private checkMut = createMutation(() => checkIntegrationHealthMutation());
	private checked = $state.raw<CheckedResult>();
	// Counts clears, so a check started before one does not put its result back.
	private generation = 0;

	checking = $state(false);

	resultFor(installation: IntegrationInstallation): GrafanaHealthResult | undefined {
		if (!this.checked || this.checked.settings !== settingsSnapshot(installation)) {
			return undefined;
		}
		return this.checked.result;
	}

	// The result no longer describes the connection, such as after installing or replacing the token.
	clear = () => {
		this.generation += 1;
		this.checked = undefined;
		this.checking = false;
	};

	check = async (installation: IntegrationInstallation) => {
		if (this.checking) return;

		const generation = this.generation;
		const settings = settingsSnapshot(installation);
		this.checking = true;
		this.checked = undefined;

		let result: GrafanaHealthResult;
		try {
			const resp = await this.checkMut.mutateAsync({ path: { id: installation.id } });
			const health = resp.data;
			if (health.ok) {
				result = { kind: "ok" };
			} else {
				result = { kind: "failed", message: health.error ?? "The check failed without a reason." };
			}
		} catch (e) {
			result = { kind: "error", error: e as ErrorModel };
		}

		if (generation !== this.generation) return;

		this.checked = { settings, result };
		this.checking = false;
	};
}
