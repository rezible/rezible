<script lang="ts">
	import { Input } from "$components/ui/input";
	import * as Select from "$components/ui/select";
	import { eventTimeOptions, type EventFilters } from "../../lib/filters";

	type Props = {
		filters: EventFilters;
		onchange: (values: Partial<EventFilters>) => void;
	};

	let { filters, onchange }: Props = $props();

	const id = $props.id();
</script>

<div class="flex flex-wrap items-end gap-2">
	<div class="min-w-0 flex-1 basis-40">
		<label for={`${id}-kind`} class="mb-1 block text-xs text-muted-foreground">Exact event kind</label>
		<Input
			id={`${id}-kind`}
			type="search"
			value={filters.kind}
			oninput={(event) => onchange({ kind: event.currentTarget.value })}
		/>
	</div>
	<div class="min-w-0">
		<label for={`${id}-time`} class="mb-1 block text-xs text-muted-foreground">Occurred time</label>
		<Select.Root
			type="single"
			value={filters.time}
			onValueChange={(value) => onchange({ time: value as EventFilters["time"] })}
		>
			<Select.Trigger id={`${id}-time`} class="w-full">
				{eventTimeOptions.find((option) => option.value === filters.time)?.label}
			</Select.Trigger>
			<Select.Content>
				{#each eventTimeOptions as option (option.value)}
					<Select.Item value={option.value}>{option.label}</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>
	</div>
</div>
