<script lang="ts" module>
	export type ListItemType = {
		label: string;
		value: string;
		archived: boolean;
	};
</script>

<script lang="ts">
	import { mdiArchive, mdiArchiveMinus, mdiPlus, mdiTrashCan } from "@mdi/js";

	type Props = {
		id: string;
		items: ListItemType[];
		deleteItems: boolean;
		onAddItem: (value: string) => void;
		onToggleArchived: (idx: number) => void;
	}

	const { id, items, deleteItems, onAddItem, onToggleArchived }: Props = $props();

	let newVal = $state("");

	const addNewItem = () => {
		onAddItem($state.snapshot(newVal));
		newVal = "";
	};

	const toggleArchiveItem = (idx: number) => {
		onToggleArchived(idx);
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
				<!--Button
					color="warning"
					icon={getItemIcon(item.archived)}
					onclick={() => (toggleArchiveItem(index))}
				></Button-->
			</div>
		{/each}
	{/if}

	<!--TextField {id} type="text" placeholder="Custom Option" bind:value={newVal}>
		<span slot="append">
			<Button
				icon={mdiPlus}
				disabled={!newVal}
				onclick={addNewItem}
				classes={{ root: "align-bottom" }}
			/>
		</span>
	</TextField-->
</div>
