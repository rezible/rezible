import { Context, watch } from "runed";
import { createMutation } from "@tanstack/svelte-query";

import { finishOrganizationSetupMutation, type AvailableIntegration, type InstalledIntegration, type Organization } from "$lib/api";
import { useAuthSessionState } from "$lib/auth-session.svelte";
import { useIntegrationsController } from "$features/settings/lib/integrationsController.svelte";

type OrgDetails = {
	name: string;
}

const makeOrgDetails = (org?: Organization): OrgDetails => {
	if (!org) return { name: "" };
	return { name: org.attributes.name };
};

type WorkspacePreferences = {
	enableIncidentManagement: boolean;
}

const makeWorkspacePreferences = (): WorkspacePreferences => {
	return { enableIncidentManagement: true };
}

const RequiredCapabilities = new Set(["chat", "users"]);

const mapEnabledCapabilityNames = (intg: InstalledIntegration) => {
	return Object.entries(intg.attributes.capabilities)
		.filter(([_, enabled]) => enabled)
		.map(([name, _]) => name)
}

const settingsEqual = (a: any, b: any) => {
	return JSON.stringify(a) === JSON.stringify(b)
}

export class InitialSetupViewController {
	private session = useAuthSessionState();
	private integrations = useIntegrationsController();

	constructor() {
		watch(() => this.session.org, org => {
			if (org) this.orgDetails = makeOrgDetails(org);
		})
	}

	private sessionOrgDetails = $derived(makeOrgDetails(this.session.org));
	orgDetails = $state(makeOrgDetails());

	canContinueOrgDetails = $derived(true);

	async onOrgDetailsNext() {
		if (settingsEqual(this.sessionOrgDetails, this.orgDetails)) {
			return;
		}
		console.log("update org details");
		return new Promise<void>((res, rej) => {
			setTimeout(() => {res()}, 1000);
		});
	};

	private sessionWorkspacePrefs = $derived(makeWorkspacePreferences());
	workspacePrefs = $state(makeWorkspacePreferences());

	canContinueWorkspacePrefs = $derived(true);

	async onWorkspacePreferencesNext() {
		if (settingsEqual(this.sessionWorkspacePrefs, this.workspacePrefs)) {
			return;
		}
		console.log("update preferences");
		return new Promise<void>((res, rej) => {
			setTimeout(() => {res()}, 1000);
		});
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

	canFinish = $derived(this.canContinueOrgDetails && this.canContinueWorkspacePrefs && this.canContinueCapabilities);

	private finishOrgSetupMut = createMutation(() => finishOrganizationSetupMutation());
	private finishing = $state(false);
	async doFinishOrganizationSetup() {
		if (this.finishOrgSetupMut.isPending) return;
		const id = this.session.org?.id;
		if (!id) return;
		this.finishing = true;
		try {
			await this.finishOrgSetupMut.mutateAsync({ path: { id } });
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
