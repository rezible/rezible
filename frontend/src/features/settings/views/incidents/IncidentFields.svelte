<script lang="ts">
	import { z } from "zod";
	import {
		archiveIncidentFieldMutation,
		createIncidentFieldMutation,
		listIncidentFieldsOptions,
		listIncidentTypesOptions,
		updateIncidentFieldMutation,
		type IncidentField,
	} from "$lib/api";
	import MutatingTable, {
		makeField,
		makeCustomField,
		makeSelectField,
	} from "$features/settings/components/mutating-table";
	import IncidentFieldsOptionsEdit from "./IncidentFieldsOptionsEdit.svelte";

	const incidentFieldOptionsSchema = z
		.array(
			z.object({
				id: z.string().optional(),
				field_option_type: z.enum(["custom", "derived"]),
				value: z.string(),
				archived: z.boolean().optional(),
			})
		)
		.min(1);

	const fields = {
		["name"]: makeField("Field Name", z.string().min(4)),
		["options"]: makeCustomField("Field Options", incidentFieldOptionsSchema, IncidentFieldsOptionsEdit),
		["incident_type"]: makeSelectField(
			"Incident Types",
			z.string().optional().describe("Restrict to Specific Incident Types?"),
			listIncidentTypesOptions
		),
		["required"]: makeField("Required", z.boolean()),
	};

	const queryOptions = {
		list: listIncidentFieldsOptions,
		create: createIncidentFieldMutation,
		update: updateIncidentFieldMutation,
		archive: archiveIncidentFieldMutation,
	};
</script>

<MutatingTable
	dataType="Incident Field"
	description="Modify incident fields"
	headers={["Name", "Options", "Incident Type", "Required"]}
	{fields}
	{queryOptions}
>
	{#snippet dataRow(f: IncidentField)}
		<td>{f.attributes.name}</td>
		<td>
			{#each f.attributes.options as opt}
				<span>{opt.attributes.value}</span>
			{/each}
		</td>
		<td>
			{f.attributes.incident_type?.attributes.name || "Any"}
		</td>
		<td>{f.attributes.required ? "Yes" : "No"}</td>
	{/snippet}
</MutatingTable>
