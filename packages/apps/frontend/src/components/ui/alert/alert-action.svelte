<script lang="ts">
	import type { HTMLAttributes } from "svelte/elements";
	import { cn, type WithElementRef } from "$lib/utils.js";

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLDivElement>> = $props();
</script>

<div
	bind:this={ref}
	data-slot="alert-action"
	class={cn("absolute top-[calc(--spacing(1.25))] right-[calc(--spacing(1.25))]", className)}
	{...restProps}
>
	{@render children?.()}
</div>
