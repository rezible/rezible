<script lang="ts">
	import { startOfMonth, endOfMonth } from "date-fns";
	import { Button, Header } from "svelte-ux";
	import { mdiDotsVertical } from "@mdi/js";

</script>

<div class="h-fit flex flex-col gap-2 border rounded px-4 py-2">
	<Header>
		<span slot="title" class="text-lg">
			Incidents <span class="opacity-80">(past 30 days)</span>
		</span>
		<svelte:fragment slot="actions">
			<Button icon={mdiDotsVertical} iconOnly />
		</svelte:fragment>
	</Header>

	<div class="h-[300px] p-4 border rounded bg-surface-100">
		
	</div>
</div>
