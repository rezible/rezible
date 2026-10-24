<script lang="ts">
	import RiArrowDownSLine from "remixicon-svelte/icons/arrow-down-s-line";

	import { resolve } from "$app/paths";
	import type { IncidentField, IncidentMetadata } from "$lib/api";
	import { Checkbox } from "$components/ui/checkbox";
	import * as Collapsible from "$components/ui/collapsible";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import * as Select from "$components/ui/select";
	import { Textarea } from "$components/ui/textarea";

	import type { IncidentDetailsFormState } from "./form";

	type Props = {
		/** Edited in place. */
		form: IncidentDetailsFormState;
		metadata: IncidentMetadata | undefined;
		titleError: string | undefined;
		isAdmin: boolean;
		idPrefix: string;
	};
	const { form, metadata, titleError, isAdmin, idPrefix }: Props = $props();

	const severities = $derived(
		(metadata?.severities ?? [])
			.filter((severity) => !severity.attributes.archived)
			.toSorted((a, b) => a.attributes.rank - b.attributes.rank)
	);
	const types = $derived((metadata?.types ?? []).filter((type) => !type.attributes.archived));
	const tags = $derived((metadata?.tags ?? []).filter((tag) => !tag.attributes.archived));
	const fields = $derived(metadata?.fields ?? []);
	const hasMoreDetails = $derived(tags.length > 0 || fields.length > 0);

	const severityLabel = $derived(
		severities.find((severity) => severity.id === form.severityId)?.attributes.name ?? "None"
	);
	const typeLabel = $derived(types.find((type) => type.id === form.typeId)?.attributes.name ?? "None");

	const toggleTag = (tagId: string) => {
		if (form.tagIds.includes(tagId)) {
			form.tagIds = form.tagIds.filter((id) => id !== tagId);
		} else {
			form.tagIds = [...form.tagIds, tagId];
		}
	};

	const setFieldSelection = (fieldId: string, optionId: string) => {
		const selections = { ...form.fieldSelections };
		if (optionId) {
			selections[fieldId] = optionId;
		} else {
			delete selections[fieldId];
		}
		form.fieldSelections = selections;
	};

	const fieldSelectionLabel = (field: IncidentField) => {
		const selectedId = form.fieldSelections[field.id];
		const option = field.attributes.options.find((candidate) => candidate.id === selectedId);
		return option?.attributes.value ?? "None";
	};
</script>

{#snippet notConfigured(what: string)}
	<p class="flex h-9 items-center text-sm text-muted-foreground">
		No {what} configured
		{#if isAdmin}
			<span class="px-1">·</span>
			<a class="underline underline-offset-4" href={resolve("/settings/incidents")}>
				Add them in settings
			</a>
		{/if}
	</p>
{/snippet}

<div class="flex flex-col gap-4">
	<div class="flex flex-col gap-2">
		<Label for="{idPrefix}-title">Title</Label>
		<Input
			id="{idPrefix}-title"
			placeholder="What's happening?"
			bind:value={form.title}
			aria-invalid={!!titleError}
		/>
		{#if titleError}
			<p class="text-xs text-destructive" role="alert">{titleError}</p>
		{/if}
	</div>

	<div class="flex flex-col gap-2">
		<Label for="{idPrefix}-summary">Summary</Label>
		<Textarea
			id="{idPrefix}-summary"
			rows={3}
			placeholder="What do you know so far?"
			bind:value={form.summary}
		/>
	</div>

	<div class="grid gap-4 sm:grid-cols-2">
		<div class="flex flex-col gap-2">
			<Label for="{idPrefix}-severity">Severity</Label>
			{#if severities.length}
				<Select.Root type="single" bind:value={form.severityId}>
					<Select.Trigger id="{idPrefix}-severity" class="w-full">{severityLabel}</Select.Trigger>
					<Select.Content>
						<Select.Item value="" label="None">None</Select.Item>
						{#each severities as severity (severity.id)}
							<Select.Item value={severity.id} label={severity.attributes.name}>
								<span
									class="size-2 shrink-0 rounded-full bg-muted-foreground"
									style:background-color={severity.attributes.color || undefined}
									aria-hidden="true"
								></span>
								{severity.attributes.name}
							</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>
			{:else}
				{@render notConfigured("severities")}
			{/if}
		</div>

		<div class="flex flex-col gap-2">
			<Label for="{idPrefix}-type">Type</Label>
			{#if types.length}
				<Select.Root type="single" bind:value={form.typeId}>
					<Select.Trigger id="{idPrefix}-type" class="w-full">{typeLabel}</Select.Trigger>
					<Select.Content>
						<Select.Item value="" label="None">None</Select.Item>
						{#each types as type (type.id)}
							<Select.Item value={type.id} label={type.attributes.name}>
								{type.attributes.name}
							</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>
			{:else}
				{@render notConfigured("types")}
			{/if}
		</div>
	</div>

	{#if hasMoreDetails}
		<Collapsible.Root>
			<Collapsible.Trigger class="group flex items-center gap-1 text-sm font-medium">
				<RiArrowDownSLine class="size-4 transition-transform group-data-[state=open]:rotate-180" />
				More details
			</Collapsible.Trigger>
			<Collapsible.Content class="flex flex-col gap-4 pt-3">
				{#if tags.length}
					<div class="flex flex-col gap-2">
						<span class="text-sm font-medium">Tags</span>
						<div class="grid gap-2 sm:grid-cols-2">
							{#each tags as tag (tag.id)}
								<label class="flex items-center gap-2 border px-2 py-2 text-xs">
									<Checkbox
										checked={form.tagIds.includes(tag.id)}
										onCheckedChange={() => toggleTag(tag.id)}
									/>
									<span>{tag.attributes.key}: {tag.attributes.value}</span>
								</label>
							{/each}
						</div>
					</div>
				{/if}

				{#if fields.length}
					<div class="grid gap-4 sm:grid-cols-2">
						{#each fields as field (field.id)}
							<div class="flex flex-col gap-2">
								<Label for="{idPrefix}-field-{field.id}">{field.attributes.name}</Label>
								<Select.Root
									type="single"
									bind:value={
										() => form.fieldSelections[field.id] ?? "",
										(value) => setFieldSelection(field.id, value)
									}
								>
									<Select.Trigger id="{idPrefix}-field-{field.id}" class="w-full">
										{fieldSelectionLabel(field)}
									</Select.Trigger>
									<Select.Content>
										<Select.Item value="" label="None">None</Select.Item>
										{#each field.attributes.options as option (option.id)}
											<Select.Item value={option.id} label={option.attributes.value}>
												{option.attributes.value}
											</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							</div>
						{/each}
					</div>
				{/if}
			</Collapsible.Content>
		</Collapsible.Root>
	{/if}
</div>
