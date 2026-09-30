<script lang="ts">
	import * as Tooltip from "$components/ui/tooltip/index.js";
	import { cn, type WithElementRef } from "$lib/utils.js";
	import type { HTMLAttributes } from "svelte/elements";
	import { SIDEBAR_WIDTH, SIDEBAR_WIDTH_ICON } from "./constants.js";
	import { setSidebar } from "./context.svelte.js";

	let {
		ref = $bindable(null),
		collapsed = false,
		class: className,
		style,
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLDivElement>> & {
		/** Collapse the desktop sidebar to the icon rail. */
		collapsed?: boolean;
	} = $props();

	setSidebar({ collapsed: () => collapsed });
</script>

<Tooltip.Provider delayDuration={0}>
	<div
		data-slot="sidebar-wrapper"
		style="--sidebar-width: {SIDEBAR_WIDTH}; --sidebar-width-icon: {SIDEBAR_WIDTH_ICON}; {style}"
		class={cn(
			"group/sidebar-wrapper has-data-[variant=inset]:bg-sidebar flex min-h-svh w-full",
			className
		)}
		bind:this={ref}
		{...restProps}
	>
		{@render children?.()}
	</div>
</Tooltip.Provider>
