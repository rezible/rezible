<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Badge } from "$components/ui/badge";
	import { useIntegrationConfigController } from "../controller.svelte";

	const ctrl = useIntegrationConfigController();

	const parsedConfig = $derived<Record<string, any>>(ctrl.configured?.attributes.config || {});
	const team = $derived(parsedConfig?.team?.name);
</script>

{#if ctrl.configured}
	<Alert.Root>
		<Alert.Title>Slack connected</Alert.Title>
		<Alert.Description class="flex items-center gap-2">
			<span>Authentication is configured via OAuth.</span>
			{#if team}
				<Badge variant="secondary">{team}</Badge>
			{/if}
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
