<script lang="ts">
	import Services from './Services.svelte';
	import LinkedIncidents from './LinkedIncidents.svelte';
	import IncidentVisibility from './IncidentVisibility.svelte';
	import IncidentEnvironments from './IncidentEnvironments.svelte';
	import IncidentSeverity from './IncidentSeverity.svelte';
	import RoleAssignments from './RoleAssignments.svelte';
	import TeamAssignments from './TeamAssignments.svelte';
	import RetrospectivePreview from './RetrospectivePreview.svelte';
	import IncidentStatus from './IncidentStatus.svelte';
    import type { Incident } from '$lib/api';

	type Props = {
		incident: Incident;
		invalidateQuery: () => void;
	};
	const { incident, invalidateQuery }: Props = $props();
</script>

<div class="flex flex-col gap-2 overflow-y-auto w-1/2">
	<div class="">
		<RetrospectivePreview {incident} />
	</div>

	<div class="grid grid-cols-2 gap-2">
		<div class="flex flex-col gap-2">
			<div class="border rounded-lg p-2 group">
				<RoleAssignments {incident} />
			</div>

			<div class="border rounded-lg p-2 group">
				<TeamAssignments {incident} />
			</div>

			<div class="border rounded-lg p-2 group">
				<Services {incident} />
			</div>

			<div class="border rounded-lg p-2 group">
				<LinkedIncidents {incident} />
			</div>
		</div>

		<div class="col-span-1 flex flex-col gap-2 w-fit">
			<div class="border rounded-lg p-2 group">
				<IncidentSeverity {incident} {invalidateQuery} />
			</div>

			<div class="border rounded-lg p-2 group">
				<IncidentVisibility {incident} {invalidateQuery} />
			</div>
		</div>
	</div>
</div>
