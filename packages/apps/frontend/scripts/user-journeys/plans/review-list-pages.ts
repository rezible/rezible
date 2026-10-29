import type { JourneyPlan } from "../types";

export default {
	description:
		"Walk Home and the incidents, situations and signals lists: filter column, shared rows, pager placement, filtering, and the narrow-screen filter sheet.",
	prerequisites: ["Running dev stack with auth skipped; setup-demo completed. No forms are submitted."],
	viewport: { width: 1440, height: 900 },
	steps: [
		{ action: "navigate", path: "/" },
		{ action: "waitFor", selector: "#home-incidents", state: "visible" },
		{ action: "waitFor", selector: '[aria-label="Loading active incidents"]', state: "absent" },
		{ action: "waitFor", selector: '[aria-label="Loading open situations"]', state: "absent" },
		{ action: "screenshot", name: "01-home", caption: "Home sections using the shared row." },

		{ action: "navigate", path: "/incidents" },
		{ action: "waitFor", selector: 'ul > li:first-child a[data-row-link][href^="/incidents/"]', state: "visible" },
		{ action: "screenshot", name: "02-incidents", caption: "Incidents with the filter column." },
		{ action: "click", selector: 'label:has(> input[value="resolved"])' },
		{ action: "waitFor", selector: 'label:has(> input[value="resolved"]:checked)', state: "visible" },
		{ action: "waitFor", selector: 'ul > li:first-child a[data-row-link][href^="/incidents/"]', state: "absent" },
		{ action: "screenshot", name: "03-incidents-resolved", caption: "Incidents filtered to resolved." },

		{ action: "navigate", path: "/situations" },
		{ action: "waitFor", selector: 'label:has(> input[value="raised"]:checked)', state: "visible" },
		{ action: "waitFor", selector: '[aria-label="Loading situations"]', state: "absent" },
		{ action: "screenshot", name: "04-situations", caption: "Situations, raised status." },
		{ action: "click", selector: 'label:has(> input[value="watching"])' },
		{ action: "waitFor", selector: 'label:has(> input[value="watching"]:checked)', state: "visible" },
		{ action: "waitFor", selector: 'ul > li:first-child a[data-row-link][href^="/situations/"]', state: "absent" },
		{ action: "screenshot", name: "05-situations-watching", caption: "Situations, watching status." },

		{ action: "navigate", path: "/signals" },
		{ action: "waitFor", selector: 'ul > li:first-child a[data-row-link][href^="/signals/"]', state: "visible" },
		{ action: "screenshot", name: "06-signals", caption: "Signals with search and attention filters." },
		{ action: "fill", selector: 'input[type="search"]', value: "zzzz-no-match" },
		{ action: "waitFor", selector: 'ul > li:first-child a[data-row-link][href^="/signals/"]', state: "absent" },
		{ action: "screenshot", name: "07-signals-empty", caption: "Signals search with no match." },

		{ action: "resize", width: 390, height: 844 },
		{ action: "navigate", path: "/incidents" },
		{ action: "waitFor", selector: 'ul > li:first-child a[data-row-link][href^="/incidents/"]', state: "visible" },
		{ action: "screenshot", name: "08-incidents-narrow", caption: "Incidents on a narrow screen." },
		{ action: "click", selector: 'main button[aria-haspopup="dialog"]' },
		{ action: "waitFor", selector: '[data-slot="sheet-content"]', state: "visible" },
		{ action: "screenshot", name: "09-filter-sheet", caption: "Filter sheet on a narrow screen." },
		{ action: "fill", selector: '[data-slot="sheet-content"] input[type="search"]', value: "zzzz-no-match" },
		{ action: "waitFor", selector: '[data-slot="sheet-content"] button[data-slot="button"]', state: "visible" },
		{ action: "screenshot", name: "10-filter-sheet-active", caption: "Filter sheet with an active filter." },
	],
} satisfies JourneyPlan;
