import { useAuthSessionState } from "$lib/auth.svelte";
import { createEventAnnotationMutation, updateEventAnnotationMutation, type Event, type EventAnnotation } from "$lib/api";
import { useUserOncallInformation } from "$lib/userOncall.svelte";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";
import { SvelteSet } from "svelte/reactivity";

class EventAnnotationAttributesState {
	notes = $state("");
	tags = $state(new SvelteSet<string>());
	draftTag = $state("");

	// TODO
	// alertAccuracy = $state<AlertFeedbackInstance["accurate"]>("yes");
	// alertRequiredAction = $state(true);
	// alertDocs = $state(true);
	// alertDocsNeedUpdate = $state(false);

	setup(ev: Event, anno?: EventAnnotation) {
		this.notes = anno?.attributes.notes ?? "";
		// const fb = anno?.attributes.alertFeedback;
		// this.alertAccuracy = fb?.accurate ?? "yes";
		// this.alertRequiredAction = fb?.actionable ?? true;
		// this.alertDocs = fb?.documentationAvailable ?? true;
		// this.alertDocsNeedUpdate = fb?.documentationNeedsUpdate ?? false;
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

	// getAlertFeedback(): AlertFeedbackInstance {
	// 	return $state.snapshot({
	// 		accurate: this.alertAccuracy,
	// 		documentationAvailable: this.alertDocs,
	// 		documentationNeedsUpdate: this.alertDocsNeedUpdate,
	// 		actionable: this.alertRequiredAction,
	// 	})
	// }
}

type OnCloseFn = (updated?: EventAnnotation) => void
type DialogOptions = {
	onClosed?: OnCloseFn;
}
export class AnnotationDialogState {
	private session = useAuthSessionState();

	event = $state<Event>();
	annotation = $state<EventAnnotation>();

	attributes = new EventAnnotationAttributesState();
	open = $derived(!!this.event);

	userOncallInfo = useUserOncallInformation();
	userRosters = $derived(this.userOncallInfo?.rosters || []);
	userActiveShifts = $derived(this.userOncallInfo?.activeShifts || []);

	allowCreating = $state(true);
	// TODO: allow selecting roster
	annotatableRosterIds = $derived(new Set(this.userRosters.map(r => r.id)));

	view = $derived.by(() => {
		if (!this.event) return false;
		if (this.annotation?.attributes.creator.id === this.session.user?.id) return "edit";
		if (!this.annotation && this.allowCreating) {
			if (this.userActiveShifts.length > 0) return "create";
			return false; // shouldn't happen??
		}
		return "view";
	});

	onCloseFn: OnCloseFn | undefined;

	constructor(opts: DialogOptions) {
		this.onCloseFn = opts.onClosed;
	}

	setOpen(event: Event, anno?: EventAnnotation) {
		this.event = event;
		this.annotation = anno;
		this.attributes.setup(event, anno);
	}

	onClose(updated?: EventAnnotation) {
		this.event = undefined;
		this.annotation = undefined;
		if (this.onCloseFn) this.onCloseFn(updated)
	}

	createMut = createMutation(() => ({
		...createEventAnnotationMutation(),
		onSuccess: ({ data }) => { this.onClose(data) }
	}));

	updateMut = createMutation(() => ({
		...updateEventAnnotationMutation(),
		onSuccess: ({ data }) => { this.onClose(data) }
	}));

	onConfirm() {
		if (!this.event) return;
		// let alertFeedback: AlertFeedbackInstance | undefined = undefined;
		// if (this.event.attributes.kind === "alert") {
		// 	alertFeedback = this.attributes.getAlertFeedback();
		// }
		const attributes = $state.snapshot({
			minutesOccupied: 0,
			notes: this.attributes.notes,
			tags: this.attributes.tags.values().toArray(),
			// alertFeedback,
		})
		if (this.annotation) {
			this.updateMut.mutate({ path: { id: this.annotation.id }, body: { attributes } });
		} else {
			this.createMut.mutate({
				body: {
					attributes: {
						...attributes,
						eventId: this.event.id,
					}
				}
			})
		}
	}
}

const annotationDialogStateCtx = new Context<AnnotationDialogState>("annotationDialogState");
export const setAnnotationDialogState = (s: AnnotationDialogState) => annotationDialogStateCtx.set(s);
export const useAnnotationDialogState = () => annotationDialogStateCtx.get();