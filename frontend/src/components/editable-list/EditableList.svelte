<script lang="ts" module>
	export type ListItemType = {
		label: string;
		value: string;
		archived: boolean;
	};
</script>

<script lang="ts">
	import { mdiArchive, mdiArchiveMinus, mdiPlus, mdiTrashCan } from '@mdi/js';
	import { createEventDispatcher } from 'svelte';
	import { Button, TextField } from 'svelte-ux';

	export let id;
	export let items: ListItemType[] = [];
	export let deleteItems = false;

	const dispatch = createEventDispatcher();

	let newVal = '';

	const addNewItem = () => {
		//items = [...items, {name: newVal, value: "", archived: false}];
		dispatch('addItem', newVal);
		newVal = '';
	};

	const toggleArchiveItem = (idx: number) => {
		//items[idx].archived = !items[idx].archived;
		dispatch('toggleArchived', idx);
	};

	const getItemIcon = (isArchived: boolean) => {
		if (deleteItems) return mdiTrashCan;
		return isArchived ? mdiArchiveMinus : mdiArchive;
	};
</script>

<div class="flex flex-col w-96">
	{#if items.length > 0}
		{#each items as item, index (item)}
			<div class="flex flex-row items-center">
				<span class="flex-1">{item.label}</span>
				<Button
					iconOnly
					color="warning"
					icon={getItemIcon(item.archived)}
					on:click={() => {
						toggleArchiveItem(index);
					}}
				></Button>
			</div>
		{/each}
	{/if}

	<TextField {id} type="text" placeholder="Custom Option" bind:value={newVal}>
		<span slot="append">
			<Button
				icon={mdiPlus}
				disabled={!newVal}
				on:click={addNewItem}
				classes={{ root: 'align-bottom' }}
			/>
		</span>
	</TextField>
</div>
