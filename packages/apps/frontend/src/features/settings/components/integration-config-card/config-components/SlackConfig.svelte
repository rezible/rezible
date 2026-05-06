<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Badge } from "$components/ui/badge";
	import { useIntegrationConfigController } from "../controller.svelte";

	const ctrl = useIntegrationConfigController();

	const teams = $derived(
		ctrl.configured.map((configured) => {
			const config = configured.attributes.config as Record<string, any>;
			return {
				id: configured.id,
				name: config?.team?.name ?? configured.attributes.displayName,
			};
		})
	);
</script>

{#if ctrl.hasConfigured}
	<Alert.Root>
		<Alert.Title>Slack connected</Alert.Title>
		<Alert.Description class="flex items-center gap-2">
			<span>Authentication is configured via OAuth.</span>
			{#each teams as team (team.id)}
				<Badge variant="secondary">{team.name}</Badge>
			{/each}
		</Alert.Description>
	</Alert.Root>
{:else}
	<Alert.Root>
		<Alert.Title>Connect Slack</Alert.Title>
		<Alert.Description>
			Use OAuth to connect the workspace and grant chat/user data access.
		</Alert.Description>
	</Alert.Root>
{/if}
