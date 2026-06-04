<script lang="ts">
	import { useInitialSetupViewController } from "../controller.svelte";

	const ctrl = useInitialSetupViewController();
	const org = $derived(ctrl.session.org);
</script>

<div class="flex flex-col gap-4">
	<div class="space-y-1">
		<h2 class="text-lg font-semibold">Organization details</h2>
		<p class="text-sm text-muted-foreground">Confirm the workspace created for this organization.</p>
	</div>

	<dl class="grid gap-3 sm:grid-cols-2">
		<div class="border border-border bg-background p-3">
			<dt class="text-sm text-muted-foreground">Name</dt>
			<dd class="mt-1 font-medium">{org?.attributes.name || "Organization"}</dd>
		</div>

		<div class="border border-border bg-background p-3">
			<dt class="text-sm text-muted-foreground">Organization ID</dt>
			<dd class="mt-1 truncate font-mono text-sm">{org?.id}</dd>
		</div>
	</dl>
</div>
