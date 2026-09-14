<script lang="ts">
	import { Select as SelectPrimitive } from "bits-ui";
	import { cn, type WithoutChild } from "$lib/utils.js";
	import RiArrowDownSLine from "remixicon-svelte/icons/arrow-down-s-line";

	let {
		ref = $bindable(null),
		class: className,
		children,
		size = "default",
		...restProps
	}: WithoutChild<SelectPrimitive.TriggerProps> & {
		size?: "sm" | "default";
	} = $props();
</script>

<SelectPrimitive.Trigger
	bind:ref
	data-slot="select-trigger"
	data-size={size}
	class={cn(
		"border-input bg-card data-placeholder:text-muted-foreground hover:bg-accent focus-visible:border-ring aria-invalid:border-destructive gap-1.5 rounded-md border py-2 pr-2 pl-2.5 text-sm transition-colors select-none focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring data-[size=default]:h-9 data-[size=sm]:h-8 *:data-[slot=select-value]:flex *:data-[slot=select-value]:gap-1.5 [&_svg:not([class*='size-'])]:size-4 flex w-fit items-center justify-between whitespace-nowrap outline-none disabled:cursor-not-allowed disabled:bg-muted disabled:text-foreground-disabled *:data-[slot=select-value]:line-clamp-1 *:data-[slot=select-value]:items-center [&_svg]:pointer-events-none [&_svg]:shrink-0",
		className
	)}
	{...restProps}
>
	{@render children?.()}
	<RiArrowDownSLine class="text-muted-foreground size-4 pointer-events-none" />
</SelectPrimitive.Trigger>
