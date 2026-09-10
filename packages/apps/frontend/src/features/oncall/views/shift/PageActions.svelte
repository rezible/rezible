<script lang="ts">
	import { Button } from "$components/ui/button";
	import RiArrowLeftSLine from "remixicon-svelte/icons/arrow-left-s-line";
	import RiArrowRightSLine from "remixicon-svelte/icons/arrow-right-s-line";

	type Props = {
		previousId?: string;
		nextId?: string;
	};
	const { previousId, nextId }: Props = $props();
</script>

<div class="flex gap-2">
	{#if previousId}
		<Button href="/shifts/{previousId}">
			<RiArrowLeftSLine class="" aria-hidden="true" />
			<span class="">Previous Shift</span>
		</Button>
	{/if}

	{#if nextId}
		<Button href="/shifts/{nextId}">
			<span class="">Next Shift</span>
			<RiArrowRightSLine class="" aria-hidden="true" />
		</Button>
	{/if}
</div>
