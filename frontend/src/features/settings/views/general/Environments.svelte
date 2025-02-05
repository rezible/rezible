<script lang="ts">
	import { z } from "zod";
	import {
		archiveEnvironmentMutation,
		createEnvironmentMutation,
		listEnvironmentsOptions,
		updateEnvironmentMutation,
		type Environment,
	} from "$lib/api";
	import MutatingTable, { makeField } from "$features/settings/components/mutating-table";

	const fields = {
		["name"]: makeField("Environment Name", z.string()),
	};

	const queryOptions = {
		list: listEnvironmentsOptions,
		create: createEnvironmentMutation,
		update: updateEnvironmentMutation,
		archive: archiveEnvironmentMutation,
	};
</script>

<MutatingTable
	dataType="Environment"
	description="Specify the different operational environments services may operate in."
	headers={["Name"]}
	{fields}
	{queryOptions}
>
	{#snippet dataRow(env: Environment)}
		<td>{env.attributes.name}</td>
	{/snippet}
</MutatingTable>
