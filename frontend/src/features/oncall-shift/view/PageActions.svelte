<script lang="ts">
	import { Button } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { mdiChevronLeft, mdiChevronRight } from "@mdi/js";
	import { page } from "$app/state";

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