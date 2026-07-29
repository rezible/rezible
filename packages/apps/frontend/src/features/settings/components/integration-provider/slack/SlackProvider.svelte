<script lang="ts">
	import { Button } from "$components/ui/button";
	import { useUserSessionState } from "$lib/user-session.svelte";

	import { useIntegrationProviderConfigController } from "../controller.svelte";
	import SlackIncidentManagement from "./SlackIncidentManagement.svelte";
	import Header from "$src/components/layout/header/Header.svelte";
	import SlackAgent from "./SlackAgent.svelte";

	const ctrl = useIntegrationProviderConfigController();
	const session = useUserSessionState();
</script>

<div class="grid gap-4">
	<div class="flex flex-col gap-2">
		<Header title="Agents">
			{#snippet actions()}
				<Button onclick={() => ctrl.startOAuthFlow("slack_agent")} variant="outline">Install</Button>
			{/snippet}
		</Header>

		{#each ctrl.installationsFor("slack_agent") as installation}
			<SlackAgent {installation} />
		{/each}
	</div>

	{#if !!session.orgPreferences?.enableIncidentManagement}
		<SlackIncidentManagement />
	{/if}
</div>
