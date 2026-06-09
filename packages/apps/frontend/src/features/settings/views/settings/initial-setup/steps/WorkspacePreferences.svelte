<script lang="ts">
	import CheckboxCard from "$components/forms/checkbox-card/CheckboxCard.svelte";
	import { useInitialSetupViewController } from "../controller.svelte";

	const ctrl = useInitialSetupViewController();
</script>

<div class="flex flex-row gap-4 justify-between">
	<div class="space-y-1">
		<h2 class="text-lg font-semibold">Workspace preferences</h2>
		<p class="text-sm text-muted-foreground">These can be changed later from settings.</p>
	</div>

	<div class="flex flex-col gap-3 flex-1 max-w-sm">
		<CheckboxCard 
			bind:checked={ctrl.workspacePrefs.enableIncidentManagement}
			label="Enable Rezible Incident Management"
			description="Use Rezible to declare and run incidents, using the web UI or Slack."
		/>
	</div>
</div>
