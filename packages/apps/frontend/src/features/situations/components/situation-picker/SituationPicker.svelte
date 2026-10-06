<script lang="ts">
	import { Input } from "$components/ui/input";
	import * as RadioGroup from "$components/ui/radio-group";
	import { Skeleton } from "$components/ui/skeleton";
	import { Button } from "$components/ui/button";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import { SituationPickerController } from "./controller.svelte";

	type Props = {
		/** The situation the choice is made for; never offered. */
		excludeId: string | undefined;
		value: string | undefined;
		onchange: (id: string) => void;
		label: string;
	};

	let { excludeId, value, onchange, label }: Props = $props();

	const controller = new SituationPickerController(() => excludeId);
	const id = $props.id();
</script>

<div class="flex min-w-0 flex-col gap-2">
	<label for={`${id}-search`} class="text-sm font-medium">{label}</label>
	<Input
		id={`${id}-search`}
		type="search"
		placeholder="Search raised and watching situations"
		value={controller.search}
		oninput={(event) => controller.setSearch(event.currentTarget.value)}
	/>

	<div class="overflow-hidden rounded-lg border bg-card">
		<div class="max-h-64 overflow-y-auto">
			{#if controller.isError}
				<div role="alert" class="flex flex-wrap items-center gap-3 p-3 text-sm">
					<span>Could not load situations.</span>
					<Button variant="outline" size="sm" onclick={controller.retry}>Retry</Button>
				</div>
			{:else if controller.isLoading}
				<div class="flex flex-col gap-2 p-3" aria-label="Loading situations">
					<Skeleton class="h-4 w-3/5" />
					<Skeleton class="h-4 w-2/5" />
				</div>
			{:else if controller.options.length}
				<RadioGroup.Root
					value={value ?? ""}
					onValueChange={onchange}
					aria-label={label}
					class="flex flex-col gap-0 divide-y"
				>
					{#each controller.options as option (option.id)}
						<label
							class="flex min-h-12 min-w-0 cursor-pointer items-center gap-3 px-3 py-1.5 hover:bg-accent has-data-[state=checked]:bg-accent"
						>
							<RadioGroup.Item value={option.id} />
							<span class="min-w-0 flex-1 text-sm wrap-anywhere">{option.title}</span>
							<StatusBadge status={option.status} variant="inline" />
						</label>
					{/each}
				</RadioGroup.Root>
			{:else}
				<p class="p-3 text-sm text-muted-foreground">No other raised or watching situations.</p>
			{/if}
		</div>
	</div>
</div>
