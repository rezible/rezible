<script lang="ts">
	import { mdiFilter } from "@mdi/js";
	import { Header, Button, ToggleGroup, ToggleOption } from "svelte-ux";
	import ShiftEventsHeatmap from "./ShiftEventsHeatmap.svelte";
	import ShiftEventsList from "./ShiftEventsList.svelte";

	let showFilters = $state(false);

	let display = $state<"list" | "heatmap">("list");

	const onHeatmapClicked = (day: number, hour: number) => {
		display = "list";
	}
</script>

<div class="flex flex-col h-full border border-surface-content/10 rounded">
	<div class="h-fit pt-2 flex flex-col gap-2">
		<Header title="Shift Events" subheading="Showing All" class="px-2">
			<svelte:fragment slot="actions">
				<Button icon={mdiFilter} iconOnly on:click={() => (showFilters = !showFilters)} />
			</svelte:fragment>
		</Header>

		{#if showFilters}
			<div class="w-full h-12 border"></div>
		{/if}
	</div>

	<div class="p-2">
		<ToggleGroup bind:value={display} inset variant="fill-surface" rounded>
			<ToggleOption value="list">List</ToggleOption>
			<ToggleOption value="heatmap">Heatmap</ToggleOption>
		</ToggleGroup>
	</div>

	<div class="flex-1 flex flex-col gap-1 px-0 overflow-y-auto">
		{#if display === "list"}
			<ShiftEventsList />
		{:else}
			<ShiftEventsHeatmap onHourClicked={onHeatmapClicked} />
		{/if}
	</div>
</div>