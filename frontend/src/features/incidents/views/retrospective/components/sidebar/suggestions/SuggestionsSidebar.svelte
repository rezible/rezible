<script lang="ts">
    import DebriefBox from "./DebriefBox.svelte";

	type Props = {
		debriefId: string;
		retrospectiveId: string;
		showDebriefDialog: boolean;
	}
	let { debriefId, retrospectiveId, showDebriefDialog = $bindable() }: Props = $props();
</script>

<div class="flex flex-col gap-2 overflow-y-auto">
	<DebriefBox {debriefId} bind:showDebriefDialog />
</div>
