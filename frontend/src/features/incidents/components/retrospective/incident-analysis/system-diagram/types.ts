export type ComponentSignal = {
	id: string;
	description: string;
	type: string;
	metadata?: Record<string, unknown>;
};

export type ComponentControl = {
	id: string;
	description: string;
	type: string;
	metadata?: Record<string, unknown>;
};

export type ComponentConstraint = {
	id: string;
	description: string;
	type: string;
	metadata?: Record<string, unknown>;
};

export type SystemComponent = {
	id: string;
	label: string;
	description?: string;
	metadata?: Record<string, any>;
	// category: string;
	constraints: ComponentConstraint[];
	controls: ComponentControl[];
	signals: ComponentSignal[];
};

export type SystemComponentRelationship = {
	id: string;
	sourceId: string;
	targetId: string;
	label: string;
	description?: string;
	parameters?: Record<string, any>;
	// category: string;
	controlActions: string[];
	signalFeedbacks: string[];
};

type ConstraintImpact = "degraded" | "failed";

type SystemEvent = {
	id: string;
	timestamp: string;
	title: string;
	description: string;
	componentConstraintStates: Record<string, Record<string, ConstraintImpact>>; // {componentId: {constraintId: ConstraintImpact}}
};

interface SystemHazard {
	label: string;
	severity: string;
	likelihood: string;
	constraints: string[]; // IDs of constraints that would be impacted
	feedbacks: string[]; // IDs of feedback relationships to detect
	controlActions: string[]; // IDs of control actions to mitigate
}
