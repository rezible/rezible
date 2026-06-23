<script lang="ts">
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";

	import { useIntegrationProviderConfigController } from "../controller.svelte";

	const ctrl = useIntegrationProviderConfigController();

    const enable = () => {
        ctrl.setEditing("demo");
        ctrl.saveInstall();
    }
</script>

{#if ctrl.installations.length > 0}
	<Alert.Root>
		<Alert.Title>Demo Data Provider</Alert.Title>
		<Alert.Description>Enabled.</Alert.Description>
	</Alert.Root>
{:else}
	<div class="flex flex-row gap-6">
		<div class="place-self-center">
			<Button
				onclick={enable}
				variant="default"
                size="lg"
			>
				Install
			</Button>
		</div>
	</div>
{/if}
