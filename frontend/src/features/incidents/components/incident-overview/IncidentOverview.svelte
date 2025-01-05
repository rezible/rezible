<script lang="ts">
	import LinkedIncidents from './LinkedIncidents.svelte';
	import IncidentVisibility from './IncidentVisibility.svelte';
	import IncidentSeverity from './IncidentSeverity.svelte';
	import RoleAssignments from './RoleAssignments.svelte';
	import TeamAssignments from './TeamAssignments.svelte';
    import type { Incident } from '$lib/api';

	type Props = {incident: Incident;};
	const { incident }: Props = $props();
</script>

<div class="flex flex-col gap-2 overflow-y-auto">
	<div class="grid grid-cols-2 gap-2">
		<div class="flex flex-col gap-2">
			<div class="border rounded-lg p-2 group">
				<RoleAssignments {incident} />
			</div>

			<div class="border rounded-lg p-2 group">
				<TeamAssignments {incident} />
			</div>

			<div class="border rounded-lg p-2 group">
				<LinkedIncidents {incident} />
			</div>
		</div>

		<div class="flex flex-col gap-2">
			<div class="border rounded-lg p-2 group">
				<IncidentSeverity {incident} />
			</div>

			<div class="border rounded-lg p-2 group">
				<IncidentVisibility {incident} />
			</div>
		</div>
	</div>
</div>
