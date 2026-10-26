import type { JourneyPlan } from "../types";

export default {
	description:
		"Inspect settings and integration coverage without changing preferences or installing anything.",
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
			path: "/settings/user",
		},
		{
			action: "waitFor",
			selector: "#user-name",
			state: "visible",
		},
		{
			action: "screenshot",
			name: "01-preferences",
			caption: "Personal preferences and theme controls; no settings changed.",
		},
		{
			action: "resize",
			width: 1024,
			height: 768,
		},
		{
			action: "screenshot",
			name: "02-preferences-1024",
			caption: "Personal preferences at a narrower viewport.",
		},
		{
			action: "resize",
			width: 1440,
			height: 900,
		},
		{
			action: "navigate",
			path: "/settings/user/notifications",
		},
		{
			action: "waitFor",
			selector: 'a[href="/settings/user/notifications"][aria-current="page"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "03-notifications",
			caption: "Known unfinished notification preferences.",
		},
		{
			action: "navigate",
			path: "/settings/organization",
		},
		{
			action: "waitFor",
			selector: 'a[href="/settings/organization"][aria-current="page"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "04-organization",
			caption: "Organization preferences; no changes submitted.",
		},
		{
			action: "navigate",
			path: "/settings/organization/members",
		},
		{
			action: "waitFor",
			selector: 'a[href="/settings/organization/members"][aria-current="page"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "05-members",
			caption: "Organization members.",
		},
		{
			action: "navigate",
			path: "/settings/incidents",
		},
		{
			action: "waitFor",
			selector: 'a[href="/settings/incidents"][aria-current="page"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "06-incidents",
			caption: "Incident management settings after setup-demo.",
		},
		{
			action: "navigate",
			path: "/settings/integrations",
		},
		{
			action: "waitFor",
			selector: 'a[aria-label="Manage Demo"]',
			state: "visible",
		},
		{
			action: "screenshot",
			name: "07-integrations",
			caption: "Installed demo and available integration providers.",
		},
		{
			action: "click",
			selector: 'a[aria-label="Manage Demo"]',
		},
		{
			action: "waitFor",
			selector: "h1",
			state: "visible",
		},
		{
			action: "screenshot",
			name: "08-demo",
			caption: "Installed demo integration management; no actions submitted.",
		},
		{
			action: "navigate",
			path: "/settings/integrations/alertmanager",
		},
		{
			action: "waitFor",
			selector: "h1",
			state: "visible",
		},
		{
			action: "screenshot",
			name: "09-alertmanager",
			caption: "Uninstalled Alertmanager setup surface; nothing installed.",
		},
	],
} satisfies JourneyPlan;
