import { addMinutes, subHours } from "date-fns";

export type EventData = {
	source: "slack" | "github";
	id: string;
};

export type IncidentStage = "impact" | "detection" | "response" | "mitigation";

export type IncidentMilestone =
	| "impact_start"
	| "metrics"
	| "alert"
	| "response_start"
	| "incident_detail"
	| "hypothesis"
	| "mitigation_attempt"
	| "mitigated";

export type EventType = "milestone" | "stage" | "note";

type BaseIncidentEvent = {
	id: string;
	title: string;
	type: EventType;
	start: Date;
	end?: Date;
	data?: any[];
	stage_change?: IncidentStage;
};

export type IncidentMilestoneEvent = BaseIncidentEvent & {
	type: "milestone";
	milestone: IncidentMilestone;
};

export type IncidentNoteEvent = BaseIncidentEvent & {
	type: "note";
	description: string;
};

export type IncidentEvent = IncidentMilestoneEvent | IncidentNoteEvent;

const start = subHours(Date.now(), 12);

export let events: IncidentEvent[] = [
	{
		id: "m2",
		title: "Impact Started",
		type: "milestone",
		start: addMinutes(start, 5),
		milestone: "metrics",
		stage_change: "detection",
	},
	{
		id: "m3",
		title: "Alert Fired",
		type: "milestone",
		start: addMinutes(start, 13),
		milestone: "alert",
	},
	{
		id: "m4",
		title: "Response Started",
		type: "milestone",
		start: addMinutes(start, 14),
		milestone: "response_start",
		stage_change: "response",
	},
	{
		id: "m8",
		title: "Incident Mitigation Attempted",
		type: "milestone",
		start: addMinutes(start, 30),
		milestone: "mitigation_attempt",
		stage_change: "mitigation",
	},
	{
		id: "m9",
		title: "Incident Mitigated",
		type: "milestone",
		start: addMinutes(start, 45),
		milestone: "mitigated",
	},
];
