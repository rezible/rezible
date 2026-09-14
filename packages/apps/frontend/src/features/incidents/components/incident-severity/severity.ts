import type { IncidentSeverity } from "$lib/api";

export const incidentPriorityClasses = {
	danger: "bg-status-danger/24 inset-shadow-[3px_0_var(--color-status-danger-foreground)]",
	warning: "bg-status-warning/24 inset-shadow-[3px_0_var(--color-status-warning-foreground)]",
	neutral: "bg-status-neutral/24 inset-shadow-[3px_0_var(--color-status-neutral-foreground)]",
} satisfies Record<ReturnType<typeof incidentSeverityVariant>, string>;

export function incidentSeverityVariant(severity?: IncidentSeverity) {
	if (!severity) return "neutral";
	return severity.attributes.rank <= 1 ? "danger" : severity.attributes.rank === 2 ? "warning" : "neutral";
}
