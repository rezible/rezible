<script lang="ts">
	import type { Snippet } from "svelte";
	import RiLinkUnlink from "remixicon-svelte/icons/link-unlink";
	import RiRefreshLine from "remixicon-svelte/icons/refresh-line";
	import RiSlackFill from "remixicon-svelte/icons/slack-fill";

	import { resolve } from "$app/paths";
	import type { IntegrationInstallation } from "$lib/api";
	import * as Alert from "$components/ui/alert";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";

	import type { IntegrationSettingsForms } from "$features/settings/lib/integrationSettingsForms.svelte";

	import { useIntegrationDataSyncController } from "../../integration-datasync-dialog/controller.svelte";
	import IntegrationOAuthInstall from "../../integration-oauth-install/IntegrationOAuthInstall.svelte";
	import { useIntegrationProviderController } from "../controller.svelte";
	import type { SlackIncidentSettings } from "./slackIncidentSettings";
	import SlackIncidentSettingsForm from "./SlackIncidentSettingsForm.svelte";
	import { slackAgentName, slackIncidentsName, type SlackWorkspace } from "./slackWorkspaces";

	type Props = {
		workspace: SlackWorkspace;
		incidentForms: IntegrationSettingsForms<SlackIncidentSettings>;
		canAddAgent: boolean;
		canAddIncidents: boolean;
		incidentManagementEnabled: boolean;
	};
	const { workspace, incidentForms, canAddAgent, canAddIncidents, incidentManagementEnabled }: Props =
		$props();

	const ctrl = useIntegrationProviderController();
	const sync = useIntegrationDataSyncController();

	const showIncidents = $derived(!!workspace.incidents || canAddIncidents);
</script>

{#snippet appRow(
	name: string,
	title: string,
	description: string,
	installation: IntegrationInstallation | undefined,
	canAdd: boolean,
	body?: Snippet<[IntegrationInstallation]>
)}
	<div class="flex flex-col gap-4 p-4">
		<div class="flex flex-wrap items-start justify-between gap-3">
			<div class="flex min-w-0 flex-col gap-0.5">
				<div class="flex items-center gap-2">
					<span class="font-medium">{title}</span>
					{#if installation}
						<Badge variant="secondary">Installed</Badge>
					{/if}
				</div>
				<span class="text-sm text-muted-foreground">{description}</span>
			</div>

			{#if installation}
				<div class="flex flex-wrap gap-2">
					{#if installation.attributes.capabilities.includes("event_sync")}
						<Button variant="outline" size="sm" onclick={() => sync.openFor(installation)}>
							<RiRefreshLine />
							Sync
						</Button>
					{/if}
					<Button variant="outline" size="sm" onclick={() => ctrl.requestRemoval(installation)}>
						<RiLinkUnlink />
						Disconnect
					</Button>
				</div>
			{:else if canAdd}
				<IntegrationOAuthInstall
					{name}
					origin="{name}:{workspace.ref}"
					label="Add to workspace"
					variant="outline"
					size="sm"
				/>
			{/if}
		</div>

		{#if installation && body}
			{@render body(installation)}
		{/if}
	</div>
{/snippet}

{#snippet incidentSettings(installation: IntegrationInstallation)}
	{#if incidentManagementEnabled}
		<SlackIncidentSettingsForm {installation} forms={incidentForms} />
	{:else}
		<Alert.Root>
			<Alert.Title>Incident management is turned off</Alert.Title>
			<Alert.Description>
				<span>
					Turn it on in
					<a class="underline underline-offset-4" href={resolve("/settings/incidents")}>
						incident management settings
					</a>
					to configure incident channels.
				</span>
			</Alert.Description>
		</Alert.Root>
	{/if}
{/snippet}

<div class="flex flex-col border">
	<div class="flex min-w-0 items-start gap-3 border-b bg-muted/40 px-4 py-3">
		<RiSlackFill class="mt-0.5 size-4 shrink-0" />
		<div class="flex min-w-0 flex-wrap items-baseline gap-x-3 gap-y-0.5">
			<span class="font-medium">{workspace.displayName}</span>
			<span class="text-xs break-all text-muted-foreground">Workspace ID {workspace.ref}</span>
		</div>
	</div>

	<div class="flex flex-col divide-y">
		{@render appRow(
			slackAgentName,
			"Rezible Agent",
			"Ask the Rezible agent questions and get investigation help in Slack.",
			workspace.agent,
			canAddAgent
		)}

		{#if showIncidents}
			{@render appRow(
				slackIncidentsName,
				"Incident Management",
				"Create a Slack channel for each incident and keep it updated from Rezible.",
				workspace.incidents,
				canAddIncidents,
				incidentSettings
			)}
		{/if}
	</div>
</div>
