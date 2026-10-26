import type { JourneyPlan } from "../types";

export default {
	description:
		"Inspect seeded incidents, analysis and report; capture declaration validation without creating a record.",
	prerequisites: [
		"Running dev stack with auth skipped; setup-demo completed. No valid forms are submitted.",
	],
	viewport: {
		width: 1440,
		height: 900,
	},
	steps: [
		{
			action: "navigate",
			path: "/incidents",
		},
		{
			action: "waitFor",
			selector: 'a[href^="/incidents/"]:first-of-type',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "01-list",
			caption: "Three seeded resolved incidents; default list filters.",
		},
		{
			action: "resize",
			width: 1024,
			height: 768,
		},
		{
			action: "screenshot",
			name: "02-list-1024",
			caption: "Incidents list at a narrower viewport.",
		},
		{
			action: "resize",
			width: 1440,
			height: 900,
		},
		{
			action: "click",
			selector: 'a[href^="/incidents/"]:first-of-type',
		},
		{
			action: "waitFor",
			selector: '[data-slot="feature-navigation-link"][href$="/analysis"]',
			state: "visible",
		},
		{
			action: "waitFor",
			selector: 'button[aria-label="More incident actions"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "03-overview",
			caption: "Seeded checkout incident overview, milestones, roles and linked situation.",
		},
		{
			action: "resize",
			width: 1024,
			height: 768,
		},
		{
			action: "screenshot",
			name: "04-overview-1024",
			caption: "Incident detail at a narrower viewport.",
		},
		{
			action: "resize",
			width: 1440,
			height: 900,
		},
		{
			action: "click",
			selector: '[data-slot="feature-navigation-link"][href$="/analysis"]',
		},
		{
			action: "waitFor",
			selector: '[data-slot="feature-navigation-link"][href$="/analysis"][aria-current="page"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "05-analysis",
			caption:
				"Analysis view after switching from Overview. Async investigation content may be unavailable.",
		},
		{
			action: "click",
			selector: '[data-slot="feature-navigation-link"][href$="/report"]',
		},
		{
			action: "waitFor",
			selector: '[data-slot="feature-navigation-link"][href$="/report"][aria-current="page"]',
			state: "visible",
		},
		{
			action: "waitFor",
			selector: "#summary",
			state: "visible",
		},
		{
			action: "screenshot",
			name: "06-report",
			caption: "Seeded retrospective draft in the Report view; no edits.",
		},
		{ action: "navigate", path: "/incidents" },
		{ action: "click", selector: 'header button[data-slot="button"]' },
		{
			action: "waitFor",
			selector: '[role="dialog"] input[id$="-title"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "07-declare",
			caption: "Empty declaration dialog.",
		},
		{
			action: "click",
			selector: '[role="dialog"] button[type="submit"]',
		},
		{
			action: "waitFor",
			selector: '[role="dialog"] [role="alert"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "08-validation",
			caption: "Client validation after attempting submission with an empty title.",
		},
		{
			action: "fill",
			selector: '[role="dialog"] input[id$="-title"]',
			value: "REVIEW-collector: checkout latency",
		},
		{
			action: "screenshot",
			name: "09-filled",
			caption: "Title entered; no valid submission.",
		},
		{
			action: "press",
			key: "Escape",
		},
		{
			action: "waitFor",
			selector: '[role="dialog"]',
			state: "absent",
		},
	],
} satisfies JourneyPlan;
