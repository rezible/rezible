import type { CreateSystemAnalysisEntryAttributes, SystemAnalysisEntry } from "$lib/api";

// Entry writes replace subjects, even when the form only edits narrative fields.
export function entryInput(
	fields: Omit<CreateSystemAnalysisEntryAttributes, "subjects">,
	entry?: SystemAnalysisEntry
): CreateSystemAnalysisEntryAttributes {
	return {
		...fields,
		subjects: (entry?.attributes.subjects ?? []).map(({ attributes }) => ({
			role: attributes.role,
			knowledgeEntityId: attributes.knowledgeEntityId,
			knowledgeRelationshipId: attributes.knowledgeRelationshipId,
			knowledgeEvidenceId: attributes.knowledgeEvidenceId,
		})),
	};
}
