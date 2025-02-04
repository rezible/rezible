<script lang="ts">
	import LinkedIncidents from "./LinkedIncidents.svelte";
	import IncidentVisibility from "./IncidentVisibility.svelte";
	import IncidentSeverity from "./IncidentSeverity.svelte";
	import RoleAssignments from "./RoleAssignments.svelte";
	import TeamAssignments from "./TeamAssignments.svelte";

	type Props = {};
	const {}: Props = $props();
</script>

<div class="flex flex-col gap-2 overflow-y-auto">
	<a href="/incidents/demo-incident">other</a>
	<div class="grid grid-cols-2 gap-2">
		<div class="flex flex-col gap-2">
			<div class="border rounded-lg p-2 group">
				<RoleAssignments />
			</div>

			<div class="border rounded-lg p-2 group">
				<TeamAssignments />
			</div>

			<div class="border rounded-lg p-2 group">
				<LinkedIncidents />
			</div>
		</div>

		<div class="flex flex-col gap-2">
			<div class="border rounded-lg p-2 group">
				<IncidentSeverity />
			</div>

			<div class="border rounded-lg p-2 group">
				<IncidentVisibility />
			</div>
		</div>
	</div>
</div>
