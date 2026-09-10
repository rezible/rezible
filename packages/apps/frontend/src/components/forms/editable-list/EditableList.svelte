<script lang="ts" generics="T">
	import RiCheckLine from "remixicon-svelte/icons/check-line";
	import RiPencilLine from "remixicon-svelte/icons/pencil-line";
	import RiAddLine from "remixicon-svelte/icons/add-line";
	import RiDeleteBinLine from "remixicon-svelte/icons/delete-bin-line";

	type Props = {
		title: string;
		addLabel?: string;
		values: string[];
	};

	let { title, addLabel = "Add New", values = $bindable() }: Props = $props();

	let editIdx = $state<number>();
	let editValue = $state<string>();
	let newValue = $state<string>();

	const confirmAdd = () => {
		if (!newValue) return;
		values.push($state.snapshot(newValue));
		newValue = undefined;
	};

	const setEditing = (idx: number) => {
		editIdx = idx;
		editValue = $state.snapshot(values[idx]);
	};

	const clearEditing = () => {
		editIdx = undefined;
		editValue = "";
	};

	const confirmEdit = () => {
		if (editIdx === undefined || editIdx < 0 || editIdx >= values.length || !editValue) return;
		values[editIdx] = $state.snapshot(editValue);
		clearEditing();
	};

	const confirmDelete = (idx: number) => {
		const val = idx >= 0 && idx < values.length ? values[idx] : undefined;
		if (val === undefined) return;
		if (!confirm(`Are you sure you want to delete "${val}"?`)) return;
		values.splice(idx, 1);
	};
</script>

<!--div class="flex flex-col gap-2 border p-2">
	<span class="text-foreground">{title}</span>

	{#each values as val, i}
		{#if editIdx !== undefined && editIdx === i}
			<TextField dense clearable label={`Editing "${val}"`}
				bind:value={editValue}
				onkeydown={e => e.key === "Enter" && confirmEdit()}
				onclear={() => {clearEditing()}}
			>
				<span slot="append">
					<Button icon={RiCheckLine} onclick={confirmEdit} />
				</span>
			</TextField>
		{:else}
			<ListItem
				title={val}
				classes={{ root: "border first:border-t rounded elevation-0" }}
				class="flex-1"
				noShadow
			>
				<div slot="actions">
					<Button icon={RiPencilLine} iconOnly onclick={() => {setEditing(i)}} />
					<Button icon={RiDeleteBinLine} iconOnly onclick={() => {confirmDelete(i)}} />
				</div>
			</ListItem>
		{/if}
	{/each}

	<TextField dense clearable label={addLabel}
		bind:value={newValue}
		onkeydown={e => e.key === "Enter" && confirmAdd()}
	>
		<span slot="append">
			<Button
				icon={RiAddLine}
				class="text-foreground/50 p-2"
				onclick={confirmAdd}
				disabled={!newValue}
			/>
		</span>
	</TextField>
</div-->
