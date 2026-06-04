<script lang="ts">
	import { Checkbox } from "$components/ui/checkbox";

	let incidentDefaults = $state(true);
	let integrationSync = $state(true);
</script>

<div class="flex flex-col gap-4">
	<div class="space-y-1">
		<h2 class="text-lg font-semibold">Workspace preferences</h2>
		<p class="text-sm text-muted-foreground">These defaults can be changed later from settings.</p>
	</div>

	<div class="grid gap-3">
		<label class="flex items-start gap-3 border border-border bg-background p-3">
			<Checkbox bind:checked={incidentDefaults} />
			<span class="flex min-w-0 flex-col gap-1">
				<span class="font-medium">Use recommended incident defaults</span>
				<span class="text-sm text-muted-foreground">Start with standard severity and response settings.</span>
			</span>
		</label>

		<label class="flex items-start gap-3 border border-border bg-background p-3">
			<Checkbox bind:checked={integrationSync} />
			<span class="flex min-w-0 flex-col gap-1">
				<span class="font-medium">Enable integration data sync</span>
				<span class="text-sm text-muted-foreground">Keep imported teams, users, and signals current.</span>
			</span>
		</label>
	</div>
</div>
