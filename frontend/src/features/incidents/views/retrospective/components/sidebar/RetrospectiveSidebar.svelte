<script lang="ts">
    import { cls } from "svelte-ux";
	import DiscussionSidebar from "./discussion/DiscussionSidebar.svelte";
    import SuggestionsSidebar from "./suggestions/SuggestionsSidebar.svelte";

	type Props = {
		incidentId: string;
		debriefId: string;
		retrospectiveId: string;
		showDebriefDialog: boolean;
	}
	let { 
		incidentId, 
		debriefId, 
		retrospectiveId,
		showDebriefDialog = $bindable(),
	}: Props = $props();

	const tabs = [
		{key: "discussion", label: "Discussion"},
		{key: "suggestions", label: "Suggestions"},
	];
	let activeTabIdx = $state(0);
	const activeTab = $derived(tabs[activeTabIdx]);
</script>

<div class="col-span-3 grid grid-cols-1 grid-rows-[40px_minmax(0,_1fr)] min-h-0 bg-surface-100/50">
	<div class="w-full row-start-1 col-start-1 border-b-2"></div>

	<ul class="row-start-1 col-start-1 flex space-y-0 -mb-px w-fit h-10">
		{#each tabs as tab, i}
			{@const active = activeTabIdx === i}
			<li class="group flex cursor-pointer" role="presentation">
				<button 
					onclick={() => {console.log("bleh"); activeTabIdx = i}}
					class={cls(
					"inline-block pt-1 px-4 gap-2 text-lg text-center border-b-2 text-surface-content", 
					active 
						? "border-secondary bg-surface-100 active" 
						: "border-transparent hover:border-secondary/50"
				)}>
					{tab.label}
				</button>
			</li>
		{/each}
	</ul>

	<div class="grid min-h-0 overflow-y-auto p-2">
		{#if activeTab.key === "discussion"}
			<DiscussionSidebar {retrospectiveId} />
		{:else if activeTab.key === "suggestions"}
			<SuggestionsSidebar
				bind:showDebriefDialog
				{retrospectiveId} 
				{debriefId}
			/>
		{/if}
	</div>
</div>