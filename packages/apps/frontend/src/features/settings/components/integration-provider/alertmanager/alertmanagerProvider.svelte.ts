import { createMutation } from "@tanstack/svelte-query";
import { toast } from "svelte-sonner";

import { type ErrorModel, issueIntegrationWebhookTokenMutation } from "$lib/api";

import type { IntegrationProviderController } from "../controller.svelte";

export const alertmanagerIntegrationName = "alertmanager";

type ProviderInstaller = Pick<IntegrationProviderController, "install">;

// Installs Alertmanager and issues webhook URLs for its installations. Rezible stores only a hash of each
// URL's token, so an issued URL lives in this controller only: it is shown once and gone after leaving the
// page. The page cannot tell whether a URL was issued before, so every generation is confirmed with a
// warning that any earlier URL stops working.
export class AlertmanagerProviderController {
	private provider: ProviderInstaller;

	// The issued URL is a secret; do not keep it in the mutation cache.
	private issueMut = createMutation(() => ({ ...issueIntegrationWebhookTokenMutation(), gcTime: 0 }));

	private issuedUrls = $state<Record<string, string>>({});
	private issueErrors = $state<Record<string, ErrorModel>>({});

	// The installation whose generation is waiting for confirmation.
	confirmingId = $state<string>();
	issuingId = $state<string>();

	// A failed generation keeps the confirmation open, so its error is shown there.
	confirmingError = $derived.by(() => {
		if (!this.confirmingId) {
			return undefined;
		}
		return this.issueErrors[this.confirmingId];
	});

	constructor(provider: ProviderInstaller) {
		this.provider = provider;
	}

	// Installing asks for nothing.
	install = async () => {
		await this.provider.install(alertmanagerIntegrationName, {});
	};

	issuedUrl(id: string): string | undefined {
		return this.issuedUrls[id];
	}

	issueError(id: string): ErrorModel | undefined {
		return this.issueErrors[id];
	}

	requestGenerate = (id: string) => {
		this.confirmingId = id;
		delete this.issueErrors[id];
	};

	cancelGenerate = () => {
		if (this.issuingId) return;

		this.confirmingId = undefined;
	};

	confirmGenerate = async () => {
		const id = this.confirmingId;
		if (!id || this.issuingId) return;

		this.issuingId = id;
		delete this.issueErrors[id];
		try {
			const resp = await this.issueMut.mutateAsync({ path: { id } });
			this.issuedUrls[id] = resp.data.url;
			this.confirmingId = undefined;
		} catch (e) {
			// The earlier URL may already have been replaced, so a URL shown before is no longer trusted.
			delete this.issuedUrls[id];
			this.issueErrors[id] = e as ErrorModel;
		} finally {
			this.issuingId = undefined;
		}
	};

	copy = async (text: string, label: string) => {
		try {
			await navigator.clipboard.writeText(text);
			toast.success(`${label} copied.`);
		} catch {
			toast.error(`Unable to copy the ${label.toLowerCase()}.`);
		}
	};
}
