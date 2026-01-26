<script lang="ts">
	import type { SupportedIntegration } from '$lib/api';
	import Button from '$components/button/Button.svelte';
	import { useSetupViewState } from "../lib/viewState.svelte";
	import { Checkbox } from 'svelte-ux';
	import { SvelteMap } from 'svelte/reactivity';
	import { watch } from 'runed';

    type Props = {
        integration: SupportedIntegration;
        dataKind: string;
    };
    const { integration, dataKind }: Props = $props();

	const view = useSetupViewState();

    const configured = $derived(view.configuredIntegrationsMap.get(integration.name));
    const configValid = $derived(configured?.attributes.configValid);

    let configMap = $state(new SvelteMap<string, string>());
    let dataKindToggles = $state(new SvelteMap<string, boolean>());

    const getDefaultConfig = (intg: SupportedIntegration) => {
        return new SvelteMap<string, any>();
    }

    type ConfigType = { [k: string]: string };
    watch(() => configured, conf => {
        if (conf) {
            configMap = new SvelteMap(Object.entries(conf.attributes.config as ConfigType));
            dataKindToggles = new SvelteMap(Object.entries(conf.attributes.dataKinds));
        } else {
            // TODO: default config for integration
            configMap = getDefaultConfig(integration);
            dataKindToggles = new SvelteMap(integration.supportedDataKinds.map(kind => ([kind, true])));
        }
    });

    const doConfigureIntegration = () => {
        let config: ConfigType = {};
        configMap.forEach((k, v) => { config[k] = v });

        let dataKinds: { [k: string]: boolean } = {};
        integration.supportedDataKinds.forEach(kind => { dataKinds[kind] = !!dataKindToggles.get(kind) });

        view.doConfigureIntegration(integration.name, { config, dataKinds });
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
        <Button onclick={() => {view.oauth.startFlow(integration.name)}}>
            {@render oauthFlowButtonContent(integration.name)}
        </Button>
    {:else}
        <div class="border p-2 flex flex-col">
            config form
        </div>
        <div class="border p-2 flex flex-col">
            <span>enable data kinds:</span>
            {#each dataKindToggles as [kind, enabled]}
                <Checkbox name={kind} checked={enabled}></Checkbox>
            {/each}
        </div>
        <Button 
            variant="fill-light"
            color="primary"
            onclick={doConfigureIntegration}
        >
            Save
        </Button>
    {/if}
</div>