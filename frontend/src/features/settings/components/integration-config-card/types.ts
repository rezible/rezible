import type { ConfiguredIntegration, SupportedIntegration } from "$lib/api";
import type { Component } from "svelte";

export type ConfigComponentProps = {
    integration: SupportedIntegration;
    configured?: ConfiguredIntegration;
    onConfigChange: (key: string, value: any) => void;
};

export type IntegrationConfigComponent = Component<ConfigComponentProps, {}>;