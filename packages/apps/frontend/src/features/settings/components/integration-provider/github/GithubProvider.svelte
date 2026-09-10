<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import RiGithubFill from "remixicon-svelte/icons/github-fill";
	import { useIntegrationDataSyncController } from "../../integration-datasync-dialog/controller.svelte";
	import { useIntegrationProviderConfigController } from "../controller.svelte";

	const ctrl = useIntegrationProviderConfigController();
	const sync = useIntegrationDataSyncController();

	const installations = $derived(ctrl.installationsFor("github"));

	const installationDetails = (curr: (typeof installations)[number]) => {
		if (!curr) return;
		const config = curr.attributes.sanitizedConfig;
		const org = typeof config.org === "string" ? config.org : curr.attributes.displayName;
		const installationId =
			typeof config.installation_id === "number" || typeof config.installation_id === "string"
				? String(config.installation_id)
				: curr.attributes.providerInstallationRef;
		return { org, installationId };
	};
</script>

<div class="grid gap-4">
	<Card.Root>
		<Card.Header>
			<Card.Title>GitHub</Card.Title>
		</Card.Header>
		<Card.Content class="grid gap-3">
			{#if installations.length === 0}
				<Alert.Root>
					<Alert.Title>Connect GitHub</Alert.Title>
					<Alert.Description>
						Sign in with GitHub to install the GitHub app and grant repository/change event
						access.
					</Alert.Description>
				</Alert.Root>
			{/if}

			<Button onclick={() => ctrl.startOAuthFlow("github")} variant="outline" class="w-fit">
				<RiGithubFill class="size-4" />
				Connect GitHub
			</Button>
		</Card.Content>
	</Card.Root>

	{#each installations as installation (installation.id)}
		{@const details = installationDetails(installation)}
		<Card.Root>
			<Card.Header>
				<Card.Title>{details?.org ?? installation.attributes.displayName}</Card.Title>
				<Card.Action>
					<Badge variant="outline"
						>Installation {details?.installationId ??
							installation.attributes.providerInstallationRef}</Badge
					>
				</Card.Action>
			</Card.Header>
			<Card.Footer class="gap-2">
				{#if installation.attributes.capabilities.includes("event_sync")}
					<Button variant="outline" onclick={() => sync.openFor(installation)}>Sync</Button>
				{/if}
				<Button variant="destructive" onclick={() => ctrl.disconnect(installation.id)}
					>Disconnect</Button
				>
			</Card.Footer>
		</Card.Root>
	{/each}
</div>
