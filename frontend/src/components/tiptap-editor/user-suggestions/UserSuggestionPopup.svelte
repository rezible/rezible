<script lang="ts">
	import type { SuggestionKeyDownProps, SuggestionProps } from "@tiptap/suggestion";
	import { cls } from '@layerstack/tailwind';

	const { items, command }: SuggestionProps<string, any> = $props();

	let highlighted = $state(0);
	$effect(() => {
		if (items.length) highlighted = 0;
	});

	const acceptUser = (idx: number) => {
		const item = items[idx];
		command({ id: "bleh", label: item });
	};

	const handleUp = () => {
		highlighted = (highlighted + items.length - 1) % items.length;
		return true;
	};
	const handleDown = () => {
		highlighted = (highlighted + 1) % items.length;
		return true;
	};
	const handleEnter = () => {
		if (items.length == 0) return true;
		if (highlighted >= 0 && highlighted < items.length) {
			acceptUser(highlighted);
		}
		return true;
	};

	export const onKeyDown = (props: SuggestionKeyDownProps): boolean => {
		switch (props.event.key) {
			case "ArrowUp":
				return handleUp();
			case "ArrowDown":
				return handleDown();
			case "Enter":
				return handleEnter();
			default:
				return false;
		}
	};
</script>

<div class="flex flex-col gap-1 border bg-surface-100">
	{#if items.length === 0}
		<div class="mx-2">No result</div>
	{:else}
		{#each items as item, i}
			<button
				onclick={() => acceptUser(i)}
				class={cls("px-2", highlighted === i ? "bg-accent text-accent-content" : "")}
			>
				{item}
			</button>
		{/each}
	{/if}
</div>
