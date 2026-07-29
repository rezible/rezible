<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";

	import { useIntegrationDataSyncController } from "../../integration-datasync-dialog/controller.svelte";
	import { useIntegrationProviderConfigController } from "../controller.svelte";

	const ctrl = useIntegrationProviderConfigController();
	const sync = useIntegrationDataSyncController();
	const installation = $derived(ctrl.installationsFor("demo").at(0));

	const enable = () => {
		ctrl.setEditing("demo");
		ctrl.saveInstall();
	};
</script>

<Card.Root>
	<Card.Header>
		<Card.Title>Demo Data Provider</Card.Title>
	</Card.Header>
	<Card.Content class="grid gap-3">
		{#if installation}
			<Alert.Root>
				<Alert.Title>Installed</Alert.Title>
				<Alert.Description>{installation.attributes.displayName}</Alert.Description>
			</Alert.Root>
			<div class="flex gap-2">
				{#if installation.attributes.capabilities.includes("event_sync")}
					<Button variant="outline" onclick={() => sync.openFor(installation)}>Sync</Button>
				{/if}
				<Button variant="destructive" onclick={() => ctrl.disconnect(installation.id)}
					>Disconnect</Button
				>
			</div>
		{:else}
			<Button onclick={enable} variant="default" class="w-fit">Install</Button>
		{/if}
	</Card.Content>
</Card.Root>
