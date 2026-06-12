import { Context, watch } from "runed";
import { createMutation, createQuery } from "@tanstack/svelte-query";

import { updateOrganizationPreferencesMutation, type AvailableIntegration, type OrganizationAttributes, type OrganizationPreferences, type UpdateOrganizationPreferencesRequestAttributes } from "$lib/api";
import { useUserSessionState } from "$src/lib/user-session.svelte";
import { getEnabledCapabilties, useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";

import { StepperController } from "$components/layout/stepper/stepper.svelte";

import OrganizationSetupStep from "./steps/OrganizationSetup.svelte";
import InstallIntegrationsStep from "./steps/InstallIntegrations.svelte";

const RequiredCapabilities = new Set(["chat", "users"]);

export type ConfigureOrganizationOptions = {
	enableIncidentManagement: boolean;
}

export class InitialSetupViewController {
	private session = useUserSessionState();
	private integrations = useIntegrationsController();

	orgName = $derived(this.session.org?.attributes.name);
	private orgId = $derived(this.session.org?.id);

	private currOrgPrefs = $derived(this.session.org?.attributes.preferences);
	orgPrefs = $state<ConfigureOrganizationOptions>({enableIncidentManagement: false});
	
	constructor() {
		watch(() => this.currOrgPrefs, prefs => {
			this.orgPrefs = {
				enableIncidentManagement: !!prefs?.enableIncidentManagement,
			}
		});
	}

	private updateOrgPrefsMut = createMutation(() => ({
		...updateOrganizationPreferencesMutation(),
		onSuccess: () => {
			this.session.refetch();
		}
	}));

	orgPrefsValid = $state(false);
	canContinueOrg = $derived(!!this.orgId && this.orgPrefsValid);

	async onOrgNext() {
		const id = $state.snapshot(this.orgId);
		if (!id || !this.orgPrefsValid) return;
		
		// check if anything changed
		if (!!this.currOrgPrefs) {
			if (!!this.currOrgPrefs.enableIncidentManagement === !!this.orgPrefs.enableIncidentManagement) return;
		}

		await this.updateOrgPrefsMut.mutateAsync({
			path: { id },
			body: {
				attributes: {
					enableIncidentManagement: this.orgPrefs.enableIncidentManagement,
				}
			}
		})
	};

	private installedEnabledCapabilities = $derived(new Set(getEnabledCapabilties(this.integrations.installed)));
	remainingRequiredCapabilities = $derived(RequiredCapabilities.difference(this.installedEnabledCapabilities).values().toArray());
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

	canContinueIntegrations = $derived(true);
	integrationsContinueButtonText = $derived(
		this.remainingRequiredCapabilities.length === 0 ? "Finish setup" : "Skip for now"
	);

	canFinish = $derived(this.canContinueOrg && this.canContinueIntegrations);

	lastCompletedStepIdx = $derived.by(() => {

		return 0;
	})

	stepper = new StepperController({
		initialStepIndex: () => this.lastCompletedStepIdx,
		steps: [
			{
				label: "Organization",
				description: "Organization details and preferences",
				component: OrganizationSetupStep,
				onNext: () => this.onOrgNext(),
				canContinue: () => this.canContinueOrg,
			},
			{
				label: "Integrations",
				description: "Install recommended integrations",
				component: InstallIntegrationsStep,
				canContinue: () => this.canContinueIntegrations,
			},
		],
		onFinish: () => this.doFinishSetup(),
	});

	private finishing = $state(false);
	async doFinishSetup() {
		if (this.updateOrgPrefsMut.isPending) return;
		const id = $state.snapshot(this.orgId);
		if (!id) return;
		this.finishing = true;
		try {
			await this.updateOrgPrefsMut.mutateAsync({ 
				path: { id },
				body: { attributes: { initialSetupComplete: true }},
			});
			this.session.refetch();
		} catch (e) {
			console.error("failed to finish setup", e);
			this.finishing = false;
			throw e;
		}
	}

	loading = $derived(this.finishing || this.integrations.loading);
}

const ctx = new Context<InitialSetupViewController>("InitialSetupViewController");
export const initInitialSetupViewController = () => ctx.set(new InitialSetupViewController());
export const useInitialSetupViewController = () => ctx.get();
