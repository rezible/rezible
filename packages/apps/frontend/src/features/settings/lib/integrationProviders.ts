import type { Component } from "svelte";

import type { InstallableIntegration } from "$lib/api";

import RiFlaskLine from "remixicon-svelte/icons/flask-line";
import RiGithubFill from "remixicon-svelte/icons/github-fill";
import RiGoogleFill from "remixicon-svelte/icons/google-fill";
import RiPlugLine from "remixicon-svelte/icons/plug-line";
import RiSlackFill from "remixicon-svelte/icons/slack-fill";

export type IntegrationProviderInfo = {
	name: string;
	displayName: string;
	description: string;
	icon: Component;
	// Plural noun for this provider's connections, such as "workspaces".
	connectionNoun: string;
};

const knownProviders: Record<string, Omit<IntegrationProviderInfo, "name">> = {
	github: {
		displayName: "GitHub",
		description: "Follow repositories, pushes, and pull requests from your GitHub accounts.",
		icon: RiGithubFill,
		connectionNoun: "accounts",
	},
	slack: {
		displayName: "Slack",
		description: "Chat with the Rezible agent and run incidents in dedicated Slack channels.",
		icon: RiSlackFill,
		connectionNoun: "workspaces",
	},
	google: {
		displayName: "Google Workspace",
		description: "Create Google Meet video conferences for incidents.",
		icon: RiGoogleFill,
		connectionNoun: "workspaces",
	},
	demo: {
		displayName: "Demo",
		description: "Sample data for exploring Rezible.",
		icon: RiFlaskLine,
		connectionNoun: "connections",
	},
};

export const getIntegrationProviderInfo = (
	name: string,
	installable: InstallableIntegration[] = []
): IntegrationProviderInfo => {
	const known = knownProviders[name];
	if (known) {
		return { name, ...known };
	}

	const firstIntegration = installable.at(0);
	return {
		name,
		displayName: firstIntegration?.displayName ?? name,
		description: firstIntegration?.description ?? "",
		icon: RiPlugLine,
		connectionNoun: "connections",
	};
};
