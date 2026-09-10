<script lang="ts">
	import RiArrowDownSLine from "remixicon-svelte/icons/arrow-down-s-line";
	import RiSlideshowLine from "remixicon-svelte/icons/slideshow-line";
	import { Button } from "$components/ui/button";
</script>

<div class="flex gap-2">
	<Button color="default">
		Present
		<RiSlideshowLine class="" aria-hidden="true" />
	</Button>

	<!-- <Button variant="fill-light" color="default">
		Schedule Review Meeting
		<RiCalendarLine class="" aria-hidden="true" />
	</Button>

	<Button variant="fill-light" color="default">
		More Actions
		<RiArrowDownSLine class="" aria-hidden="true" />
	</Button> -->
</div>
