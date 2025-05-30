<script lang="ts">
	import { cls } from "@layerstack/tailwind";
	import { uniqueId } from "@layerstack/utils";

	type Classes = {
		root?: string;
		path?: string | string[];
	};

	type Props = {
		data: string;

		size?: string | number;
		viewBox?: string;

		title?: string;
		titleId?: string;
		desc?: string;
		descId?: string;

		classes?: Classes;
	};
	const { data, size = "1.5em", viewBox = "0 0 24 24", title, desc, classes = {} }: Props = $props();

	const isLabelled = $derived(title || desc);
	const titleId = $derived(!!title ? uniqueId("title-") : "");
	const descId = $derived(!!title ? uniqueId("desc-") : "");
</script>

<svg
	width={size}
	height={size}
	{viewBox}
	class={cls("inline-block flex-shrink-0 fill-current", classes.root)}
	role={isLabelled ? "img" : "presentation"}
	aria-labelledby={isLabelled ? `${titleId} ${descId}` : undefined}
>
	{#if title}<title id={titleId}>{title}</title>{/if}
	{#if desc}<desc id={descId}>{desc}</desc>{/if}

	{#each Array.isArray(data) ? data : [data] as d, i}
		<path
			{d}
			fill="currentColor"
			class={cls(Array.isArray(classes.path) ? classes.path[i] : classes.path)}
		/>
	{/each}
</svg>
