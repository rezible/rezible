export type EventComponentState = 'normal' | 'degraded' | 'failed';

export const ContributingFactorCategories = [
	{
		id: "organizational_pressures",
		name: "Organizational Pressures",
		factors: [
			{
				id: "time_pressure",
				title: "Time Pressure",
				description: "Team was pushing to meet end-of-sprint commitments, leading to rushed testing",
				examples: [
					"End of quarter release pressure",
					"Multiple concurrent project deadlines",
					"Customer commitment deadlines",
				],
			},
			{
				id: "resource_constraints",
				title: "Resource Constraints",
				description: "Only one engineer familiar with the system was available during the incident",
				examples: [
					"Team understaffing",
					"Limited expert availability",
					"Budget constraints affecting tooling",
				],
			},
		],
	},
	{
		id: "knowledge_and_visibility",
		name: "Knowledge & Visibility",
		factors: [
			{
				id: "missing_information",
				title: "Missing Information",
				description: "",
				examples: [
					"Outdated runbooks",
					"Undocumented system interactions",
					"Missing architectural diagrams",
				],
			},
			{
				id: "monitoring_gaps",
				title: "Monitoring Gap",
				description: "No alerts existed for gradual database connection pool exhaustion",
				examples: ["Insufficient metrics coverage", "Missing threshold alerts", "Incomplete logging"],
			},
		],
	},
	{
		id: "process_and_coordination",
		name: "Process & Coordination",
		factors: [
			{
				id: "communication_breakdown",
				title: "Communication Breakdown",
				description: "Team wasn't sure when to involve senior engineers or wake up team leads",
				examples: [
					"Unclear escalation path",
					"Unclear incident roles",
					"Communication tool issues",
					"Cross-team coordination challenges",
				],
			},
			{
				id: "process_uncertainty",
				title: "Process Uncertainty",
				description: "No clear threshold for when to revert the deployment vs. trying to fix forward",
				examples: [
					"Missing playbooks",
					"Undefined rollback criteria",
					"Unclear decision authority",
					"Undefined incident severity levels",
				],
			},
		],
	},
	{
		id: "technical_complexity",
		name: "Technical Complexity",
		factors: [
			{
				id: "system_opacity",
				title: "System Opacity",
				description:
					"Initial Redis timeout led to unexpected cascading failures in multiple services",
				examples: [
					"Complex failure cascade",
					"Hidden dependencies",
					"Unclear failure modes",
					"Complex state management",
				],
			},
			{
				id: "technical_debt",
				title: "Technical Debt",
				description: "Old monitoring system couldn't be easily updated to catch new failure modes",
				examples: [
					"Legacy system constraints",
					"Outdated infrastructure",
					"Hard-to-maintain code",
					"Technical workarounds",
				],
			},
		],
	},
	{
		id: "change_management",
		name: "Change Management",
		factors: [
			{
				id: "configuration_complexity",
				title: "Configuration Complexity",
				description: "Multiple similar configuration parameters made it easy to modify the wrong one",
				examples: [
					"Risky configuration surface",
					"Complex configuration options",
					"Manual configuration steps",
					"Configuration drift",
				],
			},
			{
				id: "testing_limitations",
				title: "Testing Limitations",
				description: "Test environment didn't accurately reflect production load patterns",
				examples: ["Missing test coverage", "Environment differences", "Limited load testing"],
			},
		],
	},
	{
		id: "human_factors",
		name: "Human Factors",
		factors: [
			{
				id: "cognitive_load",
				title: "Cognitive Load",
				description: "High volume of low-priority alerts led to missing critical warnings",
				examples: [
					"Alert fatigue",
					"Information overload",
					"Fatigue during long incident",
					"Multiple concurrent issues",
				],
			},
			{
				id: "experience_mismatch",
				title: "Experience Mismatch",
				description: "Engineer was handling an incident in a system they rarely work with",
				examples: [
					"Unfamiliar territory",
					"New team members",
					"Cross-team coverage",
					"Rare system interactions",
				],
			},
		],
	},
	{
		id: "external_factors",
		name: "External Factors",
		factors: [
			{
				id: "vendor_issues",
				title: "Vendor Issues",
				description: "Unable to quickly scale due to cloud provider quota restrictions",
				examples: [
					"Cloud provider limitations",
					"Provider outages",
					"API limitations",
					"Third-party dependencies",
				],
			},
			{
				id: "customer_behavior",
				title: "Customer Behaviour",
				description: "New customer onboarding caused unexpected load spikes",
				examples: [
					"Unexpected/Changed usage patterns",
					"Traffic spikes",
					"Customer configuration issues",
				],
			},
		],
	},
];
