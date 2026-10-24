import type { IncidentAttributes, IncidentMilestone } from "$lib/api";

export type ResponseAction = "mitigated" | "mitigationFailed" | "resolve";

/**
 * The response actions an incident offers. A started incident can be mitigated or resolved; a mitigated one
 * can return to started, when its mitigation failed, or be resolved. Resolved is final. Nothing is offered
 * while incident management is off.
 */
export function responseActions(
	state: IncidentAttributes["responseState"],
	enabled: boolean
): ResponseAction[] {
	if (!enabled) {
		return [];
	}
	switch (state) {
		case "started":
			return ["mitigated", "resolve"];
		case "mitigated":
			return ["mitigationFailed", "resolve"];
		case "resolved":
			return [];
	}
}

/** Milestones oldest first. Equal times are ordered by ID. */
export function orderedMilestones(milestones: IncidentMilestone[]): IncidentMilestone[] {
	return [...milestones].sort((left, right) => {
		const byTime = Date.parse(left.attributes.timestamp) - Date.parse(right.attributes.timestamp);
		if (byTime !== 0) {
			return byTime;
		}
		if (left.id === right.id) {
			return 0;
		}
		if (left.id < right.id) {
			return -1;
		}
		return 1;
	});
}

/** The impact milestones recorded while the incident was mitigated: each records that a mitigation failed. */
export function failedMitigationIds(milestones: IncidentMilestone[]): Set<string> {
	const failed = new Set<string>();
	let mitigated = false;
	for (const milestone of orderedMilestones(milestones)) {
		const kind = milestone.attributes.kind;
		if (kind === "impact") {
			if (mitigated) {
				failed.add(milestone.id);
			}
			mitigated = false;
		} else if (kind === "mitigation") {
			mitigated = true;
		}
	}
	return failed;
}
