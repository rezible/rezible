<script lang="ts">
	import { SvelteMap } from 'svelte/reactivity';
	import type { ConfiguredIntegration, ConfigureIntegrationRequestBody, SupportedIntegration } from '$lib/api';
	import Button from '$components/button/Button.svelte';
	import type { IntegrationConfigComponent } from './types';
	import SlackConfig from './config-components/SlackConfig.svelte';
	import PlaceholderConfig from './config-components/PlaceholderConfig.svelte';
	import GoogleConfig from './config-components/GoogleConfig.svelte';

    type Props = {
        integration: SupportedIntegration;
        configured?: ConfiguredIntegration;
        startOAuthFlow: () => void;
        configureIntegration: (attrs: ConfigureIntegrationRequestBody["attributes"]) => void;
    };
    const { integration, configured, startOAuthFlow, configureIntegration }: Props = $props();

    const configs: Record<string, IntegrationConfigComponent> = {
        "slack": SlackConfig,
        "google": GoogleConfig,
    };
    const ConfigComponent = $derived((integration.name in configs) ? configs[integration.name] : PlaceholderConfig);

    let configMap = new SvelteMap<string, any>();

    const onConfigChange = (key: string, value: any) => {
        configMap.set(key, value);
    }

    const doConfigureIntegration = () => {
        let config: Record<string, any> = {};
        configMap.forEach((v, k) => { config[k] = v });

        let dataKinds: Record<string, boolean> = {};
        integration.supportedDataKinds.forEach(kind => dataKinds[kind] = true);

        configureIntegration({ config, dataKinds });
    }
</script>

{#snippet oauthFlowButtonContent(name: string)}
	{#if name === "slack"}
	<img alt="Add to Slack" height="40" width="139" 
		src="https://platform.slack-edge.com/img/add_to_slack.png" 
		srcSet="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" />
	{:else}
		<span>Start OAuth Flow</span>
	{/if}
{/snippet}

<div class="border p-2 flex flex-col gap-2">
    <span>{integration.name}</span>
    {#if !configured && integration.oauthRequired}
        <Button onclick={() => {startOAuthFlow()}}>
            {@render oauthFlowButtonContent(integration.name)}
        </Button>
    {:else}
        <ConfigComponent {integration} {configured} {onConfigChange} />

        <Button 
            variant="fill-light"
            color="primary"
            onclick={doConfigureIntegration}
        >
            Save
        </Button>
    {/if}
</div>