<script lang="ts">
	import { mdiGithub, mdiPencil, mdiPlus, mdiSlack, mdiTrashCan, mdiWeb } from "@mdi/js";
	import { Button, Field, Icon, ListItem, SelectField, TextField } from "svelte-ux";
	import { eventAttributes } from "./eventAttributes.svelte";
	import EditableStringList from "$src/components/editable-list/EditableStringList.svelte";

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

		<TextField 
			label="Rationale for the decision made"
			multiline
			bind:value={eventAttributes.decisionContext.decisionRationale} />
	</div>
</div>
