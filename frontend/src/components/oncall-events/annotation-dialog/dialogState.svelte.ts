import { createOncallAnnotationMutation, type OncallEvent, updateOncallAnnotationMutation, type OncallAnnotation, type OncallAnnotationAlertFeedback, type OncallRoster, getUserOncallInformationOptions } from "$src/lib/api";
import { session } from "$src/lib/auth.svelte";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { Context } from "runed";
import { SvelteSet } from "svelte/reactivity";

class EventAnnotationAttributesState {
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
			accurate: this.alertAccuracy,
			documentationAvailable,
			actionable: this.alertRequiredAction,
		})
	}
}

type OnCloseFn = (updated?: OncallAnnotation) => void
type DialogOptions = {
	onClosed?: OnCloseFn;
}
export class AnnotationDialogState {
	event = $state<OncallEvent>();
	annotation = $state<OncallAnnotation>();

	attributes = new EventAnnotationAttributesState();
	open = $derived(!!this.event);

	oncallInfoQuery = createQuery(() => getUserOncallInformationOptions({
		query: { 
			userId: session.userId,
			activeShifts: true,
		}
	}));
	userOncallInfo = $derived(this.oncallInfoQuery.data?.data);
	userRosters = $derived(this.userOncallInfo?.rosters || []);
	userActiveShifts = $derived(this.userOncallInfo?.activeShifts || []);

	allowCreating = $state(true);
	// TODO: allow selecting roster
	annotatableRosterIds = $derived(new Set(this.userRosters.map(r => r.id)));
	canCreate(curr: OncallAnnotation[]) {
		if (this.annotatableRosterIds.size === 0) return false;
		if (curr.length === 0) return true;
		const currIds = new Set(curr.map(a => a.attributes.roster.id));
		return !currIds.isSupersetOf(this.annotatableRosterIds);
	}

	view = $derived.by(() => {
		if (!this.event) return false;
		if (this.annotation?.attributes.creator.id === session.userId) return "edit";
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

	setOpen(event: OncallEvent, anno?: OncallAnnotation) {
		this.event = event;
		this.annotation = anno;
		this.attributes.setup(event, anno);
	}

	onClose(updated?: OncallAnnotation) {
		this.event = undefined;
		this.annotation = undefined;
		if (this.onCloseFn) this.onCloseFn(updated)
	}

	createMut = createMutation(() => ({
		...createOncallAnnotationMutation(),
		onSuccess: ({ data }) => { this.onClose(data) }
	}));

	updateMut = createMutation(() => ({
		...updateOncallAnnotationMutation(),
		onSuccess: ({ data }) => { this.onClose(data) }
	}));

	onConfirm() {
		const rosterId = this.annotatableRosterIds.values().next().value;
		if (!this.event || !rosterId) return;

		let alertFeedback: OncallAnnotationAlertFeedback | undefined = undefined;
		if (this.event.attributes.kind === "alert") {
			alertFeedback = this.attributes.getAlertFeedback();
		}
		const attributes = $state.snapshot({
			eventId: this.event.id,
			rosterId: rosterId,
			minutesOccupied: 0,
			notes: this.attributes.notes,
			tags: this.attributes.tags.values().toArray(),
			alertFeedback,
		})
		if (this.annotation) {
			this.updateMut.mutate({ path: { id: this.annotation.id }, body: { attributes } });
		} else {
			this.createMut.mutate({ body: { attributes } })
		}
	}
}

const annotationDialogStateCtx = new Context<AnnotationDialogState>("annotationDialogState");
export const setAnnotationDialogState = (s: AnnotationDialogState) => annotationDialogStateCtx.set(s);
export const useAnnotationDialogState = () => annotationDialogStateCtx.get();