<script lang="ts" module>
	export type FilterOption<V extends string> = {
		value: V;
		label: string;
		/** Matching results for this option. Omit while unknown. */
		count?: number;
	};
</script>

<script lang="ts" generics="T extends string">
	type Props = {
		label: string;
		options: readonly FilterOption<T>[];
		value: T;
		onchange: (value: T) => void;
	};

	const { label, options, value, onchange }: Props = $props();

	const id = $props.id();
</script>

<fieldset class="flex min-w-0 flex-col gap-0.5">
	<legend class="region-label mb-1.5">{label}</legend>
	{#each options as option (option.value)}
		<label
			class="relative flex h-9 cursor-pointer items-center justify-between gap-2 rounded-md px-3 text-sm text-muted-foreground transition-colors duration-150 after:absolute after:inset-y-2 after:left-0 after:w-0.5 after:rounded-full after:bg-brand after:opacity-0 hover:bg-muted hover:text-foreground has-checked:bg-selection has-checked:font-medium has-checked:text-selection-foreground has-checked:after:opacity-100 has-focus-visible:outline-2 has-focus-visible:outline-offset-2 has-focus-visible:outline-ring"
		>
			<input
				type="radio"
				class="sr-only"
				name={id}
				value={option.value}
				checked={option.value === value}
				onchange={() => onchange(option.value)}
			/>
			<span class="min-w-0 truncate">{option.label}</span>
			{#if option.count !== undefined}
				<span class="text-xs font-normal text-muted-foreground tabular-nums">{option.count}</span>
			{/if}
		</label>
	{/each}
</fieldset>
