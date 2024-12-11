<script lang="ts">
	import { z } from 'zod';
	import MutatingTable, { makeField } from '$features/settings/components/mutating-table';
	import { archiveIncidentTypeMutation, createIncidentTypeMutation, listIncidentTypesOptions, updateIncidentTypeMutation, type IncidentType } from '$lib/api';

	const fields = {
		["name"]: makeField('Name', z.string().min(4))
	};

	const queryOptions = {
		list: listIncidentTypesOptions,
		update: createIncidentTypeMutation,
		create: updateIncidentTypeMutation,
		archive: archiveIncidentTypeMutation,
	}
</script>

<MutatingTable
	dataType="Incident Type"
	description="Modify incident types."
	headers={['Name']}
	{fields}
	{queryOptions}
>
	{#snippet dataRow(it: IncidentType)}
		<td>{it.attributes.name}</td>
	{/snippet}
</MutatingTable>