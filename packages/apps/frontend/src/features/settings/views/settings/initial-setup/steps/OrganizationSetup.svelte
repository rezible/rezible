<script lang="ts">
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import CheckboxCard from "$components/forms/checkbox-card/CheckboxCard.svelte";
	import { useInitialSetupViewController } from "../controller.svelte";

	const ctrl = useInitialSetupViewController();
</script>

<div class="flex flex-col gap-2">
	<div class="flex flex-row gap-4 justify-between">
		<div class="space-y-1">
			<h2 class="text-lg font-semibold">Organization Details</h2>
			<p class="text-sm text-muted-foreground">Confirm the details for your organization.</p>
		</div>
		
		<div class="flex flex-col gap-3 flex-1 max-w-sm">
			<div class="flex w-full max-w-sm flex-col gap-1.5">
				<Label for="org-name">Name</Label>
				<Input type="text" id="org-name" placeholder="Name" disabled value={ctrl.orgName} />
			</div>
		</div>
	</div>


	<div class="flex flex-row gap-4 justify-between">
		<div class="space-y-1">
			<h2 class="text-lg font-semibold">Organization preferences</h2>
			<p class="text-sm text-muted-foreground">These can be changed later from settings.</p>
		</div>

		<div class="flex flex-col gap-3 flex-1 max-w-sm">
			<CheckboxCard 
				bind:checked={ctrl.orgPrefs.enableIncidentManagement}
				label="Enable Rezible Incident Management"
				description="Use Rezible to declare and run incidents, using the web UI or Slack."
			/>
		</div>
	</div>
</div>