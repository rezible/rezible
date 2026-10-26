import type { JourneyPlan } from "../types";

export default {
	description: "Open Home and follow the Situations navigation link.",
	prerequisites: ["Live dev stack with authentication skipped. No existing records required."],
	viewport: { width: 1440, height: 900 },
	steps: [
		{ action: "navigate", path: "/" },
		{ action: "waitFor", selector: "#home-situations", state: "visible" },
		{ action: "waitFor", selector: '[aria-label="Loading open situations"]', state: "absent" },
		{ action: "screenshot", name: "01-home", caption: "Home page before navigating to Situations." },
		{ action: "click", selector: 'a[href="/situations"][aria-label="Situations"]' },
		{ action: "waitFor", selector: 'a[href="/situations"][aria-current="page"]', state: "visible" },
		{ action: "waitFor", selector: '[role="tablist"][aria-label="Situation status"]', state: "visible" },
		{ action: "waitFor", selector: '[aria-label="Loading situations"]', state: "absent" },
		{
			action: "screenshot",
			name: "02-situations",
			caption: "Situations page reached through the navigation rail.",
		},
	],
} satisfies JourneyPlan;
