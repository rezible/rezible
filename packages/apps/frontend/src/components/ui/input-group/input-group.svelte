<script lang="ts">
	import { cn, type WithElementRef } from "$lib/utils.js";
	import type { HTMLAttributes } from "svelte/elements";

	let { ref = $bindable(null), class: className, children, ...props }: WithElementRef<HTMLAttributes<HTMLDivElement>> = $props();
</script>

<div
	bind:this={ref}
	data-slot="input-group"
	role="group"
	class={cn(
		"group/input-group border-input bg-card dark:bg-input/30 has-[[data-slot=input-group-control]:focus-visible]:border-ring has-[[data-slot=input-group-control]:focus-visible]:ring-ring/50 has-[[data-slot][aria-invalid=true]]:border-destructive has-disabled:bg-muted h-9 rounded-md border transition-colors has-[[data-slot=input-group-control]:focus-visible]:ring-1 relative flex w-full min-w-0 items-center outline-none has-[>textarea]:h-auto",
		className
	)}
	{...props}
>
	{@render children?.()}
</div>
