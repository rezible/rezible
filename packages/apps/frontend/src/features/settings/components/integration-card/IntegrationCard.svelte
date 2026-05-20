<script lang="ts">
	import type { AvailableIntegration } from "$lib/api";

	import * as Card from "$components/ui/card";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import { initIntegrationCardController } from "./controller.svelte";
	import InlineAlert from "$src/components/inline-alert/InlineAlert.svelte";
	import Spinner from "$src/components/ui/spinner/spinner.svelte";
	import RiGithubFill from "remixicon-svelte/icons/github-fill";
	import { Checkbox } from "$components/ui/checkbox";
	import { useIntegrationOAuthController } from "$features/settings/lib/integrationOAuthController.svelte";

	type Props = {
		integration: AvailableIntegration;
	};
	const { integration }: Props = $props();

	const ctrl = initIntegrationCardController(() => integration);
	const oauth = useIntegrationOAuthController();
</script>

{#snippet oauthFlowButtonContent()}
	{#if ctrl.name === "slack"}
		<img
			alt="Add to Slack"
			width="139px"
			height="40px"
			src="https://platform.slack-edge.com/img/add_to_slack.png"
			srcSet="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x"
		/>
	{:else if ctrl.name === "github"}
		<span
			class="inline-flex h-10 items-center gap-2 rounded-md bg-foreground px-4 text-sm font-medium text-background"
		>
			<RiGithubFill class="size-5" />
			Connect GitHub
		</span>
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
							{#each ctrl.enabledDataKinds as kind (kind)}
								<Badge variant="secondary">{kind}</Badge>
							{/each}
						</div>
					</div>
				{/if}
			</div>
			<Badge variant={ctrl.hasConfigured ? "secondary" : "outline"}
				>{ctrl.hasConfigured ? `${ctrl.configured.length} configured` : "Not configured"}</Badge
			>
		</div>
	</Card.Header>

	<Card.Content class="p-0 flex flex-col gap-3">
		{#if !!ctrl.configError}
			<InlineAlert bind:error={ctrl.configError} />
		{/if}
		{#if !!ctrl.syncStatusError}
			<InlineAlert error={ctrl.syncStatusError} dismissable={false} />
		{/if}

		{#if ctrl.oauthSelectionRequired}
			<div class="flex flex-col gap-3 rounded-md border p-3">
				<div class="flex flex-col gap-1">
					<span class="text-sm font-medium">Select installations</span>
					<span class="text-sm text-muted-foreground">Choose which accounts to connect.</span>
				</div>
				<div class="flex flex-col gap-2">
					{#each oauth.selectionOptions as option (option.externalRef)}
						<label class="flex items-center gap-3 rounded-md border p-3 text-sm">
							<Checkbox
								checked={oauth.selectedExternalRefs.has(option.externalRef)}
								onCheckedChange={(checked) =>
									oauth.toggleSelection(option.externalRef, !!checked)}
							/>
							<span class="flex flex-col">
								<span class="font-medium">{option.displayName}</span>
								<span class="text-muted-foreground">{option.externalRef}</span>
							</span>
						</label>
					{/each}
				</div>
				<Button
					disabled={ctrl.loading || oauth.selectedExternalRefs.size === 0}
					onclick={() => oauth.selectOAuthOptions()}
				>
					{#if ctrl.loading}
						<Spinner />
					{/if}
					Connect selected
				</Button>
			</div>
		{:else if integration.oauthRequired && !ctrl.hasConfigured}
			<div class="place-self-center">
				<Button
					onclick={() => {
						ctrl.startOAuthFlow();
					}}
					variant="ghost"
					class="w-fit h-fit cursor-pointer p-0"
				>
					{@render oauthFlowButtonContent()}
				</Button>
			</div>
		{:else}
			{#if ctrl.hasConfigured}
				<div class="flex flex-col gap-2">
					{#each ctrl.configured as configured (configured.id)}
						<div class="flex items-center justify-between gap-3 rounded-md border p-3 text-sm">
							<div class="min-w-0">
								<div class="truncate font-medium">{configured.attributes.displayName}</div>
								<div class="truncate text-muted-foreground">
									{configured.attributes.externalRef}
								</div>
							</div>
							<div class="flex flex-wrap justify-end gap-1">
								{#each Object.entries(configured.attributes.dataKinds).filter(([, enabled]) => enabled) as [kind] (kind)}
									<Badge variant="secondary">{kind}</Badge>
								{/each}
							</div>
						</div>
					{/each}
				</div>
				<div class="flex items-center justify-between gap-3 rounded-md border p-3 text-sm">
					<div class="flex min-w-0 items-center gap-2">
						<span class="text-muted-foreground">Data sync</span>
						{#if ctrl.latestSyncStatusDisplay}
							<Badge
								variant={ctrl.latestSyncStatusDisplay.variant}
								class={ctrl.latestSyncStatusDisplay.class}
							>
								{ctrl.latestSyncStatusDisplay.label}
							</Badge>
						{:else}
							<Badge variant="outline">No runs</Badge>
						{/if}
						{#if ctrl.isSyncing}
							<Spinner aria-label="Sync status updating" />
						{/if}
					</div>
					<Button
						variant="ghost"
						size="sm"
						disabled={ctrl.loading || ctrl.isSyncing}
						onclick={() => {
							ctrl.refetchSyncStatus();
						}}
					>
						Refresh
					</Button>
				</div>
			{/if}

			<ctrl.Component />

			<div class="flex flex-wrap gap-2">
				{#if integration.oauthRequired}
					<Button
						onclick={() => {
							ctrl.startOAuthFlow();
						}}
						variant="outline"
						disabled={ctrl.loading}
					>
						Connect another
					</Button>
				{/if}
				<Button disabled={!ctrl.hasChanges || ctrl.loading} onclick={() => ctrl.save()}>
					{#if ctrl.loading}
						<Spinner />
						Saving...
					{:else}
						Save
					{/if}
				</Button>
				<Button
					onclick={() => {
						ctrl.requestSync();
					}}
					variant="outline"
					disabled={!ctrl.hasConfigured || ctrl.loading || ctrl.isSyncing}
				>
					{#if ctrl.isSyncing}
						<Spinner />
						Syncing...
					{:else}
						Request Data Sync
					{/if}
				</Button>
			</div>
		{/if}
	</Card.Content>
</Card.Root>
