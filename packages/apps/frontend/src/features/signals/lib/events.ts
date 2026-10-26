import type { Event } from "$lib/api";

export type EventRowPresentation = {
	id: string;
	/** The event's own name for its subject, or its resource identifier when it has none. */
	title: string;
	/** The resource identifier, shown as secondary text when the title is a human name. */
	identifier?: string;
	kindLabel: string;
	occurredAt: string;
};

const kindLabels: Record<string, string> = {
	alert_instance: "Alert",
	code_change: "Code change",
	code_forge: "Repository",
	deployment: "Deployment",
	incident: "Incident",
	system_component: "System component",
	system_relationship: "System relationship",
	team: "Team",
	team_membership: "Team membership",
	user: "User",
};

/** Human label for an event kind; an unknown kind reads as sentence-case words. */
export function eventKindLabel(kind: string) {
	const known = kindLabels[kind];
	if (known) {
		return known;
	}

	const words = kind.replaceAll("_", " ").trim();
	return words.charAt(0).toUpperCase() + words.slice(1);
}

/** Attribute fields that name an event's subject, in order of preference. */
const titleFields = ["title", "display_name", "name"];

/** The event's attributes arrive as base64-encoded JSON. */
function decodeAttributes(encoded: string): Record<string, unknown> | undefined {
	try {
		const bytes = Uint8Array.from(atob(encoded), (char) => char.charCodeAt(0));
		const parsed: unknown = JSON.parse(new TextDecoder().decode(bytes));
		if (parsed && typeof parsed === "object") {
			return parsed as Record<string, unknown>;
		}
		return undefined;
	} catch {
		return undefined;
	}
}

function attributesTitle(attributes: Record<string, unknown> | undefined) {
	if (!attributes) {
		return undefined;
	}

	for (const field of titleFields) {
		const value = attributes[field];
		if (typeof value === "string" && value.trim()) {
			return value.trim();
		}
	}
	return undefined;
}

export function eventRowPresentation(event: Event): EventRowPresentation {
	const attrs = event.attributes;
	const identifier = attrs.resourceRef.resourceRef;
	const humanTitle = attributesTitle(decodeAttributes(attrs.attributes));

	const row: EventRowPresentation = {
		id: event.id,
		title: identifier,
		kindLabel: eventKindLabel(attrs.kind),
		occurredAt: attrs.occurredAt,
	};
	if (humanTitle) {
		row.title = humanTitle;
		row.identifier = identifier;
	}
	return row;
}
