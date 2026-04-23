import { z } from "zod";
import type { CreateIncidentAttributes } from "$lib/api";

export type CreateIncidentFormState = {
	title: string;
	summary: string;
	severityId: string;
	typeId: string;
	tagIds: string[];
	fieldSelections: Record<string, string>;
};

export const getEmptyCreateIncidentForm = (): CreateIncidentFormState => ({
	title: "",
	summary: "",
	severityId: "",
	typeId: "",
	tagIds: [],
	fieldSelections: {},
});

const createIncidentFormSchema = z.object({
	title: z.string().trim().min(1, "Title is required"),
	summary: z.string().trim().optional(),
	severityId: z.uuid("Select a severity"),
	typeId: z.uuid("Select an incident type"),
	tagIds: z.array(z.uuid()).default([]),
	fieldSelections: z.record(z.string(), z.uuid()).default({}),
});

export const CreateIncidentFormSchema = createIncidentFormSchema.transform(
	(form): CreateIncidentAttributes => ({
		title: form.title,
		severityId: form.severityId,
		typeId: form.typeId,
		summary: form.summary ? form.summary : undefined,
		tagIds: form.tagIds,
		fieldSelectionIds: Object.values(form.fieldSelections),
	}),
);
