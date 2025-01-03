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

<div class="flex flex-col gap-2">
	<Button classes={{root: "px-8 justify-self-center"}} variant="fill-light" color="secondary" on:click={() => {showDebriefDialog = true}}>
		Open Debrief
	</Button>

	<div class="divider w-full border-b"></div>

	<!--LoadingQueryWrapper query={debriefSuggestionsQuery}>
		{#snippet view(suggestions: RetrospectiveDiscussion[])}

		{/snippet}
	</LoadingQueryWrapper-->
</div>