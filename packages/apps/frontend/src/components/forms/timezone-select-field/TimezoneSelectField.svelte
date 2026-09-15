<script lang="ts">
	import { Combobox } from "bits-ui";
	import RiCheckLine from "remixicon-svelte/icons/check-line";
	import RiCloseLine from "remixicon-svelte/icons/close-line";

	type Props = {
		value?: string;
		id?: string;
		disabled?: boolean;
	};

	let { value = $bindable(""), id, disabled = false }: Props = $props();

	const CLEAR_TIMEZONE_VALUE = "__clear__";
	const localZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
	const browserZones =
		typeof Intl.supportedValuesOf === "function" ? Intl.supportedValuesOf("timeZone") : [];

	let open = $state(false);
	let search = $state("");
	let highlightedZone = $state(value || localZone);

	const zones = $derived.by(() => {
		// Include saved and local zones even if the browser does not list them.
		const availableZones = ["UTC", localZone, value, ...browserZones].filter(Boolean);
		return [...new Set(availableZones)].sort();
	});
	const searchQuery = $derived(search.toLowerCase());
	const visibleZones = $derived(zones.filter(matchesSearch));
	const showClear = $derived(Boolean(value) && "clear timezone".includes(searchQuery));
	const triggerLabel = $derived(value ? formatTimezone(value) : "Select timezone");

	function formatTimezone(zone: string) {
		return zone.replaceAll("_", " ");
	}

	function matchesSearch(zone: string) {
		const searchableText = `${zone} ${formatTimezone(zone)}`.toLowerCase();
		return searchableText.includes(searchQuery);
	}

	const handleValueChange = (nextValue: string) => {
		value = nextValue === CLEAR_TIMEZONE_VALUE ? "" : nextValue;
		open = false;
	};

	const handleOpenChange = (nextOpen: boolean) => {
		open = nextOpen;
		if (nextOpen) {
			search = "";
			highlightedZone = value || localZone;
		}
	};
</script>

<Combobox.Root
	bind:open
	bind:value={highlightedZone}
	onValueChange={handleValueChange}
	onOpenChange={handleOpenChange}
	inputValue={search}
	type="single"
>
	<Combobox.Trigger
		{id}
		class="border-input bg-card text-foreground hover:bg-accent flex h-9 w-full items-center justify-between rounded-md border px-3 text-sm font-normal"
		{disabled}
		aria-label="Timezone"
	>
		{triggerLabel}
	</Combobox.Trigger>
	<Combobox.Content class="w-[var(--bits-combobox-trigger-width)] p-0">
		<Combobox.Input
			oninput={(event) => (search = event.currentTarget.value)}
			placeholder="Search timezones…"
		/>
		<Combobox.Viewport class="max-h-80 overflow-y-auto p-1">
			{#if !showClear && visibleZones.length === 0}
				<div class="text-muted-foreground py-6 text-center text-sm">No timezone found.</div>
			{:else}
				<Combobox.Group>
					{#if showClear}
						<Combobox.Item value={CLEAR_TIMEZONE_VALUE}>
							<RiCloseLine /> Clear timezone
						</Combobox.Item>
					{/if}
					{#each visibleZones as zone (zone)}
						<Combobox.Item value={zone}>
							{formatTimezone(zone)}
							{#if value === zone}
								<RiCheckLine class="ml-auto" />
							{/if}
						</Combobox.Item>
					{/each}
				</Combobox.Group>
			{/if}
		</Combobox.Viewport>
	</Combobox.Content>
</Combobox.Root>
