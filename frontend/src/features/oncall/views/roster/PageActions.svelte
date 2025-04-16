<script lang="ts">
	import { Button } from "svelte-ux";
	import type { RosterViewState } from "./context.svelte";

	type Props = {
		viewState: RosterViewState;
	}
	const { viewState }: Props = $props();

</script>


<div class="flex flex-row gap-2">
	<Button variant="fill-light">Edit Roster</Button>
</div>