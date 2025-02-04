<script lang="ts">
	import {
		archiveIncidentTagMutation,
		createIncidentTagMutation,
		listIncidentTagsOptions,
		updateIncidentTagMutation,
		type IncidentTag,
	} from "$lib/api";
	import { z } from "zod";
	import MutatingTable, {
		makeField,
	} from "$features/settings/components/mutating-table";

	const fields = {
		["value"]: makeField("Value", z.string().min(4)),
	};

	const queryOptions = {
		list: listIncidentTagsOptions,
		create: createIncidentTagMutation,
		update: updateIncidentTagMutation,
		archive: archiveIncidentTagMutation,
	};
</script>

<MutatingTable
	dataType="Incident Tag"
	description="Modify incident tags."
	headers={["Value"]}
	{fields}
	{queryOptions}
>
	{#snippet dataRow(tag: IncidentTag)}
		<td>{tag.attributes.value}</td>
	{/snippet}
</MutatingTable>
