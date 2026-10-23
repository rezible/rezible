import { createMutation } from "@tanstack/svelte-query";
import type { Getter } from "runed";
import { toast } from "svelte-sonner";

import { errorDisplay, retrySituationNoticeMutation, type Situation } from "$lib/api";
import { type NoticeLine, noticeLines, noticeRatingsText } from "$features/situations/lib/notices";

/** The situation's chat notices, and retrying the ones that failed. */
export class SituationNoticesController {
	private situation: Getter<Situation | undefined>;
	private applySituation: (situation: Situation) => Promise<void>;

	constructor(
		situation: Getter<Situation | undefined>,
		applySituation: (situation: Situation) => Promise<void>
	) {
		this.situation = situation;
		this.applySituation = applySituation;
		this.lines = $derived(noticeLines(situation()?.attributes.notices));
		this.ratings = $derived(noticeRatingsText(situation()?.attributes.noticeRatings));
	}

	lines: NoticeLine[];
	ratings: string | undefined;

	retryMutation = createMutation(() => ({
		...retrySituationNoticeMutation(),
		onSuccess: (response) => this.applySituation(response.data),
		onError: (error) => {
			const display = errorDisplay(error);
			toast.error(display.title, { description: display.detail });
		},
	}));
	retrying = $derived(this.retryMutation.isPending);

	retry = () => {
		const id = this.situation()?.id;
		if (!id || this.retrying) {
			return;
		}
		this.retryMutation.mutate({ path: { id } });
	};
}
