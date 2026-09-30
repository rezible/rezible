<script lang="ts">
	import RiExternalLinkLine from "remixicon-svelte/icons/external-link-line";
	import RiGithubFill from "remixicon-svelte/icons/github-fill";

	import * as Card from "$components/ui/card";
	import * as Empty from "$components/ui/empty";

	import IntegrationConnections from "../../integration-connections/IntegrationConnections.svelte";
	import IntegrationInstallTargetSelect from "../../integration-install-target-selection/IntegrationInstallTargetSelect.svelte";
	import IntegrationOAuthInstall from "../../integration-oauth-install/IntegrationOAuthInstall.svelte";
	import { useIntegrationProviderController } from "../controller.svelte";

	const integrationName = "github";

	const ctrl = useIntegrationProviderController();

	const canInstall = $derived(ctrl.integrations.canInstall(integrationName));
	const hasConnections = $derived(ctrl.integrations.installationsFor(integrationName).length > 0);
	const installAppLink = $derived(
		ctrl.integrations
			.installableIntegration(integrationName)
			?.links.find((link) => link.kind === "install_app")
	);
</script>

{#snippet connectActions()}
	{#if canInstall}
		<div class="flex flex-wrap items-start gap-2">
			<IntegrationOAuthInstall name={integrationName} label="Connect account">
				{#snippet icon()}
					<RiGithubFill />
				{/snippet}
			</IntegrationOAuthInstall>
			{#if installAppLink}
				<a
					href={installAppLink.url}
					target="_blank"
					rel="noopener noreferrer"
					class="inline-flex h-9 items-center gap-1.5 px-3 text-sm underline-offset-4 hover:underline"
				>
					{installAppLink.label}
					<RiExternalLinkLine class="size-4" />
				</a>
			{/if}
		</div>
	{/if}
{/snippet}

<Card.Root>
	<Card.Header>
		<Card.Title>Accounts</Card.Title>
		<Card.Description>
			Each connection is a GitHub user or organization account. Rezible can only find accounts where the
			Rezible GitHub App is installed, so install the app on an account before connecting it.
		</Card.Description>
	</Card.Header>

	<Card.Content class="flex flex-col gap-4">
		<IntegrationInstallTargetSelect
			name={integrationName}
			title="Choose accounts to connect"
			description="Your GitHub sign-in can access these accounts with the Rezible GitHub App installed."
		/>

		<IntegrationConnections name={integrationName} referenceLabel="Account ID">
			{#snippet empty()}
				<Empty.Root class="border">
					<Empty.Header>
						<Empty.Media variant="icon">
							<RiGithubFill />
						</Empty.Media>
						<Empty.Title>No accounts connected</Empty.Title>
						<Empty.Description>
							Connect a GitHub account to follow its repositories and changes.
						</Empty.Description>
					</Empty.Header>
					<Empty.Content>
						{@render connectActions()}
					</Empty.Content>
				</Empty.Root>
			{/snippet}
		</IntegrationConnections>

		{#if hasConnections}
			{@render connectActions()}
		{/if}
	</Card.Content>
</Card.Root>
