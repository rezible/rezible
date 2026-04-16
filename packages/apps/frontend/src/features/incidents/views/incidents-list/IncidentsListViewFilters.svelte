<script lang="ts">
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import { Separator } from "$components/ui/separator";
	import { incidentStatusOptions, useIncidentsListView } from "./controller.svelte";

    const controller = useIncidentsListView();
    $inspect(controller.filters.severity);
</script>

<div class="flex flex-col gap-2">
    <Label for="incident-search">Search</Label>
    <Input
        id="incident-search"
        placeholder="Title, summary, or slug"
        bind:value={controller.filters.search}
    />
</div>

<Separator />

<div class="flex flex-col gap-2">
    <Label for="incident-status">Status</Label>
    <select
        id="incident-status"
        class="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
        bind:value={controller.filters.status}
    >
        {#each incidentStatusOptions as option}
            <option value={option.value}>{option.label}</option>
        {/each}
    </select>
</div>

<div class="grid grid-cols-1 gap-3">
    <div class="flex flex-col gap-2">
        <Label for="incident-severity">Severity</Label>
        <select
            id="incident-severity"
            class="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
            bind:value={controller.filters.severity}
            disabled={controller.severityOptions.length === 0}
        >
            <option value={undefined}>Any</option>
            {#each controller.severityOptions as option}
                <option value={option.value}>{option.label}</option>
            {/each}
        </select>
    </div>

    <div class="flex flex-col gap-2">
        <Label for="incident-type">Type</Label>
        <select
            id="incident-type"
            class="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
            bind:value={controller.filters.type}
            disabled={controller.typeOptions.length === 0}
        >
            <option value={undefined}>Any</option>
            {#each controller.typeOptions as option}
                <option value={option.value}>{option.label}</option>
            {/each}
        </select>
    </div>

    <div class="flex flex-col gap-2">
        <Label for="incident-tag">Tag</Label>
        <select
            id="incident-tag"
            class="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
            bind:value={controller.filters.tag}
            disabled={controller.tagOptions.length === 0}
        >
            <option value={undefined}>Any</option>
            {#each controller.tagOptions as option}
                <option value={option.value}>{option.label}</option>
            {/each}
        </select>
    </div>
</div>

