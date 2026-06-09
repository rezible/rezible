import { Context, watch } from "runed";
import { createMutation, createQuery } from "@tanstack/svelte-query";

import { finishOrganizationSetupMutation, getOrganizationOptions, updateOrganizationMutation, type AvailableIntegration, type InstalledIntegration, type Organization, type OrganizationPreferences } from "$lib/api";
import { useAuthSessionState } from "$lib/auth-session.svelte";
import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";

type OrgDetails = {
	name: string;
}

const makeOrgDetails = (org?: Organization): OrgDetails => {
	if (!org) return { name: "" };
	return { name: org.attributes.name };
};

const RequiredCapabilities = new Set(["chat", "users"]);

const mapEnabledCapabilityNames = (intg: InstalledIntegration) => {
	return Object.entries(intg.attributes.capabilities)
		.filter(([_, enabled]) => enabled)
		.map(([name, _]) => name)
}

export class InitialSetupViewController {
	private session = useAuthSessionState();
	private integrations = useIntegrationsController();

	private orgId = $derived(this.session.org?.id);

	private orgDetailsQuery = createQuery(() => ({
		...getOrganizationOptions({ path: { id: this.orgId || "" } }),
		enabled: !!this.orgId,
	}));
	private orgDetailsQueryData = $derived(this.orgDetailsQuery.data?.data);

	orgDetails = $state(makeOrgDetails());
	private updateOrgMut = createMutation(() => ({
		...updateOrganizationMutation(),
		onSuccess: () => {
			this.orgDetailsQuery.refetch();
		}
	}));
	canContinueOrgDetails = $derived(this.orgDetailsQuery.isSuccess);

	async onOrgDetailsNext() {
		if (!this.orgId) return;
		// check anything changed
		const currAttrs = this.orgDetailsQueryData?.attributes;
		if (!!currAttrs) {
			if (currAttrs.name === this.orgDetails.name) return;
		}

		await this.updateOrgMut.mutateAsync({
			path: { id: this.orgId },
			body: {
				attributes: {
					name: this.orgDetails.name,
				}
			}
		})
	};

	canContinueOrgPrefs = $derived(!!this.orgId);

	orgPrefs = $state<OrganizationPreferences>({ enableIncidentManagement: false });

	async onOrganizationPreferencesNext() {
		if (!this.orgId) return;
		// check anything changed
		const currPrefs = this.orgDetailsQueryData?.attributes.preferences;
		if (!!currPrefs) {
			if (currPrefs.enableIncidentManagement === this.orgPrefs.enableIncidentManagement) return;
		}

		await this.updateOrgMut.mutateAsync({
			path: { id: this.orgId },
			body: {
				attributes: {
					preferences: this.orgPrefs,
				}
			}
		})
	}

	remainingRequiredCapabilities = $derived.by(() => {
		const enabledInstalled = new Set(this.integrations.installed.flatMap(mapEnabledCapabilityNames));
		return RequiredCapabilities.difference(enabledInstalled).values().toArray()
	});
	availableOptions = $derived(this.integrations.available.filter((intg) => !this.integrations.installationsByName.has(intg.name)));
	availableIntegrationsForCapabilities = $derived.by(() => {
		const capMap = new Map<string, AvailableIntegration[]>();
		this.availableOptions.forEach((intg) => {
			intg.supportedCapabilities.forEach((cap) => {
				capMap.set(cap, [...(capMap.get(cap) || []), intg]);
			});
		});
		return capMap;
	});

	canContinueCapabilities = $derived(this.remainingRequiredCapabilities.length === 0);

	canFinish = $derived(this.canContinueOrgDetails && this.canContinueOrgPrefs && this.canContinueCapabilities);

	private finishOrgSetupMut = createMutation(() => finishOrganizationSetupMutation());
	private finishing = $state(false);
	async doFinishOrganizationSetup() {
		if (this.finishOrgSetupMut.isPending) return;
		if (!this.orgId) return;
		this.finishing = true;
		try {
			await this.finishOrgSetupMut.mutateAsync({ path: { id: this.orgId } });
			this.session.refetch();
		} catch (e) {
			console.error("failed to finish setup", e);
			this.finishing = false;
			throw e;
		}
	}

	constructor() {
		watch(() => this.orgDetailsQueryData, org => {
			if (!org) return;
			this.orgDetails = makeOrgDetails(org);
			this.orgPrefs = org.attributes.preferences;
		});
	}


	loading = $derived(this.finishing || this.integrations.loading);
}

const ctx = new Context<InitialSetupViewController>("InitialSetupViewController");
export const initInitialSetupViewController = () => ctx.set(new InitialSetupViewController());
export const useInitialSetupViewController = () => ctx.get();
