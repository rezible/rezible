import { Context, watch } from "runed";
import { createMutation } from "@tanstack/svelte-query";

import { updateOrganizationPreferencesMutation, type AvailableIntegration, type InstalledIntegration, type OrganizationPreferences } from "$lib/api";
import { useUserSessionState } from "$src/lib/user-session.svelte";
import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";

import { StepperController } from "$components/layout/stepper/stepper.svelte";

import OrganizationSetupStep from "./steps/OrganizationSetup.svelte";
import InstallIntegrationsStep from "./steps/InstallIntegrations.svelte";

const chatIntegrations = new Set<string>(["slack_agent"]);
const incidentManagementIntegrations = new Set<string>([]);
const getAvailableNames = (av: AvailableIntegration[], names: Set<string>) => av.filter(a => names.has(a.name));

export type IntegrationCapabilitySuggestion = {
	label: string;
	available: AvailableIntegration[];
};

const makeSuggestedIntegrations = (prefs: OrganizationPreferences, available: AvailableIntegration[], installed: InstalledIntegration[]) => {
	const suggestions: IntegrationCapabilitySuggestion[] = [];
	const installedNames = new Set(installed.map(intg => intg.attributes.integrationName));

	if (installedNames.isDisjointFrom(chatIntegrations)) {
		suggestions.push({
			label: "Chat",
			available: getAvailableNames(available, chatIntegrations),
		});
	};

	if (!prefs.enableIncidentManagement && installedNames.isDisjointFrom(incidentManagementIntegrations)) {
		suggestions.push({
			label: "Incident Management",
			available: getAvailableNames(available, incidentManagementIntegrations),
		});
	};

	return suggestions;
}

export type ConfigureOrganizationOptions = {
	enableIncidentManagement: boolean;
}

export class InitialSetupController {
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

	integrationSuggestions = $derived(makeSuggestedIntegrations(this.orgPrefs, this.integrations.available, this.integrations.installed))

	canContinueIntegrations = $derived(true);
	integrationsContinueButtonText = $derived(
		this.integrationSuggestions.length === 0 ? "Finish setup" : "Skip for now"
	);

	canFinish = $derived(this.canContinueOrg && this.canContinueIntegrations);

	lastCompletedStepIdx = $derived.by(() => {
		// if (this.integrationSuggestions.length === 0) return 1;
		if (!!this.currOrgPrefs) return 1;
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

const ctx = new Context<InitialSetupController>("InitialSetupController");
export const initInitialSetupController = () => ctx.set(new InitialSetupController());
export const useInitialSetupController = () => ctx.get();
