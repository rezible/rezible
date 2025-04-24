import { type CreateOncallAnnotationRequestAttributes, type CreateOncallAnnotationRequestBody, type OncallAnnotation, type OncallAnnotationAlertFeedback, type OncallEvent, type UpdateOncallAnnotationRequestBody } from "$lib/api";
import { SvelteSet } from "svelte/reactivity";

export class EventAnnotationAttributesState {
	notes = $state("");
	alertAccuracy = $state<OncallAnnotationAlertFeedback["accurate"]>("yes");
	alertRequiredAction = $state(true);
	alertDocs = $state(true);
	alertDocsNeedUpdate = $state(false);
	tags = $state(new SvelteSet<string>());
	draftTag = $state("");

	setup(ev: OncallEvent, anno?: OncallAnnotation) {
		this.notes = anno?.attributes.notes ?? "";
		const alertFb = anno?.attributes.alertFeedback;
		this.alertAccuracy = alertFb?.accurate ?? "yes";
		this.alertRequiredAction = alertFb?.actionable ?? true;
		const docs = alertFb?.documentationAvailable ?? "yes";
		this.alertDocs = docs !== "no";
		this.alertDocsNeedUpdate = docs === "needs_update";
		this.tags = new SvelteSet(anno?.attributes.tags ?? []);
		this.draftTag = "";
	}

	addDraftTag() {
		this.tags.add($state.snapshot(this.draftTag));
		this.draftTag = "";
	}

	removeTag(tag: string) {
		this.tags.delete(tag)
	}

	getAlertFeedback(): OncallAnnotationAlertFeedback {
		const documentationAvailable = !this.alertDocs ? "no" : (this.alertDocsNeedUpdate ? "needs_update" : "yes");
		return $state.snapshot({
			accurate: attributesState.alertAccuracy,
			documentationAvailable,
			actionable: this.alertRequiredAction,
		})
	}
}
export const attributesState = new EventAnnotationAttributesState();