import type { JourneyPlan } from "../types";

export default {
	description: "Inspect system catalogue and zero-result filtering, then the saved-views placeholder.",
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
			path: "/system/catalogue",
		},
		{
			action: "waitFor",
			selector: "#catalogue-search",
			state: "visible",
		},
		{
			action: "waitFor",
			selector: '[role="tabpanel"] > div > div:first-child > ul > li:first-child',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "01-catalogue",
			caption: "System catalogue, proposed matches area and service rows.",
		},
		{
			action: "fill",
			selector: "#catalogue-search",
			value: "REVIEW-no-such-service-987654",
		},
		{
			action: "waitFor",
			selector: '[role="tabpanel"] > div > div:first-child > ul',
			state: "absent",
		},
		{
			action: "screenshot",
			name: "02-no-results",
			caption: "Catalogue after filtering removes all service rows; no matching services.",
		},
		{
			action: "fill",
			selector: "#catalogue-search",
			value: "",
		},
		{
			action: "navigate",
			path: "/system/saved-views",
		},
		{
			action: "waitFor",
			selector: '[data-slot="feature-navigation-link"][href$="/saved-views"][aria-current="page"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "03-saved-views",
			caption: "Known unfinished saved views surface.",
		},
	],
} satisfies JourneyPlan;
