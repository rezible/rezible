import type { InstallableIntegration, IntegrationInstallation } from "$lib/api";

const capabilityLabels: Record<string, string> = {
	chat_context: "Chat",
	incident_management: "Incident channels",
	video_conferencing: "Video calls",
	code_changes: "Code changes",
	repositories: "Repositories",
};

const maxCapabilityPills = 3;

// Short names for providers that offer several apps, such as Slack.
const appLabels: Record<string, string> = {
	slack_agent: "Agent",
	slack_incidents: "Incidents",
};

export type CatalogueProviderInput = {
	displayName: string;
	// Plural noun for this provider's connections, such as "workspaces".
	connectionNoun: string;
	installable: InstallableIntegration[];
	installations: IntegrationInstallation[];
};

// User-facing capability labels. Internal capabilities, such as event_sync, have no label and are left out.
export const capabilityPills = (provider: CatalogueProviderInput): string[] => {
	const capabilities = [
		...provider.installable.flatMap((intg) => intg.capabilities),
		...provider.installations.flatMap((installation) => installation.attributes.capabilities),
	];

	const labels = new Set<string>();
	for (const capability of capabilities) {
		const label = capabilityLabels[capability];
		if (label) {
			labels.add(label);
		}
	}
	return Array.from(labels).slice(0, maxCapabilityPills);
};

// Describes what is connected, or undefined when nothing is.
// Connections named after the app itself (such as "Google Workspace") say nothing new, so only
// names that identify an account or workspace are shown.
export const connectionStatus = (provider: CatalogueProviderInput): string | undefined => {
	if (provider.installations.length === 0) {
		return undefined;
	}

	const appNames = new Set([provider.displayName, ...provider.installable.map((intg) => intg.displayName)]);
	// Several apps can connect the same workspace, so names are counted once.
	const connectionNames = new Set<string>();
	for (const installation of provider.installations) {
		const name = installation.attributes.displayName;
		if (!appNames.has(name)) {
			connectionNames.add(name);
		}
	}

	if (connectionNames.size > 1) {
		return `${connectionNames.size} ${provider.connectionNoun}`;
	}

	const [connectionName] = connectionNames;
	const label = connectionName ?? "Connected";

	const definitionNames = new Set([
		...provider.installable.map((intg) => intg.name),
		...provider.installations.map((installation) => installation.attributes.name),
	]);
	if (definitionNames.size < 2) {
		return label;
	}

	const installedApps = new Set(provider.installations.map((installation) => installation.attributes.name));
	const appList = Array.from(installedApps, (name) => appLabels[name] ?? name).toSorted();
	return `${label} · ${appList.join(", ")}`;
};
