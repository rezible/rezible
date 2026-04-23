<script lang="ts">
	import InlineAlert from "$components/inline-alert/InlineAlert.svelte";
	import { Button } from "$components/ui/button";
	import * as Dialog from "$components/ui/dialog";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import * as Select from "$components/ui/select";
	import { Textarea } from "$components/ui/textarea";
	import { Checkbox } from "$components/ui/checkbox";
	import type { ErrorModel } from "$lib/api";
	import { useIncidentCreateDialog } from "./controller.svelte";

	const controller = useIncidentCreateDialog();

	const metadataError = $derived(controller.metadataQuery.error as ErrorModel | undefined);

	const titleError = $derived(!!controller.form.title ? controller.fieldErrors.title?.[0] : null);
</script>


<form
    class="flex max-h-[calc(100dvh-2rem)] flex-col"
    onsubmit={(event) => {
        event.preventDefault();
        controller.submit();
    }}
>
    <div class="flex-1 space-y-4 overflow-y-auto px-4 py-4">
        <InlineAlert error={metadataError ?? controller.error} />

        <div class="grid gap-4 md:grid-cols-2">
            <div class="space-y-2 md:col-span-2">
                <Label for="incident-title">Title</Label>
                <Input
                    id="incident-title"
                    placeholder="Briefly describe the incident"
                    bind:value={controller.form.title}
                    aria-invalid={!!titleError}
                />
                {#if !!titleError}
                    <p class="text-destructive text-xs" role="alert">{titleError}</p>
                {/if}
            </div>

            <div class="space-y-2">
                <Label for="incident-severity">Severity</Label>
                <Select.Root
                    type="single"
                    name="incident-severity"
                    bind:value={controller.form.severityId}
                    disabled={controller.metadataQuery.isLoading || controller.severities.length === 0}
                >
                    <Select.Trigger id="incident-severity" class="w-full">
                        {controller.getSeverityLabel(controller.form.severityId)}
                    </Select.Trigger>
                    <Select.Content>
                        {#each controller.severities as severity}
                            <Select.Item value={severity.id} label={severity.attributes.name}>
                                {severity.attributes.name}
                            </Select.Item>
                        {/each}
                    </Select.Content>
                </Select.Root>
                {#if controller.fieldErrors.severityId?.[0]}
                    <p class="text-destructive text-xs" role="alert">
                        {controller.fieldErrors.severityId[0]}
                    </p>
                {/if}
            </div>

            <div class="space-y-2">
                <Label for="incident-type">Type</Label>
                <Select.Root
                    type="single"
                    name="incident-type"
                    bind:value={controller.form.typeId}
                    disabled={controller.metadataQuery.isLoading || controller.types.length === 0}
                >
                    <Select.Trigger id="incident-type" class="w-full">
                        {controller.getTypeLabel(controller.form.typeId)}
                    </Select.Trigger>
                    <Select.Content>
                        {#each controller.types as type}
                            <Select.Item value={type.id} label={type.attributes.name}>
                                {type.attributes.name}
                            </Select.Item>
                        {/each}
                    </Select.Content>
                </Select.Root>
                {#if controller.fieldErrors.typeId?.[0]}
                    <p class="text-destructive text-xs" role="alert">{controller.fieldErrors.typeId[0]}</p>
                {/if}
            </div>
        </div>

        <div class="space-y-2">
            <Label for="incident-summary">Summary</Label>
            <Textarea
                id="incident-summary"
                placeholder="Optional incident summary"
                bind:value={controller.form.summary}
            />
        </div>

        {#if controller.tags.length > 0}
            <div class="space-y-2">
                <div class="space-y-1">
                    <h3 class="font-medium">Tags</h3>
                    <p class="text-muted-foreground text-xs">Optional tags to attach at incident creation.</p>
                </div>

                <div class="grid gap-2 sm:grid-cols-2">
                    {#each controller.tags as tag}
                        <label class="border-border flex items-center gap-2 border px-2 py-2 text-xs">
                            <Checkbox
                                bind:checked={() => controller.hasTag(tag.id), () =>
                                    controller.toggleTag(tag.id)}
                                aria-label={`Select tag ${tag.attributes.value}`}
                            />
                            <span>{tag.attributes.value}</span>
                        </label>
                    {/each}
                </div>
            </div>
        {/if}

        {#if controller.fields.length > 0}
            <div class="space-y-3">
                <div class="space-y-1">
                    <h3 class="font-medium">Custom Fields</h3>
                    <p class="text-muted-foreground text-xs">
                        Optional metadata fields backed by incident field options.
                    </p>
                </div>

                <div class="grid gap-4 md:grid-cols-2">
                    {#each controller.fields as field}
                        <div class="space-y-2">
                            <Label for={`incident-field-${field.id}`}>{field.attributes.name}</Label>
                            <Select.Root
                                type="single"
                                name={`incident-field-${field.id}`}
                                bind:value={() => controller.getFieldSelection(field.id), (value) =>
                                    controller.setFieldSelection(field.id, value)}
                            >
                                <Select.Trigger id={`incident-field-${field.id}`} class="w-full">
                                    {controller.getFieldSelectionLabel(field)}
                                </Select.Trigger>
                                <Select.Content>
                                    <Select.Item value="" label="None">None</Select.Item>
                                    {#each field.attributes.options as option}
                                        <Select.Item value={option.id} label={option.attributes.value}>
                                            {option.attributes.value}
                                        </Select.Item>
                                    {/each}
                                </Select.Content>
                            </Select.Root>
                        </div>
                    {/each}
                </div>
            </div>
        {/if}
    </div>

    <div class="border-t px-4 py-3">
        <Dialog.Footer>
            <Button type="button" variant="outline" onclick={() => controller.setOpen(false)}>
                Cancel
            </Button>
            <Button type="submit" disabled={!controller.canSubmit}>
                {controller.isPending ? "Creating..." : "Create Incident"}
            </Button>
        </Dialog.Footer>
    </div>
</form>