<script lang="ts">
	import Spinner from "$components/ui/spinner/spinner.svelte";
	import * as Alert from "$components/ui/alert";
	import { Button } from "$components/ui/button";

	import { initConnectIntegrationController } from "./controller.svelte";

	type Props = {
		name: string;
	};

	const { name }: Props = $props();
	const ctrl = initConnectIntegrationController(() => name);
</script>

<div class="grid min-h-full w-full place-items-center p-6">
	<div class="flex w-full max-w-md flex-col gap-4 rounded-md border border-border bg-background p-4">
		{#if !ctrl.done}
			<div class="flex items-center gap-2">
				<Spinner />
				<span>Finishing integration setup...</span>
			</div>
		{:else if ctrl.error}
			<Alert.Root variant="destructive">
				<Alert.Title>{ctrl.error.title}</Alert.Title>
				<Alert.Description>{ctrl.error.detail}</Alert.Description>
			</Alert.Root>
			{#if ctrl.notifiedOpener}
				<p class="text-sm text-muted-foreground">This window will close automatically.</p>
			{/if}
		{:else}
			<Alert.Root>
				<Alert.Title>Integration connected</Alert.Title>
				<Alert.Description>This window will close automatically.</Alert.Description>
			</Alert.Root>
		{/if}

		{#if ctrl.done}
			<Button variant="outline" onclick={() => ctrl.close()}>Close</Button>
		{/if}
	</div>
</div>
