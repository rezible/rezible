<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Badge } from "$components/ui/badge";
	import { useIntegrationConfigController } from "../controller.svelte";

	const ctrl = useIntegrationConfigController();

	const installations = $derived(
		ctrl.configured.map((configured) => {
			const config = configured.attributes.config;
			const org = typeof config.org === "string" ? config.org : configured.attributes.displayName;
			const installationId =
				typeof config.installation_id === "number" || typeof config.installation_id === "string"
					? String(config.installation_id)
					: configured.attributes.externalRef;
			return { id: configured.id, org, installationId };
		})
	);
</script>

{#if ctrl.hasConfigured}
	<Alert.Root>
		<Alert.Title>GitHub connected</Alert.Title>
		<Alert.Description class="flex flex-wrap items-center gap-2">
			<span>Repository and change event access is configured via OAuth.</span>
			{#each installations as installation (installation.id)}
				<Badge variant="secondary">{installation.org}</Badge>
				<Badge variant="outline">Installation {installation.installationId}</Badge>
			{/each}
		</Alert.Description>
	</Alert.Root>
{:else}
	<Alert.Root>
		<Alert.Title>Connect GitHub</Alert.Title>
		<Alert.Description>
			Use OAuth to install the GitHub app and grant repository/change event access.
		</Alert.Description>
	</Alert.Root>
{/if}
