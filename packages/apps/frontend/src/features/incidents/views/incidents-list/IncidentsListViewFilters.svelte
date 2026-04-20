<script lang="ts">
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
    import * as Select from "$components/ui/select";
	import { Separator } from "$components/ui/separator";
	import { incidentStatusOptions, useIncidentsListView } from "./controller.svelte";

    const controller = useIncidentsListView();
</script>

<div class="flex flex-col gap-2">
    <Label for="incident-search">Search</Label>
    <Input id="incident-search" placeholder="Title, summary, or slug"
        bind:value={controller.filters.search}
    />
</div>

<Separator />

<div class="grid grid-cols-1 gap-3">
    <div class="flex flex-col gap-2">
        <Label for="incident-status">Status</Label>
        <Select.Root type="single" name="incident-status" bind:value={controller.filters.status}>
            <Select.Trigger class="w-full">
                {controller.statusFilterLabel}
            </Select.Trigger>
            <Select.Content>
                <Select.Item value="" label="Any">Any</Select.Item>
                {#each incidentStatusOptions as option}
                    <Select.Item value={option.value} label={option.label}>
                        {option.label}
                    </Select.Item>
                {/each}
            </Select.Content>
        </Select.Root>
    </div>

    <div class="flex flex-col gap-2">
        <Label for="incident-severity">Severity</Label>
        <Select.Root type="single" name="incident-severity" bind:value={controller.filters.severity}>
            <Select.Trigger class="w-full">
                {controller.severityFilterLabel}
            </Select.Trigger>
            <Select.Content>
                <Select.Item value="" label="Any">Any</Select.Item>
                {#each controller.severityOptions as option}
                    <Select.Item {...option}>{option.label}</Select.Item>
                {/each}
            </Select.Content>
        </Select.Root>
    </div>

    <div class="flex flex-col gap-2">
        <Label for="incident-type">Type</Label>
        <Select.Root type="single" name="incident-type" bind:value={controller.filters.type}>
            <Select.Trigger class="w-full">
                {controller.typeFilterLabel}
            </Select.Trigger>
            <Select.Content>
                <Select.Item value="" label="Any">Any</Select.Item>
                {#each controller.typeOptions as option}
                    <Select.Item {...option}>{option.label}</Select.Item>
                {/each}
            </Select.Content>
        </Select.Root>
    </div>


    <div class="flex flex-col gap-2">
        <Label for="incident-tag">Tag</Label>
        <Select.Root type="single" name="incident-tag" bind:value={controller.filters.tag} disabled={!controller.tagOptions.length}>
            <Select.Trigger class="w-full">
                {controller.tagFilterLabel}
            </Select.Trigger>
            <Select.Content>
                {#each controller.tagOptions as option}
                    <Select.Item {...option}>{option.label}</Select.Item>
                {/each}
            </Select.Content>
        </Select.Root>
    </div>
</div>

