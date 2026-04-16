<script lang="ts">
	import { Input } from "$components/ui/input";
	import { Button } from "$components/ui/button";
	import { Badge } from "$components/ui/badge";
	import { Label } from "$components/ui/label";
	import { Separator } from "$components/ui/separator";
	import { useIncidentsListView } from "./controller.svelte";

    const controller = useIncidentsListView();
</script>

    <div class="flex flex-col gap-2">
        <Label for="incident-search">Search</Label>
        <Input
            id="incident-search"
            placeholder="Title, summary, or slug"
            bind:value={controller.searchValue}
        />
    </div>

    <Separator />

    <div class="flex flex-col gap-2">
        <Label>Archive scope</Label>
        <div class="grid grid-cols-3 gap-1">
            {#each controller.archiveScopeOptions as option}
                <Button
                    variant={controller.archiveScope === option.value ? "secondary" : "outline"}
                    size="sm"
                    onclick={() => (controller.archiveScope = option.value)}
                >
                    {option.label}
                </Button>
            {/each}
        </div>
    </div>

    <div class="flex flex-col gap-2">
        <Label for="incident-status">Status</Label>
        <select
            id="incident-status"
            class="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
            bind:value={controller.statusFilter}
        >
            <option value="all">Any status</option>
            {#each controller.statusOptions as option}
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
                bind:value={controller.severityFilter}
                disabled={controller.severityOptions.length === 0}
            >
                <option value="all">Any severity</option>
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
                bind:value={controller.typeFilter}
                disabled={controller.typeOptions.length === 0}
            >
                <option value="all">Any type</option>
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
                bind:value={controller.tagFilter}
                disabled={controller.tagOptions.length === 0}
            >
                <option value="all">Any tag</option>
                {#each controller.tagOptions as option}
                    <option value={option.value}>{option.label}</option>
                {/each}
            </select>
        </div>
    </div>

    <div class="flex flex-col gap-2">
        <Label>Visibility</Label>
        <div class="grid grid-cols-3 gap-1">
            {#each controller.visibilityOptions as option}
                <Button
                    variant={controller.visibilityFilter === option.value ? "secondary" : "outline"}
                    size="sm"
                    onclick={() => (controller.visibilityFilter = option.value)}
                >
                    {option.label}
                </Button>
            {/each}
        </div>
    </div>

    {#if controller.activeFilterCount > 0}
        <div class="flex flex-wrap gap-1">
            <Badge variant="outline">{controller.filteredIncidents.length} matches</Badge>
            {#if controller.activeStatusLabel}
                <Badge variant="secondary">{controller.activeStatusLabel}</Badge>
            {/if}
            {#if controller.visibilityFilter !== "all"}
                <Badge variant="secondary">{controller.visibilityFilter}</Badge>
            {/if}
        </div>
    {/if}
