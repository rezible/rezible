<script lang="ts">
	import EditableList from "$components/forms/editable-list/EditableList.svelte";
	import { useEventDialogAttributes } from "./attributes.svelte";

	const attributes = useEventDialogAttributes();
</script>

<div class="flex flex-col gap-1 bg-surface-100">
	<EditableList 
		title="Options Considered"
		addLabel="Add Option"
		bind:values={attributes.decisionContext.optionsConsidered}
	/>

	<EditableList 
		title="Constraints"
		addLabel="Add Constraint"
		bind:values={attributes.decisionContext.constraints}
	/>

	<div class="flex flex-col gap-2 border p-2">
		<span class="text-surface-content">Decision Rationale</span>

		<span>rationale field</span>
		<!-- <TextField 
			label="Rationale for the decision made"
			multiline
			bind:value={eventAttributes.decisionContext.decisionRationale} /> -->
	</div>
</div>
