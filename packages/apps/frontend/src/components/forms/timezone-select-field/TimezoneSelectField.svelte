<script lang="ts">
	import { Button } from "$src/components/ui/button";
	import * as Command from "$src/components/ui/command";
	import * as Popover from "$src/components/ui/popover";
	import RiCheckLine from "remixicon-svelte/icons/check-line";
	import RiCloseLine from "remixicon-svelte/icons/close-line";
	import { tick } from "svelte";

	type Props = {
		value?: string;
		id?: string;
		disabled?: boolean;
		"aria-invalid"?: boolean;
		"aria-describedby"?: string;
	};
	let { value = $bindable(""), id, disabled = false, ...aria }: Props = $props();

	const CLEAR_TIMEZONE_VALUE = "__clear__";
	const localZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
	const browserZones = Intl.supportedValuesOf("timeZone");

	let search = $state("");
	const searchQuery = $derived(search.toLowerCase());

	function formatTimezone(zone: string) {
		return zone.replaceAll("_", " ");
	}

	function matchesSearch(zone: string) {
		const searchableText = `${zone} ${formatTimezone(zone)}`.toLowerCase();
		return searchableText.includes(searchQuery);
	}

	let open = $state(false);

	let triggerRef = $state<HTMLButtonElement>(null!);
	function closeAndFocusTrigger() {
		open = false;
		tick().then(() => {
			triggerRef?.focus();
		});
	}

	const zones = $derived.by(() => {
		// Include saved and local zones even if the browser does not list them.
		const availableZones = ["UTC", localZone, value, ...browserZones].filter(Boolean);
		return [...new Set(availableZones)].sort();
	});
	const visibleZones = $derived(zones.filter(matchesSearch));
	const showClear = $derived(Boolean(value) && "clear timezone".includes(searchQuery));
	const triggerLabel = $derived(value ? formatTimezone(value) : "Select timezone");
</script>

<Popover.Root
	bind:open
	onOpenChange={(nextOpen) => {
		if (nextOpen) search = "";
	}}
>
	<Popover.Trigger bind:ref={triggerRef} {id} {disabled}>
		{#snippet child({ props })}
			<Button
				variant="outline"
				class="w-full min-w-0 justify-between"
				{...props}
				{...aria}
				role="combobox"
				aria-expanded={open}
			>
				<span class="truncate">{triggerLabel}</span>
			</Button>
		{/snippet}
	</Popover.Trigger>

	<Popover.Content class="w-[var(--bits-popover-anchor-width)] p-0">
		<Command.Root shouldFilter={false}>
			<Command.Input
				oninput={(event) => (search = event.currentTarget.value)}
				placeholder="Search timezones…"
			/>
			<Command.List>
				{#if !showClear && visibleZones.length === 0}
					<div class="text-muted-foreground py-6 text-center text-sm">No timezone found.</div>
				{:else}
					<Command.Group>
						{#if showClear}
							<Command.Item
								value={CLEAR_TIMEZONE_VALUE}
								onSelect={() => {
									value = "";
									closeAndFocusTrigger();
								}}
							>
								<RiCloseLine /> Clear timezone
							</Command.Item>
						{/if}
						{#each visibleZones as zone (zone)}
							<Command.Item
								value={zone}
								onSelect={() => {
									value = zone;
									closeAndFocusTrigger();
								}}
							>
								{formatTimezone(zone)}
								{#if value === zone}
									<RiCheckLine class="ml-auto" />
								{/if}
							</Command.Item>
						{/each}
					</Command.Group>
				{/if}
			</Command.List>
		</Command.Root>
	</Popover.Content>
</Popover.Root>
