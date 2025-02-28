<script lang="ts">
	import type { Incident } from "$lib/api";
	import { incidentCtx } from "$features/incidents/lib/context.ts";
	import LinkedIncidents from "./LinkedIncidents.svelte";
	import IncidentVisibility from "./IncidentVisibility.svelte";
	import IncidentSeverity from "./IncidentSeverity.svelte";
	import RoleAssignments from "./RoleAssignments.svelte";
	import TeamAssignments from "./TeamAssignments.svelte";

	type Props = {
		incident: Incident;
	};
	const { incident }: Props = $props();


	incidentCtx.set(incident);
</script>

<div class="flex flex-col gap-2 overflow-y-auto">
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
