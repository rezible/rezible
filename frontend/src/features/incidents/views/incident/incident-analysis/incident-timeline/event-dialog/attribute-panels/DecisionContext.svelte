<script lang="ts">
	import { TextField } from "svelte-ux";
	import { useEventDialog } from "../dialogState.svelte";
	import EditableStringList from "$components/editable-list/EditableStringList.svelte";

	const eventDialog = useEventDialog();
</script>

<div class="flex flex-col gap-1 bg-surface-100">
	<EditableStringList 
		title="Options Considered"
		addLabel="Add Option"
		bind:values={eventDialog.eventAttributes.decisionContext.optionsConsidered}
	/>

	<EditableStringList 
		title="Constraints"
		addLabel="Add Constraint"
		bind:values={eventDialog.eventAttributes.decisionContext.constraints}
	/>

	<div class="flex flex-col gap-2 border p-2">
		<span class="text-surface-content">Decision Rationale</span>

		<TextField 
			label="Rationale for the decision made"
			multiline
			bind:value={eventDialog.eventAttributes.decisionContext.decisionRationale} />
	</div>
</div>
