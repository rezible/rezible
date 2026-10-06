import {
	type ErrorModel,
	getAlertOptions,
	getAlertQueryKey,
	listAlertsQueryKey,
	listSituationsQueryKey,
	setAlertIdentityGroupLabelsMutation,
	setAlertSituationSignalAttentionMutation,
} from "$lib/api";
import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";
import { toast } from "svelte-sonner";
import type { AttentionLevel } from "$features/signals/lib/attention";

/** Every situation detail and alert episode query; their records carry the definition's attention badge. */
const situationDetailQueriesKey = [{ _id: "getSituation" }];
const situationEpisodeQueriesKey = [{ _id: "listSituationAlertEpisodes" }];

export class AlertViewController {
	private queryClient = useQueryClient();

	alertId = $state<string>(null!);

	alertQuery = createQuery(() => getAlertOptions({ path: { id: this.alertId } }));
	alert = $derived(this.alertQuery.data?.data);
	alertTitle = $derived(this.alert?.attributes.title ?? "");

	attention = $derived(this.alert?.attributes.situationSignalAttention);
	identityGroupLabels = $derived(this.alert?.attributes.identityGroupLabels ?? []);

	attentionMutation = createMutation(() => ({
		...setAlertSituationSignalAttentionMutation(),
		onSuccess: (response) => this.refreshAfterChange(response.data.id),
		onError: (error: ErrorModel) => {
			toast.error("Could not change the signal attention", { description: error.detail });
		},
	}));

	labelsMutation = createMutation(() => ({
		...setAlertIdentityGroupLabelsMutation(),
		onSuccess: async (response) => {
			const definition = response.data;
			// The page may have moved to another definition while this one saved.
			if (definition.id === this.alertId) {
				this.labelsInput = definition.attributes.identityGroupLabels.join(", ");
			}
			await this.refreshAfterChange(definition.id);
		},
		onError: (error: ErrorModel) => {
			toast.error("Could not change the identity group labels", { description: error.detail });
		},
	}));

	/** The level shown while a change is in flight. */
	attentionLevel = $derived(this.getAttentionLevel());

	labelsInput = $state("");
	parsedLabels = $derived(parseLabels(this.labelsInput));
	labelsChanged = $derived(this.parsedLabels.join("\n") !== this.identityGroupLabels.join("\n"));

	constructor(idFn: Getter<string>) {
		this.alertId = idFn();
		watch(idFn, (id) => {
			this.alertId = id;
		});
		watch(
			() => this.identityGroupLabels.join(", "),
			(labels) => {
				this.labelsInput = labels;
			}
		);
	}

	setAttentionLevel = (value: string) => {
		const level = value as AttentionLevel;
		if (this.attentionMutation.isPending || level === this.attention?.level) {
			return;
		}
		this.attentionMutation.mutate({ path: { id: this.alertId }, body: { attributes: { level } } });
	};

	setLabelsInput = (value: string) => {
		this.labelsInput = value;
	};

	saveLabels = () => {
		this.setLabels(this.parsedLabels);
	};

	clearLabels = () => {
		this.setLabels([]);
	};

	retryAlert = () => {
		this.alertQuery.refetch();
	};

	private getAttentionLevel(): AttentionLevel {
		const pending = this.attentionMutation.variables;
		if (this.attentionMutation.isPending && pending?.path.id === this.alertId) {
			return pending.body.attributes.level;
		}
		return this.attention?.level ?? "default";
	}

	private setLabels(labels: string[]) {
		if (this.labelsMutation.isPending) {
			return;
		}
		this.labelsMutation.mutate({ path: { id: this.alertId }, body: { attributes: { labels } } });
	}

	/** Refetches the changed definition and situation queries; the backend re-evaluates affected candidates itself. */
	private async refreshAfterChange(definitionId: string) {
		await Promise.all([
			this.queryClient.invalidateQueries({
				queryKey: getAlertQueryKey({ path: { id: definitionId } }),
			}),
			this.queryClient.invalidateQueries({ queryKey: listAlertsQueryKey() }),
			this.queryClient.invalidateQueries({ queryKey: listSituationsQueryKey() }),
			this.queryClient.invalidateQueries({ queryKey: situationDetailQueriesKey }),
			this.queryClient.invalidateQueries({ queryKey: situationEpisodeQueriesKey }),
		]);
	}
}

/** Comma- or newline-separated label names, trimmed, without blanks or repeats. */
function parseLabels(input: string) {
	const labels = new Set<string>();
	for (const part of input.split(/[,\n]/)) {
		const label = part.trim();
		if (label) {
			labels.add(label);
		}
	}
	return [...labels];
}

const ctx = new Context<AlertViewController>("alertView");
export const initAlertViewController = (idFn: Getter<string>) => ctx.set(new AlertViewController(idFn));
export const useAlertViewController = () => ctx.get();
