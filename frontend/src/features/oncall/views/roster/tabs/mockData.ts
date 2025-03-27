import type { ActivityItem } from "../types";

export const mockRecentActivity: ActivityItem[] = [
	{
		id: "incident-1",
		type: "incident",
		title: "Database Outage",
		timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24),
		user: {
			id: "user-2",
			attributes: {
				name: "User Name",
				email: "user@example.com"
			}
		}
	},
	{
		id: "handover-1",
		type: "handover",
		title: "Weekly Handover Completed",
		timestamp: new Date(Date.now() - 1000 * 60 * 60 * 48),
		user: {
			id: "user-1",
			attributes: {
				name: "User Name",
				email: "user@example.com"
			}
		}
	},
	{
		id: "playbook-1",
		type: "playbook",
		title: "API Outage Playbook Updated",
		timestamp: new Date(Date.now() - 1000 * 60 * 60 * 72),
		user: {
			id: "user-3",
			attributes: {
				name: "User Name",
				email: "user@example.com"
			}
		}
	},
	{
		id: "backlog-1",
		type: "backlog",
		title: "Improve Alert Filtering",
		timestamp: new Date(Date.now() - 1000 * 60 * 60 * 96),
		user: {
			id: "user-3",
			attributes: {
				name: "User Name",
				email: "user@example.com"
			}
		}
	}
];