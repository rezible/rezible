<script lang="ts">
	import Spinner from "$components/ui/spinner/spinner.svelte";

	import { initConnectIntegrationController } from "./controller.svelte";

	type Props = {
		name: string;
	};

	const { name }: Props = $props();
	initConnectIntegrationController(() => name);
</script>

<div class="grid min-h-full w-full place-items-center p-6">
	<div class="flex w-full max-w-md flex-col gap-4 rounded-md border border-border bg-background p-4">
		<div class="flex items-center gap-2">
			<Spinner />
			<span>Finishing installation...</span>
		</div>
	</div>
</div>
