import type { ListQueryOptionsFunc } from '$lib/api';
import type { Component } from 'svelte';
import type { MenuOption } from 'svelte-ux';
import { z } from 'zod';

export type EditorSnippetProps<T> = {id: string, value: T, onUpdate: (value: T) => void};
type EditorComponent<T> = Component<EditorSnippetProps<T>>;

type BaseField<T> = {
	label: string;
	schema: z.ZodTypeAny;
	editor?: EditorComponent<T>;
};

export type Field = BaseField<any>;

export type SelectFieldOptions = MenuOption<string>[] | ListQueryOptionsFunc<any>;

type SelectFieldValueType = string | string[] | undefined;
export type SelectField<T extends SelectFieldValueType> = BaseField<T> & {options: SelectFieldOptions};

export const unwrappedSchema = (s: z.ZodTypeAny): z.ZodTypeAny => {
	// if ("unwrap" in s) return unwrappedSchema(s.unwrap())
	if (s instanceof z.ZodOptional) return unwrappedSchema(s.unwrap());
	if (s instanceof z.ZodNullable) return unwrappedSchema(s.unwrap());
	return s;
}
export const isSelectField = (f: Field): f is SelectField<SelectFieldValueType> => 'options' in f;

// export type FormFields = { [name: string]: Field };
export type FormFields = Record<string, BaseField<any>>;

export const makeField = <T>(
	label: string,
	schema: z.ZodType<T>
): BaseField<T> => ({label, schema});

export const makeSelectField = <T extends SelectFieldValueType>(
	label: string,
	schema: z.ZodType<T>,
	options: SelectFieldOptions
): SelectField<T> => ({label, schema, options});

export const makeCustomField = <T>(
	label: string,
	schema: z.ZodType<T>,
	editor: EditorComponent<T>,
): Field => ({ label, schema, editor });
