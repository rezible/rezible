<script lang="ts">
	import { Button } from "svelte-ux";
	import type { OncallRoster } from "./types";
	import type { OncallRosterViewRouteParam } from "$src/params/oncallRosterView";

	type Props = {
		roster?: OncallRoster;
	}
	const { roster }: Props = $props();

</script>


<div class="flex flex-row gap-2">
	<Button variant="fill-light">Edit Roster</Button>
</div>