<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Badge } from "$components/ui/badge";
	import { Button } from "$components/ui/button";
	import RiGithubFill from "remixicon-svelte/icons/github-fill";
	import { useIntegrationProviderConfigController } from "../controller.svelte";

	const ctrl = useIntegrationProviderConfigController();

	const installation = $derived.by(() => {
		const curr = ctrl.installations.at(0);
		if (!curr) return;
		const config = curr.attributes.config;
		const org = typeof config.org === "string" ? config.org : curr.attributes.displayName;
		const installationId =
			typeof config.installation_id === "number" || typeof config.installation_id === "string"
				? String(config.installation_id)
				: curr.attributes.externalRef;
		return { org, installationId };
	});
</script>

{#if installation}
	<Alert.Root>
		<Alert.Title>GitHub connected</Alert.Title>
		<Alert.Description class="flex flex-wrap items-center gap-2">
			<span>Repository and change event access is configured via OAuth.</span>
			<Badge variant="secondary">{installation.org}</Badge>
			<Badge variant="outline">Installation {installation.installationId}</Badge>
		</Alert.Description>
	</Alert.Root>
{:else}
	<Alert.Root>
		<Alert.Title>Connect GitHub</Alert.Title>
		<Alert.Description>
			Sign in with GitHub to install the GitHub app and grant repository/change event access.
		</Alert.Description>
	</Alert.Root>

	<div class="place-self-center">
		<Button
			onclick={() => {
				ctrl.startOAuthFlow("github");
			}}
			variant="ghost"
			class="w-fit h-fit cursor-pointer p-0"
		>
			<span
				class="inline-flex h-10 items-center gap-2 rounded-md bg-foreground px-4 text-sm font-medium text-background"
			>
				<RiGithubFill class="size-5" />
				Connect GitHub
			</span>
		</Button>
	</div>
{/if}
