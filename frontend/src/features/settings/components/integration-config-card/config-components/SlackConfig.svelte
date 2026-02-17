<script lang="ts">
	import type { ConfigComponentProps } from '../types';
	import * as Alert from "$components/ui/alert";
	import { Badge } from "$components/ui/badge";

	const { configured }: ConfigComponentProps = $props();
	const parsedConfig = $derived.by(() => {
		const raw = configured?.attributes.config;
		if (!raw || typeof raw !== "object") return null;
		return raw as Record<string, unknown>;
	});
	const team = $derived.by(() => {
		const candidate = parsedConfig?.Team;
		if (!candidate || typeof candidate !== "object") return null;
		const name = (candidate as Record<string, unknown>).Name;
		if (typeof name !== "string" || name.length === 0) return null;
		return name;
	});
</script>

{#if configured}
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
