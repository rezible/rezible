import { z } from "zod";
import type {
	CreateIncidentAttributes,
	Incident,
	SituationAttributes,
	UpdateIncidentAttributes,
} from "$lib/api";

/** The fields that describe an incident. An empty severity or type means none. */
export type IncidentDetailsFormState = {
	title: string;
	summary: string;
	severityId: string;
	typeId: string;
	tagIds: string[];
	fieldSelections: Record<string, string>;
};

/** A situation chosen to link when declaring. */
export type DeclareSituation = {
	id: string;
	title: string;
	stage: SituationAttributes["stage"];
	linkedIncidentIds: string[];
};

export type CreateIncidentFormState = IncidentDetailsFormState & {
	situations: DeclareSituation[];
};

export const getEmptyDetailsForm = (): IncidentDetailsFormState => ({
	title: "",
	summary: "",
	severityId: "",
	typeId: "",
	tagIds: [],
	fieldSelections: {},
});

export const getEmptyCreateIncidentForm = (): CreateIncidentFormState => ({
	...getEmptyDetailsForm(),
	situations: [],
});

export const detailsFormFromIncident = (incident: Incident): IncidentDetailsFormState => {
	const attributes = incident.attributes;
	const fieldSelections: Record<string, string> = {};
	for (const selection of attributes.fieldSelections) {
		fieldSelections[selection.fieldId] = selection.option.id;
	}
	return {
		title: attributes.title,
		summary: attributes.summary,
		severityId: attributes.severity?.id ?? "",
		typeId: attributes.type?.id ?? "",
		tagIds: attributes.tags.map((tag) => tag.id),
		fieldSelections,
	};
};

const optionalId = z.union([z.literal(""), z.uuid()]);

const detailsFormSchema = z.object({
	title: z.string().trim().min(1, "Title is required"),
	summary: z.string().trim(),
	severityId: optionalId,
	typeId: optionalId,
	tagIds: z.array(z.uuid()),
	fieldSelections: z.record(z.string(), z.uuid()),
});

export const CreateIncidentFormSchema = detailsFormSchema
	.extend({ situations: z.array(z.looseObject({ id: z.uuid() })) })
	.transform((form): CreateIncidentAttributes => ({
		title: form.title,
		summary: form.summary || undefined,
		severityId: form.severityId || undefined,
		typeId: form.typeId || undefined,
		tagIds: form.tagIds,
		fieldSelectionIds: Object.values(form.fieldSelections),
		situationIds: form.situations.map((situation) => situation.id),
	}));

/** Every detail is sent, so clearing a severity or type sends null. */
export const UpdateIncidentFormSchema = detailsFormSchema.transform((form): UpdateIncidentAttributes => ({
	title: form.title,
	summary: form.summary,
	severityId: form.severityId || null,
	typeId: form.typeId || null,
	tagIds: form.tagIds,
	fieldSelectionIds: Object.values(form.fieldSelections),
}));
