<script lang="ts">
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import { useInitialSetupViewController } from "../controller.svelte";

	const ctrl = useInitialSetupViewController();
</script>

<div class="flex flex-row gap-4 justify-between">
	<div class="space-y-1">
		<h2 class="text-lg font-semibold">Organization</h2>
		<p class="text-sm text-muted-foreground">Confirm the details for your organization.</p>
	</div>
	
	<div class="flex flex-col gap-3 flex-1 max-w-sm">
		<div class="flex w-full max-w-sm flex-col gap-1.5">
			<Label for="org-name">Organization Name</Label>
			<Input type="text" id="org-name" placeholder="Organization Name" bind:value={ctrl.orgDetails.name} />
		</div>
	</div>
</div>
