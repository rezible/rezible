<script lang="ts">
    import DiscussionSidebar from "./discussions/DiscussionSidebar.svelte";

	type Props = {
		debriefId: string;
		retrospectiveId: string;
		showDebriefDialog: boolean;
	}
	let { debriefId, retrospectiveId, showDebriefDialog = $bindable() }: Props = $props();
</script>

<DiscussionSidebar
    bind:showDebriefDialog
    {retrospectiveId} 
    {debriefId} />