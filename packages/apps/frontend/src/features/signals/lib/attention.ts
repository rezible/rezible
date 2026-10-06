import type { AlertSituationSignalAttention } from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";
import RiEyeLine from "remixicon-svelte/icons/eye-line";
import RiLinksLine from "remixicon-svelte/icons/links-line";

export type AttentionLevel = AlertSituationSignalAttention["level"];

export type AttentionLevelOption = { value: AttentionLevel; label: string; help: string };

export const attentionLevelOptions: AttentionLevelOption[] = [
	{
		value: "default",
		label: "Default",
		help: "Can raise a situation on its own when a reason applies.",
	},
	{
		value: "watch_only",
		label: "Watch only",
		help: "Starts watching, but needs another alert source before it can raise automatically.",
	},
	{
		value: "join_only",
		label: "Join only",
		help: "Only joins an existing situation; does not count toward automatic raising.",
	},
];

export function attentionLevelOption(level: AttentionLevel) {
	return attentionLevelOptions.find((option) => option.value === level) ?? attentionLevelOptions[0];
}

/** A neutral badge for a definition that cannot raise on its own; none for the default level. */
export function attentionStatus(
	attention: AlertSituationSignalAttention | undefined
): StatusPresentation | undefined {
	switch (attention?.level) {
		case "watch_only":
			return { label: "Watch only", tone: "neutral", icon: RiEyeLine };
		case "join_only":
			return { label: "Join only", tone: "neutral", icon: RiLinksLine };
		default:
			return undefined;
	}
}
