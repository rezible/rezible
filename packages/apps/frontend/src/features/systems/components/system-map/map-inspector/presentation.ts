import type {
	InspectedRelationship,
	InspectionVisibility,
	SystemMapInspection,
} from "$features/systems/lib/system-map/inspection";

export type InspectionRelationshipView = {
	id: string;
	text: string;
};

export type InspectorViewData = {
	title: string;
	visibilityText: string;
	representativeId?: string;
	relationship?: InspectionRelationshipView;
	summary?: {
		count: number;
		relationships: readonly InspectionRelationshipView[];
		missingRelationshipIds: readonly string[];
	};
	annotationTargetLabel?: string;
	memberships: readonly InspectionRelationshipView[];
	relationshipCountText: string;
};

const entityLabel = (entity: { label: string } | undefined, fallback: string): string =>
	entity?.label || fallback;

const predicateLabel = (predicate: string): string => predicate.replaceAll("_", " ");

const inspectionTitle = (inspection: SystemMapInspection): string => {
	if (inspection.entity) return inspection.entity.label;
	if (inspection.relationship?.relationship) {
		return predicateLabel(inspection.relationship.relationship.predicate);
	}
	if (inspection.summary) return "Connection summary";
	if (inspection.annotation?.entity) return inspection.annotation.entity.label;
	return "Inspection";
};

const visibilityLabels: Record<InspectionVisibility, string> = {
	visible: "Visible",
	represented: "Represented by another map subject",
	"not-rendered": "Not rendered",
	unavailable: "Unavailable",
};

const inspectionVisibilityText = (visibility: InspectionVisibility): string => visibilityLabels[visibility];

const relationshipDisplayText = (item: InspectedRelationship): string =>
	`${entityLabel(item.source, item.relationship.source)} → ${predicateLabel(item.relationship.predicate)} → ${entityLabel(item.target, item.relationship.target)}`;

const membershipDisplayText = (item: InspectedRelationship): string =>
	`${entityLabel(item.source, item.relationship.source)} → ${entityLabel(item.target, item.relationship.target)}`;

const relationshipView = (item: InspectedRelationship): InspectionRelationshipView => ({
	id: item.relationship.id,
	text: relationshipDisplayText(item),
});

export const buildInspectionViewData = (inspection?: SystemMapInspection): InspectorViewData | undefined => {
	if (!inspection) return undefined;

	return {
		title: inspectionTitle(inspection),
		visibilityText: inspectionVisibilityText(inspection.visibility),
		representativeId: inspection.representativeId,
		relationship: inspection.relationship ? relationshipView(inspection.relationship) : undefined,
		summary: inspection.summary
			? {
					count: inspection.summary.relationships.length,
					relationships: inspection.summary.relationships.map(relationshipView),
					missingRelationshipIds: inspection.summary.missingRelationshipIds,
				}
			: undefined,
		annotationTargetLabel: inspection.annotation
			? (inspection.annotation.originalTarget?.label ?? "Unavailable source target")
			: undefined,
		memberships: inspection.memberships.map((item) => ({
			id: item.relationship.id,
			text: membershipDisplayText(item),
		})),
		relationshipCountText: `${inspection.relationships.length} supplied relationship${inspection.relationships.length === 1 ? "" : "s"} in this inspection.`,
	};
};
