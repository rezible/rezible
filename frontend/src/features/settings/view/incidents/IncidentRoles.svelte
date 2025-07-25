<script lang="ts">
	import { z } from "zod";
	import MutatingTable, { makeField } from "$features/settings/components/mutating-table";
	import {
		archiveIncidentRoleMutation,
		createIncidentRoleMutation,
		listIncidentRolesOptions,
		updateIncidentRoleMutation,
		type IncidentRole,
	} from "$lib/api";
	import Icon from "$components/icon/Icon.svelte";
	import { mdiThumbDown, mdiThumbUp } from "@mdi/js";

	const fields = {
		["name"]: makeField("Role Name", z.string().min(4)),
		["required"]: makeField("Required", z.boolean()),
	};

	const queryOptions = {
		list: listIncidentRolesOptions,
		create: createIncidentRoleMutation,
		update: updateIncidentRoleMutation,
		archive: archiveIncidentRoleMutation,
	};
</script>

<MutatingTable
	dataType="Incident Role"
	description="Modify the roles assigned to incident responders."
	headers={["Name", "Required"]}
	{fields}
	{queryOptions}
>
	{#snippet dataRow(role: IncidentRole)}
		<td>{role.attributes.name}</td>
		<td>
			<Icon data={role.attributes.required ? mdiThumbUp : mdiThumbDown} />
		</td>
	{/snippet}
</MutatingTable>
