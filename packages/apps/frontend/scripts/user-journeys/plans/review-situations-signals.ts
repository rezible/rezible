import type { JourneyPlan } from "../types";

export default {
	description:
		"Follow a raised situation into investigation, then inspect signal definition and recent activity.",
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
			path: "/situations",
		},
		{
			action: "waitFor",
			selector: 'ul > li:first-child a[href^="/situations/"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "01-situations",
			caption: "Raised situations after demo seeding.",
		},
		{
			action: "click",
			selector: 'ul > li:first-child a[href^="/situations/"]',
		},
		{
			action: "waitFor",
			selector: '[data-slot="feature-navigation-link"][href$="/investigation"]',
			state: "visible",
		},
		{
			action: "waitFor",
			selector: 'button[aria-label="Inspect Search API response time high"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "02-brief",
			caption:
				"Situation brief with observations and the current investigation state. No retry is requested.",
		},
		{
			action: "resize",
			width: 1024,
			height: 768,
		},
		{
			action: "screenshot",
			name: "03-brief-1024",
			caption: "Situation brief at a narrower viewport.",
		},
		{
			action: "resize",
			width: 1440,
			height: 900,
		},
		{
			action: "click",
			selector: '[data-slot="feature-navigation-link"][href$="/investigation"]',
		},
		{
			action: "waitFor",
			selector: '[data-slot="feature-navigation-link"][href$="/investigation"][aria-current="page"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "04-investigation",
			caption:
				"Investigation view; any seeded failed turn is retained as evidence rather than retried.",
		},
		{
			action: "navigate",
			path: "/signals",
		},
		{
			action: "waitFor",
			selector: 'li:first-child > a[href^="/signals/"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "05-signals",
			caption: "Five seeded signal definitions.",
		},
		{
			action: "click",
			selector: 'li:first-child > a[href^="/signals/"]',
		},
		{
			action: "waitFor",
			selector: '[data-slot="feature-navigation-link"][href$="/events"]',
			state: "visible",
		},
		{
			action: "waitFor",
			selector: "#scroll-body h1",
			state: "visible",
		},
		{
			action: "screenshot",
			name: "06-signal",
			caption:
				"Signal overview including attention controls, grouping labels and metrics; nothing changed.",
		},
		{
			action: "click",
			selector: '[data-slot="feature-navigation-link"][href$="/events"]',
		},
		{
			action: "waitFor",
			selector: '[data-slot="feature-navigation-link"][href$="/events"][aria-current="page"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "07-signal-events",
			caption: "Recent Activity view after switching from signal overview.",
		},
		{
			action: "click",
			selector: '[data-slot="feature-navigation-link"][href$="/incidents"]',
		},
		{
			action: "waitFor",
			selector: '[data-slot="feature-navigation-link"][href$="/incidents"][aria-current="page"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "08-signal-incidents",
			caption:
				"Known unfinished Incidents view for context; incompleteness alone is excluded from findings.",
		},
	],
} satisfies JourneyPlan;
