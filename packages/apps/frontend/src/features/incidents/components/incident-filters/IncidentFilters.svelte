<script lang="ts">
	import { Input } from "$components/ui/input";
	import * as Select from "$components/ui/select";
	import { incidentStatusOptions, type IncidentFilters } from "../../lib/filters";
	import type { IncidentSeverity } from "$lib/api";

	type Props = {
		filters: IncidentFilters;
		onchange: (values: Partial<IncidentFilters>) => void;
		severities?: IncidentSeverity[];
	};

	let { filters, onchange, severities = [] }: Props = $props();

	const id = $props.id();
	const severityOptions = $derived([
		{ value: "", label: "Any severity" },
		...severities.map((s) => ({ value: s.id, label: s.attributes.name })),
	]);
</script>

<div class="flex flex-wrap items-end gap-2">
	<div class="min-w-0 flex-1 basis-40">
		<label for={`${id}-search`} class="mb-1 block text-xs text-muted-foreground">Search incidents</label>
		<Input
			id={`${id}-search`}
			type="search"
			value={filters.search}
			oninput={(event) => onchange({ search: event.currentTarget.value })}
		/>
	</div>
	<div class="min-w-0">
		<label for={`${id}-status`} class="mb-1 block text-xs text-muted-foreground">Incident status</label>
		<Select.Root
			type="single"
			value={filters.status}
			onValueChange={(value) => onchange({ status: value as IncidentFilters["status"] })}
		>
			<Select.Trigger id={`${id}-status`} class="w-full">
				{incidentStatusOptions.find((option) => option.value === filters.status)?.label}
			</Select.Trigger>
			<Select.Content>
				{#each incidentStatusOptions as option (option.value)}
					<Select.Item value={option.value}>{option.label}</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>
	</div>
	<div class="min-w-0">
		<label for={`${id}-severityId`} class="mb-1 block text-xs text-muted-foreground">Severity</label>
		<Select.Root
			type="single"
			value={filters.severityId}
			onValueChange={(value) => onchange({ severityId: value as IncidentFilters["severityId"] })}
		>
			<Select.Trigger id={`${id}-severityId`} class="w-full">
				{severityOptions.find((option) => option.value === filters.severityId)?.label}
			</Select.Trigger>
			<Select.Content>
				{#each severityOptions as option (option.value)}
					<Select.Item value={option.value}>{option.label}</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>
	</div>
</div>
