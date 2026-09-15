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
	};
	let { value = $bindable(""), id, disabled = false }: Props = $props();

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
	const handleValueChange = (nextValue: string) => {
		// value = nextValue === CLEAR_TIMEZONE_VALUE ? "" : nextValue;
		// open = false;
	};

	let highlightedZone = $state(value || localZone);
	const handleOpenChange = (nextOpen: boolean) => {
		// open = nextOpen;
		// if (nextOpen) {
		// 	search = "";
		// 	highlightedZone = value || localZone;
		// }
	};


	let triggerRef = $state<HTMLButtonElement>(null!);
	function closeAndFocusTrigger() {
		open = false;
		tick().then(() => {
			triggerRef.focus();
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

<Popover.Root bind:open>
  <Popover.Trigger bind:ref={triggerRef}>
    {#snippet child({ props })}
      <Button
        variant="outline"
        class="w-[200px] justify-between"
        {...props}
        role="combobox"
        aria-expanded={open}
      >
		{triggerLabel}
        <!-- <ChevronsUpDownIcon class="ms-2 size-4 shrink-0 opacity-50" /> -->
      </Button>
    {/snippet}
  </Popover.Trigger>

  <Popover.Content class="w-[200px] p-0">
    <Command.Root>
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
					<Command.Item value={CLEAR_TIMEZONE_VALUE} onSelect={() => {value = ""}}>
						<RiCloseLine /> Clear timezone
					</Command.Item>
				{/if}
				{#each visibleZones as zone (zone)}
					<Command.Item value={zone} onSelect={() => {value = zone}}>
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