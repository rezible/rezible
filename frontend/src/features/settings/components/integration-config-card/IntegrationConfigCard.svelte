<script lang="ts">
	import type { ConfiguredIntegration, ConfigureIntegrationRequestBody, SupportedIntegration } from '$lib/api';
	import * as Card from "$components/ui/card";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import * as Alert from "$components/ui/alert";
	import type { IntegrationConfigComponent, IntegrationConfigPayload } from './types';
	import SlackConfig from './config-components/SlackConfig.svelte';
	import PlaceholderConfig from './config-components/PlaceholderConfig.svelte';
	import GoogleConfig from './config-components/GoogleConfig.svelte';

	type Props = {
		integration: SupportedIntegration;
		configured?: ConfiguredIntegration;
		startOAuthFlow?: () => void;
		configureIntegration?: (attrs: ConfigureIntegrationRequestBody["attributes"]) => Promise<unknown> | unknown;
		isSaving?: boolean;
		errorMessage?: string;
	};
	const {
		integration,
		configured,
		startOAuthFlow,
		configureIntegration,
		isSaving = false,
		errorMessage = "",
	}: Props = $props();

	const configs: Record<string, IntegrationConfigComponent> = {
		slack: SlackConfig,
		google: GoogleConfig,
	};
	const ConfigComponent = $derived((integration.name in configs) ? configs[integration.name] : PlaceholderConfig);
	const enabledDataKinds = $derived(configured?.attributes.enabledDataKinds ?? []);
	const requiresOAuthConnect = $derived(integration.oauthRequired && !configured);
	const supportsManualSave = $derived(integration.name === "google");

	let configPayload = $state<IntegrationConfigPayload>({});
	let hasConfigChanges = $state(false);

	const onConfigChange = (payload: IntegrationConfigPayload) => {
		if (payload.config !== undefined) {
			configPayload.config = payload.config;
		}
		if (payload.preferences !== undefined) {
			configPayload.preferences = payload.preferences;
		}
		hasConfigChanges = true;
	};

	const doConfigureIntegration = async () => {
		if (!configureIntegration) return;
		if (!hasConfigChanges) return;

		await configureIntegration({
			config: configPayload.config,
			preferences: configPayload.preferences,
		});
		hasConfigChanges = false;
	};
</script>

{#snippet oauthFlowButtonContent(name: string)}
	{#if name === "slack"}
	<img alt="Add to Slack" width="139px" height="40px"
		src="https://platform.slack-edge.com/img/add_to_slack.png" 
		srcSet="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" />
	{:else}
		<span>Start OAuth Flow</span>
	{/if}
{/snippet}

<Card.Root class="gap-4 p-4 min-w-xs">
	<Card.Header class="p-0">
		<div class="flex items-center justify-between gap-4 h-fit">
			<div class="flex flex-col gap-2">
				<Card.Title class="capitalize">{integration.name}</Card.Title>
				{#if enabledDataKinds.length > 0}
					<div class="flex items-center gap-2 text-sm text-muted-foreground">
						<span>Enabled:</span>
						<div class="flex flex-wrap gap-2">
							{#each enabledDataKinds as kind}
								<Badge variant="secondary">{kind}</Badge>
							{/each}
						</div>
					</div>
				{/if}
			</div>
			<Badge variant={configured ? "secondary" : "outline"}>{configured ? "Configured" : "Not configured"}</Badge>
		</div>
		{#if requiresOAuthConnect}
			<div class="place-self-center">
				<Button onclick={() => startOAuthFlow?.()} class="w-fit h-fit cursor-pointer p-0">
					{@render oauthFlowButtonContent(integration.name)}
				</Button>
			</div>
		{/if}
	</Card.Header>

	{#if !requiresOAuthConnect}
		<Card.Content class="p-0 flex flex-col gap-3">
			{#if !!errorMessage}
				<Alert.Root variant="destructive">
					<Alert.Title>Could not save integration</Alert.Title>
					<Alert.Description>{errorMessage}</Alert.Description>
				</Alert.Root>
			{/if}

			{#if !requiresOAuthConnect}
				<ConfigComponent {integration} {configured} onChange={onConfigChange} />

				{#if supportsManualSave}
					<Button onclick={doConfigureIntegration} disabled={!hasConfigChanges || isSaving}>
						{isSaving ? "Saving..." : "Save"}
					</Button>
				{/if}
			{/if}
		</Card.Content>
	{/if}
</Card.Root>
