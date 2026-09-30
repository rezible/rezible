import type { Component } from "svelte";

/** Semantic tone; maps 1:1 to the Atlas `status-*` token pairs. */
export type StatusTone = "success" | "warning" | "danger" | "neutral" | "info";

export type StatusPresentation = {
	/** Human label, sentence case, no raw enum values. */
	label: string;
	tone: StatusTone;
	/** Remix icon component; always rendered alongside the label. */
	icon: Component;
	/** Optional extra context, shown as a tooltip and appended to the accessible name. */
	description?: string;
};

/** Text colour for a tone when shown without a fill (inline badges, icons). */
export const statusToneTextClass = {
	success: "text-status-success-foreground",
	warning: "text-status-warning-foreground",
	danger: "text-status-danger-foreground",
	info: "text-status-info-foreground",
	neutral: "text-muted-foreground",
} satisfies Record<StatusTone, string>;
