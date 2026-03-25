import type { ConfiguredIntegration, AvailableIntegration } from "$lib/api";
import type { Component } from "svelte";

export type ConfigComponentProps = {
	integration: AvailableIntegration;
	configured?: ConfiguredIntegration;
	onConfigChange: (cfg: {[key: string]: unknown}) => void;
	onPreferencesChange: (prefs: {[key: string]: unknown}) => void;
};

export type IntegrationConfigComponent = Component<ConfigComponentProps, {}>;
