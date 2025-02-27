<script lang="ts">
	import { mdiChevronDown, mdiCalendar } from "@mdi/js";
	import { Button, Icon } from "svelte-ux";

    
</script>

<div class="flex gap-2">
	<Button variant="fill-light" color="default">
		Schedule Review Meeting
		<Icon data={mdiCalendar} />
	</Button>

	<Button variant="fill-light" color="default">
		More Actions
		<Icon data={mdiChevronDown} />
	</Button>
</div>