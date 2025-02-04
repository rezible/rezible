<script lang="ts" module>
	import type { AvatarProps } from "$components/avatar/Avatar.svelte";
	export type Breadcrumb = {
		label: string;
		href?: string;
		avatar?: AvatarProps;
	};
</script>

<script lang="ts">
	import Avatar from "$components/avatar/Avatar.svelte";

	type Props = {
		crumbs: Breadcrumb[];
	};
	const { crumbs }: Props = $props();
</script>

<span
	class="text-xl text-surface-content/50 w-fit px-2 self-bottom flex gap-1 items-end"
>
	{#each crumbs as c, i}
		{@const last = i === crumbs.length - 1}
		{#if i > 0}
			<span>/</span>
		{/if}

		<span class="flex items-stretch gap-2">
			{#if c.avatar}
				<Avatar {...c.avatar} size={30} />
			{/if}

			{#if c.href}
				<a
					href={c.href}
					class:text-2xl={last}
					class:text-surface-content={last}>{c.label}</a
				>
			{:else}
				<span class:text-2xl={last} class:text-surface-content={last}
					>{c.label}</span
				>
			{/if}
		</span>
	{/each}
</span>
