import type { IntegrationProviderController } from "../controller.svelte";

export const grafanaIntegrationName = "grafana";

type ProviderInstaller = Pick<IntegrationProviderController, "install">;

// The token stays in this form only, and is cleared after installation.
// Replacing the token installs again with the installation's URL, which keeps its settings.
export class GrafanaInstallForm {
	private provider: ProviderInstaller;
	readonly replacing: boolean;

	url = $state("");
	token = $state("");
	urlTouched = $state(false);
	tokenTouched = $state(false);

	urlError = $derived(this.url.trim() ? undefined : "Enter your Grafana URL.");
	tokenError = $derived(this.token.trim() ? undefined : "Enter a service account token.");
	canSubmit = $derived(!this.urlError && !this.tokenError);

	constructor(provider: ProviderInstaller, replaceUrl?: string) {
		this.provider = provider;
		this.replacing = !!replaceUrl;
		this.url = replaceUrl ?? "";
	}

	installConfig() {
		if (!this.canSubmit) {
			return undefined;
		}
		return {
			url: this.url.trim(),
			token: this.token.trim(),
		};
	}

	// Returns whether Grafana was installed.
	install = async () => {
		this.urlTouched = true;
		this.tokenTouched = true;
		const config = this.installConfig();
		if (!config) {
			return false;
		}

		let successMessage: string | undefined;
		if (this.replacing) {
			successMessage = "Replaced the service account token.";
		}
		const installed = await this.provider.install(grafanaIntegrationName, config, successMessage);
		if (installed) {
			this.reset();
		}
		return installed;
	};

	private reset() {
		if (!this.replacing) {
			this.url = "";
		}
		this.token = "";
		this.urlTouched = false;
		this.tokenTouched = false;
	}
}
