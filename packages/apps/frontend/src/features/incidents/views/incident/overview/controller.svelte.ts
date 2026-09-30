import { Context } from "runed";
import { useIncidentView } from "../controller.svelte";

class IncidentOverviewController {
	view = useIncidentView();
	incident = $derived(this.view.incident);
	retrospective = $derived(this.view.retrospective);
	retrospectiveState = $derived(this.retrospective?.attributes.state.replaceAll("_", " "));
	attributes = $derived(this.incident?.attributes);
	milestones = $derived(
		[...(this.attributes?.milestones ?? [])].sort(
			(left, right) => Date.parse(left.attributes.timestamp) - Date.parse(right.attributes.timestamp)
		)
	);
	impacts = $derived(
		(this.attributes?.impacts ?? []).map((impact) => ({
			id: impact.id,
			name: impact.knowledgeEntity.attributes?.latestState?.displayName ?? "Unavailable entity",
			note: impact.note,
			source: impact.source,
		}))
	);
	roles = $derived(
		(this.attributes?.roles ?? []).map(({ id, attributes }) => ({
			id,
			user: attributes.user.attributes?.name ?? "Unknown user",
			role: attributes.role.attributes?.name ?? "Unknown role",
		}))
	);
	ticketUrl = $derived.by(() => {
		const value = this.attributes?.externalTicket?.url;
		if (!value) return undefined;
		try {
			const url = new URL(value);
			if (url.protocol === "https:" || url.protocol === "http:") return url.href;
		} catch {
			return undefined;
		}
	});

	dateLabel(value: string | null | undefined) {
		return value ? new Date(value).toLocaleString() : "Not recorded";
	}
}

const context = new Context<IncidentOverviewController>("IncidentOverviewController");
export const initIncidentOverviewController = () => context.set(new IncidentOverviewController());
