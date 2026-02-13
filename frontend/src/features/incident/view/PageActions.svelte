<script lang="ts">
	import { mdiChevronDown, mdiCalendar, mdiPresentationPlay } from "@mdi/js";
	import { Button } from "$components/ui/button";
	import Icon from "$components/icon/Icon.svelte";
    
</script>

<div class="flex gap-2">
	<Button color="default">
		Present
		<Icon data={mdiPresentationPlay} />
	</Button>

	<!-- <Button variant="fill-light" color="default">
		Schedule Review Meeting
		<Icon data={mdiCalendar} />
	</Button>

	<Button variant="fill-light" color="default">
		More Actions
		<Icon data={mdiChevronDown} />
	</Button> -->
</div>