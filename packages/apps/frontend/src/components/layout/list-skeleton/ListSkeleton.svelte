<script lang="ts">
	import { Skeleton } from "$components/ui/skeleton";

	type Props = {
		/** Accessible label, e.g. "Loading findings". */
		label: string;
		rows?: number;
	};

	const { label, rows = 3 }: Props = $props();
</script>

<ul class="divide-y rounded-lg border bg-card" aria-label={label}>
	{#each { length: rows }, row (row)}
		<li class="flex flex-col gap-2 px-4 py-3">
			<Skeleton class="h-4 w-3/5" />
			<Skeleton class="h-3 w-4/5" />
		</li>
	{/each}
</ul>
