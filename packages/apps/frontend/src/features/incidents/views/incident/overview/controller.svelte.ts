import { Context } from "runed";
import { getIncidentOptions, listIncidentsQueryKey, updateIncidentMutation } from "$lib/api";
import { createMutation, useQueryClient } from "@tanstack/svelte-query";
import { useIncidentView } from "../controller.svelte";

class IncidentOverviewController {
	viewController = useIncidentView();
	incident = $derived(this.viewController.incident);
	retrospective = $derived(this.viewController.retrospective);

	editing = $state(false);
	title = $state("");
	summary = $state("");
	error = $state("");
	saving = $state(false);

	queryClient = useQueryClient();
	updateMutation = createMutation(() => ({ ...updateIncidentMutation() }));
	async updateSummary(title: string, summary: string) {
		await this.updateMutation.mutateAsync({
			path: { id: this.viewController.incidentId },
			body: { attributes: { title, summary } },
		});
		await this.queryClient.invalidateQueries({
			queryKey: getIncidentOptions({ path: { id: this.viewController.slug } }).queryKey,
		});
		await this.queryClient.invalidateQueries({ queryKey: listIncidentsQueryKey() });
	}
	beginEdit = () => {
		this.title = this.incident?.attributes.title ?? "";
		this.summary = this.incident?.attributes.summary ?? "";
		this.error = "";
		this.editing = true;
	};
	save = async () => {
		if (this.saving) return;
		this.saving = true;
		try {
			await this.updateSummary(this.title, this.summary);
			this.editing = false;
		} catch (error) {
			this.error = error instanceof Error ? error.message : "Unable to save changes";
		} finally {
			this.saving = false;
		}
	};
}

const context = new Context<IncidentOverviewController>("IncidentOverviewController");
export const initIncidentOverviewController = () => context.set(new IncidentOverviewController());
