<script lang="ts">
	import type { 
		ConfiguredIntegration,
		ConfigureIntegrationRequestBody,
		AvailableIntegration,
	} from '$lib/api';

	import * as Card from "$components/ui/card";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { initIntegrationConfigController } from './controller.svelte';
	import InlineAlert from '$src/components/inline-alert/InlineAlert.svelte';
	import Spinner from '$src/components/ui/spinner/spinner.svelte';

	type Props = {
		integration: AvailableIntegration;
	};
	const { integration }: Props = $props();

	const ctrl = initIntegrationConfigController(() => integration);
</script>

{#snippet oauthFlowButtonContent()}
	{#if ctrl.name === "slack"}
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
				{#if ctrl.enabledDataKinds.length > 0}
					<div class="flex items-center gap-2 text-sm text-muted-foreground">
						<span>Enabled:</span>
						<div class="flex flex-wrap gap-2">
							{#each ctrl.enabledDataKinds as kind}
								<Badge variant="secondary">{kind}</Badge>
							{/each}
						</div>
					</div>
				{/if}
			</div>
			<Badge variant={ctrl.configured ? "secondary" : "outline"}>{ctrl.configured ? "Configured" : "Not configured"}</Badge>
		</div>
	</Card.Header>

	<Card.Content class="p-0 flex flex-col gap-3">
		{#if !!ctrl.configError}
			<InlineAlert bind:error={ctrl.configError} />
		{/if}

		{#if integration.oauthRequired && !ctrl.configured}
			<div class="place-self-center">
				<Button onclick={() => {ctrl.startOAuthFlow()}} variant="ghost" class="w-fit h-fit cursor-pointer p-0">
					{@render oauthFlowButtonContent()}
				</Button>
			</div>
		{:else}
			<ctrl.Component />

			<Button disabled={!ctrl.hasChanges || ctrl.loading}>
				{#if ctrl.loading}
					<Spinner />
					Saving...
				{:else}
					Save
				{/if}
			</Button>
		{/if}
	</Card.Content>
</Card.Root>
