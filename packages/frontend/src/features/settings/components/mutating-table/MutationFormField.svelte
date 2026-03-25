<script lang="ts">
	import { z } from "zod";
	import LoadingSelect from './LoadingSelect.svelte';
	import { isSelectField, unwrappedSchema, type Field } from './fields.svelte';


	type Props = {
		name: string;
		dataType: string;
		field: Field;
		editing: boolean;
		values: Record<string, any>;
		errors: Record<string, string>;
		onFieldUpdate: (name: string) => void;	
	}
	const { name, dataType, field, editing, values, errors, onFieldUpdate }: Props = $props();

	const id = `${editing ? "edit" : "create"}-${dataType}-${name}`
	const label = field.label
	const error = errors[name]
	const schemaBase = unwrappedSchema(field.schema)
	const onChange = () => onFieldUpdate(name);
</script>

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
		<LoadingSelect {id} bind:value={values[name]} options={field.options} {label} />
	{:else}
		<span>regular select</span>
	{/if}
{:else if schemaBase instanceof z.ZodString}
	<span>string</span>
	<!-- <TextField
		{label}
		{error}
		labelPlacement="float"
		value={values[name]}
		debounceChange={100}
		on:change={({ detail }) => {
			values[name] = detail.value ?? "";
			onChange();
		}}
	/> -->
{:else if schemaBase instanceof z.ZodBoolean}
	<span>boolean</span>
	<!-- <UxField label={field.label} let:id {error}>
		<Switch {id} bind:checked={values[name]} on:change={onChange} />
	</UxField> -->
{:else if schemaBase instanceof z.ZodNumber}
	<span>number</span>
{:else}
	<span>unknown {typeof field}</span>
	<span>{field.label} {typeof field.schema}</span>
{/if}