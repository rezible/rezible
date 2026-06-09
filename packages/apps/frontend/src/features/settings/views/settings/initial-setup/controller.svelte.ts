import { Context, watch } from "runed";
import { createMutation, createQuery } from "@tanstack/svelte-query";

import { finishOrganizationSetupMutation, getOrganizationOptions, getOrganizationPreferencesOptions, updateOrganizationDetailsMutation, updateOrganizationPreferencesMutation, type AvailableIntegration, type InstalledIntegration, type Organization, type OrganizationPreferences, type UpdateOrganizationPreferencesRequestAttributes } from "$lib/api";
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
		...getOrganizationOptions({ path: { id: this.orgId || "" }}),
		enabled: !!this.orgId,
	}));
	private orgDetailsQueryData = $derived(this.orgDetailsQuery.data?.data);
	
	orgDetails = $state(makeOrgDetails());
	private updateOrgDetailsMut = createMutation(() => ({
		...updateOrganizationDetailsMutation(),
		onSuccess: () => {
			this.orgDetailsQuery.refetch();
		}
	}));
	canContinueOrgDetails = $derived(this.orgDetailsQuery.isSuccess);

	async onOrgDetailsNext() {
		if (!this.orgId) return;
		// check anything changed
		if (this.orgDetailsQueryData?.attributes.name === this.orgDetails.name) return;
		
		await this.updateOrgDetailsMut.mutateAsync({
			path: { id: this.orgId },
			body: { 
				attributes: {
					name: this.orgDetails.name,
				}
			}
		})
	};

	private orgPrefsQuery = createQuery(() => ({ 
		...getOrganizationPreferencesOptions({ path: { id: this.orgId || "" }}),
		enabled: !!this.orgId && this.canContinueOrgDetails,
	}));
	private orgPrefsQueryData = $derived(this.orgPrefsQuery.data?.data);

	canContinueOrgPrefs = $derived(!!this.orgId && this.orgPrefsQuery.isSuccess);

	orgPrefs = $state<OrganizationPreferences>({ enableIncidentManagement: false });
	private updateOrgPrefsMut = createMutation(() => ({
		...updateOrganizationPreferencesMutation(),
		onSuccess: () => {
			this.orgPrefsQuery.refetch();
		}
	}));

	async onOrganizationPreferencesNext() {
		if (!this.orgId) return;
		// check anything changed
		if (this.orgPrefsQueryData?.enableIncidentManagement === this.orgPrefs.enableIncidentManagement) return;

		await this.updateOrgPrefsMut.mutateAsync({
			path: {id: this.orgId},
			body: { 
				attributes: {
					enableIncidentManagement: this.orgPrefs.enableIncidentManagement,
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
			if (org) this.orgDetails = makeOrgDetails(org);
		});
		watch(() => this.orgPrefsQueryData, prefs => {
			if (prefs) this.orgPrefs = $state.snapshot(prefs);
		});
	}


	loading = $derived(this.finishing || this.integrations.loading);
}

const ctx = new Context<InitialSetupViewController>("InitialSetupViewController");
export const initInitialSetupViewController = () => ctx.set(new InitialSetupViewController());
export const useInitialSetupViewController = () => ctx.get();
