<script lang="ts">
	import RiSlackFill from "remixicon-svelte/icons/slack-fill";

	import * as Card from "$components/ui/card";
	import * as Empty from "$components/ui/empty";
	import { useUserSessionState } from "$lib/user-session.svelte";

	import { IntegrationSettingsForms } from "$features/settings/lib/integrationSettingsForms.svelte";

	import IntegrationInstallTargetSelect from "../../integration-install-target-selection/IntegrationInstallTargetSelect.svelte";
	import IntegrationOAuthInstall from "../../integration-oauth-install/IntegrationOAuthInstall.svelte";
	import { useIntegrationProviderController } from "../controller.svelte";
	import { slackIncidentSettings } from "./slackIncidentSettings";
	import SlackWorkspaceCard from "./SlackWorkspaceCard.svelte";
	import { groupSlackWorkspaces, slackAgentName, slackIncidentsName } from "./slackWorkspaces";

	const ctrl = useIntegrationProviderController();
	const session = useUserSessionState();

	const agentInstallations = $derived(ctrl.integrations.installationsFor(slackAgentName));
	const incidentInstallations = $derived(ctrl.integrations.installationsFor(slackIncidentsName));
	const workspaces = $derived(groupSlackWorkspaces(agentInstallations, incidentInstallations));

	const canAddAgent = $derived(ctrl.integrations.canInstall(slackAgentName));
	// Incident management is only installable while Rezible incident management is enabled
	// and no other integration manages incidents.
	const canAddIncidents = $derived(ctrl.integrations.canInstall(slackIncidentsName));
	const incidentManagementEnabled = $derived(!!session.orgPreferences?.enableIncidentManagement);

	// A new workspace starts with the agent, or with incident management when only that is offered.
	const newWorkspaceIntegration = $derived.by(() => {
		if (canAddAgent) {
			return slackAgentName;
		}
		if (canAddIncidents) {
			return slackIncidentsName;
		}
		return undefined;
	});

	const incidentForms = new IntegrationSettingsForms(slackIncidentSettings, () => incidentInstallations);
</script>

{#snippet addWorkspace()}
	{#if newWorkspaceIntegration}
		<IntegrationOAuthInstall name={newWorkspaceIntegration} origin="new-workspace" label="Add to Slack">
			{#snippet icon()}
				<RiSlackFill />
			{/snippet}
		</IntegrationOAuthInstall>
	{/if}
{/snippet}

<Card.Root>
	<Card.Header>
		<Card.Title>Workspaces</Card.Title>
		<Card.Description>
			Add Rezible to a Slack workspace, then choose which Rezible apps it uses. Slack asks which
			workspace to add an app to.
		</Card.Description>
	</Card.Header>

	<Card.Content class="flex flex-col gap-4">
		<IntegrationInstallTargetSelect
			name={slackAgentName}
			title="Choose workspaces"
			description="Select the Slack workspaces to add the Rezible Agent to."
		/>
		<IntegrationInstallTargetSelect
			name={slackIncidentsName}
			title="Choose a workspace"
			description="Select the Slack workspace to manage incidents in."
		/>

		{#if workspaces.length === 0}
			<Empty.Root class="border">
				<Empty.Header>
					<Empty.Media variant="icon">
						<RiSlackFill />
					</Empty.Media>
					<Empty.Title>No workspaces connected</Empty.Title>
					<Empty.Description>
						Add Rezible to Slack to chat with the agent and run incidents in Slack.
					</Empty.Description>
				</Empty.Header>
				<Empty.Content>
					{@render addWorkspace()}
				</Empty.Content>
			</Empty.Root>
		{:else}
			{#each workspaces as workspace (workspace.ref)}
				<SlackWorkspaceCard
					{workspace}
					{incidentForms}
					{canAddAgent}
					{canAddIncidents}
					{incidentManagementEnabled}
				/>
			{/each}
			{@render addWorkspace()}
		{/if}
	</Card.Content>
</Card.Root>
