import { describe, expect, test } from "bun:test";

import type { InstallableIntegration, IntegrationInstallation } from "$lib/api";

import { capabilityPills, connectionStatus, type CatalogueProviderInput } from "../integrationCatalogue";

const definition = (name: string, displayName: string, capabilities: string[]): InstallableIntegration => ({
	name,
	displayName,
	description: "",
	provider: "test",
	oauthInstall: false,
	capabilities,
	links: [],
});

const installation = (name: string, displayName: string, capabilities: string[] = []) =>
	({
		id: `${name}-${displayName}`,
		attributes: {
			name,
			displayName,
			provider: "test",
			providerInstallationRef: displayName,
			capabilities,
			sanitizedConfig: {},
			userSettings: {},
		},
	}) as IntegrationInstallation;

const slack = (installations: IntegrationInstallation[]): CatalogueProviderInput => ({
	displayName: "Slack",
	connectionNoun: "workspaces",
	installable: [definition("slack_agent", "Slack Agent", ["chat_context"])],
	installations,
});

describe("connectionStatus", () => {
	test("is empty without connections", () => {
		expect(connectionStatus(slack([]))).toBeUndefined();
	});

	test("hides connection names that repeat the app name", () => {
		const google: CatalogueProviderInput = {
			displayName: "Google Workspace",
			connectionNoun: "workspaces",
			installable: [definition("google", "Google Workspace", ["video_conferencing"])],
			installations: [installation("google", "Google Workspace")],
		};

		expect(connectionStatus(google)).toBe("Connected");
	});

	test("names one workspace and lists its apps", () => {
		const status = connectionStatus(
			slack([installation("slack_agent", "Acme"), installation("slack_incidents", "Acme")])
		);

		expect(status).toBe("Acme · Agent, Incidents");
	});

	test("counts several connections", () => {
		const status = connectionStatus(
			slack([installation("slack_agent", "Acme"), installation("slack_agent", "Globex")])
		);

		expect(status).toBe("2 workspaces");
	});
});

describe("capabilityPills", () => {
	test("labels user-facing capabilities once and leaves out internal ones", () => {
		const github: CatalogueProviderInput = {
			displayName: "GitHub",
			connectionNoun: "accounts",
			installable: [definition("github", "Github", ["event_sync", "code_changes", "repositories"])],
			installations: [installation("github", "acme", ["event_sync", "code_changes", "repositories"])],
		};

		expect(capabilityPills(github)).toEqual(["Code changes", "Repositories"]);
	});
});
