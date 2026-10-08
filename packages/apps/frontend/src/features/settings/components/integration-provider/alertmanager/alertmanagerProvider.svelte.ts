import type { IntegrationProviderController } from "../controller.svelte";

export const alertmanagerIntegrationName = "alertmanager";

type ProviderInstaller = Pick<IntegrationProviderController, "install">;

// Installs Alertmanager. Its webhook URLs are issued by the shared WebhookUrlController.
export class AlertmanagerProviderController {
	private provider: ProviderInstaller;

	constructor(provider: ProviderInstaller) {
		this.provider = provider;
	}

	// Installing asks for nothing.
	install = async () => {
		await this.provider.install(alertmanagerIntegrationName, {});
	};
}
