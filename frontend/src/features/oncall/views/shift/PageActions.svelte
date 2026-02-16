<script lang="ts">
	import { Button } from "$components/ui/button";
	import Icon from "$components/icon/Icon.svelte";
	import { mdiChevronLeft, mdiChevronRight } from "@mdi/js";

	type Props = {
		previousId?: string;
		nextId?: string;
	};
	const { previousId, nextId }: Props = $props();
</script>

<div class="flex gap-2">
	{#if previousId}
		<Button href="/shifts/{previousId}">
			<Icon data={mdiChevronLeft} />
			<span class="">Previous Shift</span>
		</Button>
	{/if}

	{#if nextId}
		<Button href="/shifts/{nextId}">
			<span class="">Next Shift</span>
			<Icon data={mdiChevronRight} />
		</Button>
	{/if}
</div>