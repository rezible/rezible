import { createMutation, useQueryClient } from "@tanstack/svelte-query";
import { Context } from "runed";
import {
	setRetrospectiveOwnerMutation,
	requestRetrospectiveReviewMutation,
	addRetrospectiveReviewerMutation,
	removeRetrospectiveReviewerMutation,
	setRetrospectiveReviewDecisionMutation,
	closeRetrospectiveMutation,
	skipRetrospectiveMutation,
	getRetrospectiveQueryKey,
	type ApiError,
	type User,
} from "$lib/api";
import { useUserSessionState } from "$lib/user-session.svelte";
import { reviewPanelModel } from "$features/incidents/lib/review";
import { useIncidentView } from "../controller.svelte";

type ReviewDialog = "owner" | "request" | "add" | "remove" | "close" | "skip";

export class ReviewPanelController {
	private view = useIncidentView();
	private session = useUserSessionState();
	private queryClient = useQueryClient();
	model = $derived(reviewPanelModel(this.view.retrospective?.attributes, this.session.user?.id, this.session.isAdmin));
	skipReason = $derived(this.view.retrospective?.attributes.skipReason);
	dialog = $state<ReviewDialog>();
	selected = $state<User[]>([]);
	removeUserId = $state("");
	removeName = $state("");
	reason = $state("");
	feedback = $state("");
	showFeedback = $state(false);
	error = $state<ApiError>();
	pending = $state(false);
	reviewerIds = $derived(this.model.reviewers.map((row) => row.reviewer.id));
	excludeIds = $derived(this.buildExcluded());
	selectedIds = $derived(this.selected.map((user) => user.id));
	selectedNames = $derived(this.selected.map((user) => user.attributes.name).join(", "));
	dialogTitle = $derived(this.title());

	private ownerMutation = createMutation(() => setRetrospectiveOwnerMutation());
	private requestMutation = createMutation(() => requestRetrospectiveReviewMutation());
	private addMutation = createMutation(() => addRetrospectiveReviewerMutation());
	private removeMutation = createMutation(() => removeRetrospectiveReviewerMutation());
	private decisionMutation = createMutation(() => setRetrospectiveReviewDecisionMutation());
	private closeMutation = createMutation(() => closeRetrospectiveMutation());
	private skipMutation = createMutation(() => skipRetrospectiveMutation());

	private buildExcluded() {
		const ownerId = this.model.owner?.id;
		const ids = ownerId ? [ownerId] : [];
		if (this.dialog === "owner" || this.dialog === "add") {
			return [...ids, ...this.reviewerIds];
		}
		return ids;
	}

	private title() {
		switch (this.dialog) {
			case "owner":
				return "Hand over this retrospective";
			case "request":
				return "Request review";
			case "add":
				return "Add reviewer";
			case "remove":
				return `Remove ${this.removeName} as a reviewer?`;
			case "close":
				return `${this.model.closeLabel}?`;
			case "skip":
				return "Skip this retrospective";
			default:
				return "Review";
		}
	}

	private async run(action: (id: string) => Promise<unknown>) {
		const id = this.view.retrospectiveId;
		if (!id || this.pending) {
			return;
		}
		this.pending = true;
		this.error = undefined;
		try {
			await action(id);
			await this.queryClient.invalidateQueries({
				queryKey: getRetrospectiveQueryKey({ path: { id } }),
			});
			this.dialog = undefined;
			this.showFeedback = false;
		} catch (error) {
			this.error = error as ApiError;
		} finally {
			this.pending = false;
		}
	}

	open = (dialog: ReviewDialog) => {
		this.dialog = dialog;
		this.selected = [];
		this.reason = "";
		this.error = undefined;
	};
	cancel = () => {
		if (!this.pending) {
			this.dialog = undefined;
		}
	};
	select = (user: User) => {
		if (this.dialog !== "request") {
			this.selected = [user];
			return;
		}
		if (this.selectedIds.includes(user.id)) {
			this.selected = this.selected.filter((row) => row.id !== user.id);
		} else {
			this.selected = [...this.selected, user];
		}
	};
	remove = (userId: string, name: string) => {
		this.open("remove");
		this.removeUserId = userId;
		this.removeName = name;
	};
	take = () => {
		const userId = this.session.user?.id;
		if (!userId) return;
		this.run((id) => this.ownerMutation.mutateAsync({ path: { id }, body: { attributes: { userId } } }));
	};
	approve = () => this.decide("approved");
	requestChanges = () => {
		this.feedback = this.model.ownReview?.feedback ?? "";
		this.showFeedback = true;
	};
	sendFeedback = () => this.decide("request_changes");
	private decide(decision: "approved" | "request_changes") {
		return this.run((id) =>
			this.decisionMutation.mutateAsync({
				path: { id },
				body: { attributes: { decision, feedback: this.feedback } },
			})
		);
	}
	submit = () =>
		this.run(async (id) => {
			const path = { id };
			switch (this.dialog) {
				case "owner":
					return this.ownerMutation.mutateAsync({
						path,
						body: { attributes: { userId: this.selected[0].id } },
					});
				case "request":
					return this.requestMutation.mutateAsync({
						path,
						body: { attributes: { reviewerIds: this.selectedIds } },
					});
				case "add":
					return this.addMutation.mutateAsync({
						path,
						body: { attributes: { userId: this.selected[0].id } },
					});
				case "remove":
					return this.removeMutation.mutateAsync({ path: { id, userId: this.removeUserId } });
				case "close":
					return this.closeMutation.mutateAsync({ path });
				case "skip":
					return this.skipMutation.mutateAsync({
						path,
						body: { attributes: { reason: this.reason } },
					});
			}
		});
	canSubmit = $derived(
		!this.pending &&
			((this.dialog === "skip" && !!this.reason.trim()) ||
				(this.dialog === "close" && this.model.actions.canClose) ||
				this.dialog === "remove" ||
				(["owner", "request", "add"].includes(this.dialog ?? "") && this.selected.length > 0))
	);
}

const ctx = new Context<ReviewPanelController>("ReviewPanelController");
export const initReviewPanel = () => ctx.set(new ReviewPanelController());
