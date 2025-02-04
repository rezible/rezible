<script module lang="ts">
	interface ApiDataObject {
		attributes: Record<string, any>;
	}
</script>

<script lang="ts" generics="DataType extends ApiDataObject, CreateAttributes">
	import { z } from "zod";
	import type { ErrorModel } from "$lib/api";
	import ConfirmChangeButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import LoadingSelect from "$components/loading-select/LoadingSelect.svelte";
	import {
		type FormFields,
		type Field,
		isSelectField,
		unwrappedSchema,
	} from "./fields.svelte";
	import { Field as UxField, Switch, TextField, Toggle } from "svelte-ux";

	type Props = {
		fields: FormFields;
		dataType: string;
		onClose: () => void;
		onMutate: (attr: CreateAttributes) => void;
		isPending: boolean;
		mutationError: ErrorModel | null;
		data?: DataType;
	};
	const {
		fields,
		dataType,
		onClose,
		onMutate,
		isPending,
		mutationError,
		data,
	}: Props = $props();

	const editing = $derived(!!data);
	const fieldNames = $derived(Object.keys(fields));

	let submitEnabled = $state(false);

	let values = $state<Record<string, any>>({});
	let errors = $state<Record<string, string>>({});
	let apiError = $state<string | undefined>();

	const clearErrors = () => {
		fieldNames.forEach((name) => {
			delete errors[name];
		});
		apiError = undefined;
	};

	const setValues = (data?: DataType) => {
		fieldNames.forEach((name) => {
			values[name] = data?.attributes[name];
		});
		clearErrors();
	};
	setValues(data);

	const maybeValidateFields = () => {
		Object.entries(values).forEach(([name, value]) => {
			const res = fields[name].schema.safeParse(values[name]);
			if (res.error) {
				errors[name] = res.error?.formErrors.formErrors.join("\n");
			} else {
				delete errors[name];
			}
		});
	};

	// TODO: is this expensive?
	const onFieldUpdate = (name: string) => {
		const res = fields[name].schema.safeParse(values[name]);
		if (res.error) {
			errors[name] = res.error?.formErrors.formErrors.join("\n");
			submitEnabled = false;
		} else {
			delete errors[name];
			maybeValidateFields();
		}
		const errorVals = Object.values(errors);
		console.log(Object.keys(errors), errorVals);
		submitEnabled = errorVals.length == 0;
	};

	const onError = (err: Error) => {
		console.error(err);
		if (true) {
			apiError = "Request failed: " + err.message;
			return;
		}
		/*
		const responseErrors = apiErr.body.errors || [];
		if (responseErrors.length === 0) {
			apiError = apiErr.message ?? "unknown error";
			return
		}
		responseErrors.forEach((e) => {
			if (!e.location) return;
			let name = e.location;
			if (name.startsWith('body.attributes')) {
				name = name.replaceAll('body.attributes.', '');
			}
			errors[name] = e.message || "invalid";
		});
		*/
	};
</script>

<form
	class="w-full bg-base-300 shadow-xl p-4 border-accent/20 border-2 rounded"
>
	<div class="flex flex-col gap-2">
		<span>{editing ? "Edit" : "Create New"} {dataType}</span>
		{#each Object.entries(fields) as [name, field]}
			{#if field.schema.isOptional()}
				{@render optionalFormField(name, field)}
			{:else}
				{@render formField(name, field)}
			{/if}
		{/each}
	</div>

	{#if !!mutationError}
		<div class="border border-red-300 flex flex-col gap-1">
			<span class="text-md font-medium text-red-400">Error</span>
			<span class="text-sm text-red-300">{mutationError}</span>
		</div>
	{/if}

	<div class="mt-2 w-full flex flex-row-reverse">
		<ConfirmChangeButtons
			confirmText={isPending ? "Saving..." : "Save"}
			closeText={"Cancel"}
			loading={isPending}
			saveEnabled={submitEnabled}
			onConfirm={() => onMutate(values as CreateAttributes)}
			{onClose}
		/>
	</div>
</form>

{#snippet optionalFormField(name: string, field: Field)}
	{@const isOptional = field.schema.isOptional()}
	<Toggle on={!isOptional} let:toggle let:on>
		<div class="grid gap-2">
			<label for="#" class="flex gap-2 items-center text-sm">
				<Switch on:change={toggle} />
				{field.schema.description ?? field.label}
			</label>
		</div>

		{#if on}
			{@render formField(name, field)}
		{/if}
	</Toggle>
{/snippet}

{#snippet formField(name: string, field: Field)}
	{@const id = `${editing ? "edit" : "create"}-${dataType}-${name}`}
	{@const label = field.label}
	{@const error = errors[name]}
	{@const schemaBase = unwrappedSchema(field.schema)}
	{@const onChange = () => onFieldUpdate(name)}

	{#if field.editor}
		<field.editor
			{id}
			value={values[name]}
			onUpdate={(value) => {
				values[name] = value;
				onChange();
			}}
		/>
	{:else if isSelectField(field)}
		{#if !Array.isArray(field.options)}
			<LoadingSelect
				{id}
				bind:value={values[name]}
				options={field.options}
				{label}
			/>
		{:else}
			<span>regular select</span>
		{/if}
	{:else if schemaBase instanceof z.ZodString}
		<TextField
			{label}
			{error}
			labelPlacement="float"
			value={values[name]}
			debounceChange={100}
			on:change={({ detail }) => {
				values[name] = detail.value ?? "";
				onChange();
			}}
		/>
	{:else if schemaBase instanceof z.ZodBoolean}
		<UxField label={field.label} let:id {error}>
			<Switch {id} bind:checked={values[name]} on:change={onChange} />
		</UxField>
	{:else if schemaBase instanceof z.ZodNumber}
		<span>number</span>
	{:else}
		<span>unknown {typeof field}</span>
		<span>{field.label} {typeof field.schema}</span>
	{/if}
{/snippet}
