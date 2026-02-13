<script lang="ts">
	import EditableStringList from "$components/editable-list/EditableStringList.svelte";
	import { eventAttributes } from "./eventAttributesState.svelte";

</script>

<div class="flex flex-col gap-1 bg-surface-100">
	<EditableStringList 
		title="Options Considered"
		addLabel="Add Option"
		bind:values={eventAttributes.decisionContext.optionsConsidered}
	/>

	<EditableStringList 
		title="Constraints"
		addLabel="Add Constraint"
		bind:values={eventAttributes.decisionContext.constraints}
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
