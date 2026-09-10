<script lang="ts">
	import { cn, type WithElementRef } from "$lib/utils.js";
	import type { HTMLAttributes } from "svelte/elements";

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLTableRowElement>> = $props();
</script>

<tr
	bind:this={ref}
	data-slot="table-row"
	class={cn(
		"hover:bg-accent data-[state=selected]:bg-selection data-[state=selected]:text-selection-foreground data-[state=selected]:hover:bg-selection data-[state=selected]:hover:text-selection-foreground border-b h-11 transition-colors",
		className
	)}
	{...restProps}
>
	{@render children?.()}
</tr>
