import type { ConfiguredIntegration, SupportedIntegration } from "$lib/api";
import type { Component } from "svelte";

export type IntegrationConfigPayload = {
	config?: Record<string, unknown>;
	preferences?: Record<string, unknown>;
};

export type ConfigComponentProps = {
	integration: SupportedIntegration;
	configured?: ConfiguredIntegration;
	onChange: (payload: IntegrationConfigPayload) => void;
};

export type IntegrationConfigComponent = Component<ConfigComponentProps, {}>;
