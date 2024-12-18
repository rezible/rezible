<script lang="ts">
    import { createQuery } from "@tanstack/svelte-query";
    import { Button, Collapse } from "svelte-ux";
    import LoadingQueryWrapper from "$components/loader/LoadingQueryWrapper.svelte";

	type Props = {
		debriefId: string;
		showDebriefDialog: boolean;
	}
	let { debriefId, showDebriefDialog = $bindable() }: Props = $props();

	// const queryOptions = $derived(listRetrospectiveDiscussionsOptions({path: {id: retrospectiveId}}));
	// const query = createQuery(() => debriefSuggestionsQueryOptions);
</script>

<Collapse open={!showDebriefDialog} class="p-2 rounded-lg elevation-1 bg-surface-100">
	<div slot="trigger" let:open class="flex-1 px-3 py-3 {open ? "text-surface-content/50" : ""}">Post-Incident Debrief</div>
	<div class="flex flex-col">
		<div class="px-3 pb-3 text-surface-content">
			<p>Suggestions</p>
			<!--LoadingQueryWrapper query={debriefSuggestionsQuery}>
				{#snippet view(suggestions: RetrospectiveDiscussion[])}

				{/snippet}
			</LoadingQueryWrapper-->
		</div>
		<Button classes={{root: "px-8 justify-self-center"}} variant="fill-light" color="secondary" on:click={() => {showDebriefDialog = true}}>
			Open Debrief Window
		</Button>
	</div>
</Collapse>