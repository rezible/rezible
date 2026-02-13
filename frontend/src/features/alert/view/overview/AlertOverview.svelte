<script lang="ts">
	import { useAlertViewState } from "$features/alert";
	import Avatar from "$components/avatar/Avatar.svelte";
	import AlertMetrics from "./AlertMetrics.svelte";

	const viewState = useAlertViewState();

	const attrs = $derived(viewState.alert?.attributes);
	const roster = $derived(attrs?.roster);
</script>

<div class="flex gap-2">
	<div class="flex flex-col gap-2 min-w-96 max-w-xl">
		<div class="flex flex-col gap-2 border p-2 h-fit">
			<span class="uppercase font-semibold text-surface-content/90">Description</span>
			{#if !!attrs && !attrs?.description}
				<span class="text-surface-content/60">No Description Provided</span>
			{:else}
				<span>{attrs?.description}</span>
			{/if}
		</div>

		{#if !!attrs?.definition}
			<div class="flex flex-col gap-2 border p-2 h-fit">
				<span class="uppercase font-semibold text-surface-content/90">Definition</span>
				<span>{attrs?.definition}</span>
			</div>
		{/if}
		
		{#if !!roster}
			<div class="flex flex-col gap-2 border p-2 h-fit">
				<span class="uppercase font-semibold text-surface-content/90">Roster</span>
				
				<a href="/rosters/{roster.id}" class="flex items-center gap-2 bg-surface-100 rounded-lg hover:bg-accent-800/40 p-1 px-3">
					<Avatar kind="roster" size={24} id={roster.id} />
					<span class="text-lg">{roster.attributes?.name}</span>
				</a>
			</div>
		{/if}
	</div>

	<div class="flex flex-col gap-2 border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Metrics</span>

		<AlertMetrics />
	</div>
</div>
