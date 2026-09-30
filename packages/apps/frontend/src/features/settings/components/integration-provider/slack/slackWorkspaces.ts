import type { IntegrationInstallation } from "$lib/api";

export const slackAgentName = "slack_agent";
export const slackIncidentsName = "slack_incidents";

// Both Slack apps identify a connection by its workspace, so their installations can be shown together.
export type SlackWorkspace = {
	ref: string;
	displayName: string;
	agent?: IntegrationInstallation;
	incidents?: IntegrationInstallation;
};

export const groupSlackWorkspaces = (
	agentInstallations: IntegrationInstallation[],
	incidentInstallations: IntegrationInstallation[]
): SlackWorkspace[] => {
	const workspaces = new Map<string, SlackWorkspace>();

	const workspaceFor = (installation: IntegrationInstallation) => {
		const ref = installation.attributes.providerInstallationRef;
		let workspace = workspaces.get(ref);
		if (!workspace) {
			workspace = { ref, displayName: installation.attributes.displayName };
			workspaces.set(ref, workspace);
		}
		return workspace;
	};

	for (const installation of agentInstallations) {
		workspaceFor(installation).agent = installation;
	}
	for (const installation of incidentInstallations) {
		workspaceFor(installation).incidents = installation;
	}

	return Array.from(workspaces.values()).toSorted((a, b) => a.displayName.localeCompare(b.displayName));
};
