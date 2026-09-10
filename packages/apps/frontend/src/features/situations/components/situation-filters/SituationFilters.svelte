<script lang="ts">
	import { Input } from "$components/ui/input";
	import * as Select from "$components/ui/select";
	import { situationStatusOptions, type SituationFilters } from "../../lib/filters";

	type Props = {
		filters: SituationFilters;
		onchange: (values: Partial<SituationFilters>) => void;
	};

	let { filters, onchange }: Props = $props();

	const id = $props.id();
</script>

<div class="flex flex-wrap items-end gap-2">
	<div class="min-w-0 flex-1 basis-40">
		<label for={`${id}-search`} class="mb-1 block text-xs text-muted-foreground">
			Search situation titles
		</label>
		<Input
			id={`${id}-search`}
			type="search"
			value={filters.search}
			oninput={(event) => onchange({ search: event.currentTarget.value })}
		/>
	</div>
	<div class="min-w-0">
		<label for={`${id}-status`} class="mb-1 block text-xs text-muted-foreground">Situation status</label>
		<Select.Root
			type="single"
			value={filters.status}
			onValueChange={(value) => onchange({ status: value as SituationFilters["status"] })}
		>
			<Select.Trigger id={`${id}-status`} class="w-full">
				{situationStatusOptions.find((option) => option.value === filters.status)?.label}
			</Select.Trigger>
			<Select.Content>
				{#each situationStatusOptions as option (option.value)}
					<Select.Item value={option.value}>{option.label}</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>
	</div>
</div>
