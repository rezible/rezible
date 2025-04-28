<script lang="ts">
	import { Icon, Header } from "svelte-ux";
	import { mdiGhostOutline } from "@mdi/js";
</script>

<div class="flex flex-col h-fit border rounded-lg overflow-y-auto">
	<Header title="Relevant & Upcoming" subheading="Items you may be interested in" classes={{root: "p-2 w-full"}} />

	<div class="flex flex-col gap-2 overflow-y-auto border-t p-2">
		<div class="grid place-items-center">
			<div class="flex gap-2 items-center text-surface-content/80">
				<span class="leading-none">Nothing Found</span>
				<Icon data={mdiGhostOutline} />
			</div>
		</div>
	</div>
</div>
