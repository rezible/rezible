import { createMutation } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import { completeOrgSetupMutation, type ErrorModel } from "$lib/api";
import { useUserSessionState } from "$lib/user-session.svelte";

export class InitialSetupController {
	private session = useUserSessionState();
	private orgId = $derived(this.session.org?.id);

	name = $state("");
	timezone = $state("");

	private setupMutation = createMutation(() => completeOrgSetupMutation());
	error = $derived(this.setupMutation.error as ErrorModel | null);
	loading = $derived(!this.session.ready && !this.session.error);
	saving = $derived(this.setupMutation.isPending);

	constructor() {
		watch(
			() => this.session.org,
			(org) => {
				this.name = org?.attributes.name ?? "";
				this.timezone = org?.attributes.preferences?.timezone ?? "";
			}
		);
	}

	isAdmin = $derived(this.session.isAdmin);
	canSubmit = $derived(this.isAdmin && this.name.trim().length > 0 && !this.loading && !this.saving);

	finish = async () => {
		if (!this.canSubmit || !this.orgId) return;
		await this.setupMutation.mutateAsync({
			path: { id: this.orgId },
			body: { attributes: { name: this.name.trim(), timezone: this.timezone.trim() } },
		});
		this.session.refetch();
	};
}

const ctx = new Context<InitialSetupController>("InitialSetupController");
export const initInitialSetupController = () => ctx.set(new InitialSetupController());
export const useInitialSetupController = () => ctx.get();
