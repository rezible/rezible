import { createMutation } from "@tanstack/svelte-query";
import { toast } from "svelte-sonner";

import { type ApiError, type IntegrationInstallation, issueIntegrationWebhookTokenMutation } from "$lib/api";

// Issues webhook URLs for a provider's installations. Rezible stores only a hash of each URL's token, so an
// issued URL lives in this controller only: it is shown once and gone after leaving the page. The first URL is
// generated at once; replacing one is confirmed first, because the earlier URL stops working.
export class WebhookUrlController {
	// The issued URL is a secret; do not keep it in the mutation cache.
	private issueMut = createMutation(() => ({ ...issueIntegrationWebhookTokenMutation(), gcTime: 0 }));

	private issuedUrls = $state<Record<string, string>>({});
	private issueErrors = $state<Record<string, ApiError>>({});
	// Installations issued a URL on this page, including failed attempts, which may have replaced one.
	private attemptedIds = $state<Record<string, true>>({});

	// The installation whose replacement is waiting for confirmation.
	confirmingId = $state<string>();
	issuingId = $state<string>();

	// A failed generation keeps the confirmation open, so its error is shown there.
	confirmingError = $derived.by(() => {
		if (!this.confirmingId) {
			return undefined;
		}
		return this.issueErrors[this.confirmingId];
	});

	issuedUrl(id: string): string | undefined {
		return this.issuedUrls[id];
	}

	issueError(id: string): ApiError | undefined {
		return this.issueErrors[id];
	}

	hasUrl(installation: IntegrationInstallation): boolean {
		return installation.attributes.webhookUrlIssued || !!this.attemptedIds[installation.id];
	}

	requestGenerate = async (installation: IntegrationInstallation) => {
		const id = installation.id;
		delete this.issueErrors[id];
		if (this.hasUrl(installation)) {
			this.confirmingId = id;
			return;
		}
		await this.issue(id);
	};

	cancelGenerate = () => {
		if (this.issuingId) return;

		this.confirmingId = undefined;
	};

	confirmGenerate = async () => {
		const id = this.confirmingId;
		if (!id) return;

		if (await this.issue(id)) {
			this.confirmingId = undefined;
		}
	};

	private issue = async (id: string): Promise<boolean> => {
		if (this.issuingId) return false;

		this.issuingId = id;
		delete this.issueErrors[id];
		try {
			const resp = await this.issueMut.mutateAsync({ path: { id } });
			this.issuedUrls[id] = resp.data.url;
			return true;
		} catch (e) {
			// The earlier URL may already have been replaced, so a URL shown before is no longer trusted.
			delete this.issuedUrls[id];
			this.issueErrors[id] = e as ApiError;
			return false;
		} finally {
			this.attemptedIds[id] = true;
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
