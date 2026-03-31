<script module lang="ts">
	interface ApiDataObject {
		attributes: Record<string, any>;
	}
</script>

<script lang="ts" generics="DataType extends ApiDataObject, CreateAttributes">
	import type { ErrorModel } from "$lib/api";
	import ConfirmChangeButtons from "$components/confirm-buttons/ConfirmButtons.svelte";
	import type { FormFields, Field } from "./fields.svelte";
	import MutationFormField from "./MutationFormField.svelte";

	type Props = {
		fields: FormFields;
		dataType: string;
		onClose: () => void;
		onMutate: (attr: CreateAttributes) => void;
		isPending: boolean;
		mutationError: ErrorModel | null;
		data?: DataType;
	};
	const { fields, dataType, onClose, onMutate, isPending, mutationError, data }: Props = $props();

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
				// errors[name] = res.error?.formErrors.formErrors.join("\n");
			} else {
				delete errors[name];
			}
		});
	};

	// TODO: is this expensive?
	const onFieldUpdate = (name: string) => {
		const res = fields[name].schema.safeParse(values[name]);
		if (res.error) {
			// errors[name] = res.error?.formErrors.formErrors.join("\n");
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

<form class="w-full bg-base-300 shadow-xl p-4 border-accent/20 border-2 rounded">
	<div class="flex flex-col gap-2">
		<span>{editing ? "Edit" : "Create New"} {dataType}</span>
		{#each Object.entries(fields) as [name, field]}
			{#if field.schema.isOptional()}
				<span>optional field</span><!-- {@render optionalFormField(name, field)} -->
			{:else}
				<MutationFormField {name} {dataType} {field} {editing} {values} {errors} {onFieldUpdate} />
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

<!-- 
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
			<MutationFormField {name} {field} {dataType} {editing} {values} {errors} {onFieldUpdate} />
		{/if}
	</Toggle>
{/snippet}
 -->
