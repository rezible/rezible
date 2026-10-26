import type { RetrospectiveAttributes } from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";
import { formatTime } from "$lib/time";
import RiDraftLine from "remixicon-svelte/icons/draft-line";
import RiEditLine from "remixicon-svelte/icons/edit-line";
import RiEyeLine from "remixicon-svelte/icons/eye-line";
import RiCheckboxCircleLine from "remixicon-svelte/icons/checkbox-circle-line";
import RiSkipForwardLine from "remixicon-svelte/icons/skip-forward-line";

// An admin can act as the owner, and can assign an owner to a review that has none.
export function reviewPanelModel(
	retrospective?: RetrospectiveAttributes,
	currentUserId?: string,
	isAdmin = false
) {
	const reviews = retrospective?.reviews ?? [];
	const owner = retrospective?.owner;
	const closed = retrospective?.state === "closed";
	const isOwner = !!currentUserId && owner?.id === currentUserId;
	const actsAsOwner = isOwner || (isAdmin && !!owner);
	const ownReview = reviews.find((row) => row.reviewer.id === currentUserId);
	const waiting = reviews.filter((row) => row.state !== "approved");
	let status: StatusPresentation = {
		label: "Not started",
		tone: "neutral",
		icon: RiDraftLine,
		description: "Rezible review stage",
	};
	let explanation = "The retrospective starts when the incident is resolved.";
	if (retrospective?.state === "draft") {
		status = { ...status, label: "Draft", icon: RiEditLine };
		if (owner) {
			explanation = "The owner is writing the retrospective.";
		} else {
			explanation = "No owner yet. Take ownership to start the retrospective.";
		}
	} else if (retrospective?.state === "in_review") {
		status = { ...status, label: "In review", tone: "warning", icon: RiEyeLine };
		const noun = reviews.length === 1 ? "reviewer" : "reviewers";
		explanation = `Waiting on ${waiting.length} of ${reviews.length} ${noun}.`;
	} else if (closed) {
		const closer = retrospective.closedBy?.attributes?.name ?? "a member";
		let closedOn = "";
		if (retrospective.closedAt) {
			closedOn = ` on ${formatTime(retrospective.closedAt).absolute}`;
		}
		status = { ...status, label: "Closed without review", icon: RiCheckboxCircleLine };
		explanation = `Closed by ${closer}${closedOn} without review.`;
		if (retrospective.outcome === "reviewed") {
			status = { ...status, label: "Reviewed", tone: "success" };
			explanation = `Closed by ${closer}${closedOn}, after review.`;
		} else if (retrospective.outcome === "skipped") {
			status = { ...status, label: "Skipped", icon: RiSkipForwardLine };
			explanation = `Skipped by ${closer}${closedOn}.`;
		}
	}
	return {
		status,
		explanation,
		owner,
		isOwner,
		closed,
		ownReview,
		reviewers: reviews.map((row) => {
			let label = "Waiting";
			if (row.state === "approved") {
				label = "Approved";
			} else if (row.state === "request_changes") {
				label = "Changes requested";
			}
			return { ...row, name: row.reviewer.attributes?.name ?? "Member", label };
		}),
		actions: {
			takeOwnership: !!retrospective && !closed && !owner,
			setOwner: (isOwner || isAdmin) && !!retrospective && !closed,
			requestReview: actsAsOwner && retrospective?.state === "draft",
			changeReviewers: actsAsOwner && retrospective?.state === "in_review",
			decide: !!ownReview && retrospective?.state === "in_review",
			close: actsAsOwner && !closed,
			canClose: actsAsOwner && !closed && waiting.length === 0,
			skip: actsAsOwner && !closed,
		},
		closeLabel: retrospective?.state === "draft" ? "Close without review" : "Close",
		closeReason: waiting.length
			? `Waiting on ${waiting.map((row) => row.reviewer.attributes?.name ?? "a reviewer").join(", ")}`
			: "",
	};
}

// The closed banner's second line, for example "Closed by Ava Patel after review · Tue 6 Oct, 14:32".
export function closedReviewLine(retrospective?: RetrospectiveAttributes) {
	if (retrospective?.state !== "closed") {
		return "";
	}
	const closer = retrospective.closedBy?.attributes?.name ?? "a member";
	let line = `Closed without review by ${closer}`;
	if (retrospective.outcome === "reviewed") {
		line = `Closed by ${closer} after review`;
	} else if (retrospective.outcome === "skipped") {
		line = `Skipped by ${closer}`;
	}
	if (retrospective.closedAt) {
		line += ` · ${formatTime(retrospective.closedAt).absolute}`;
	}
	if (retrospective.outcome === "skipped" && retrospective.skipReason) {
		line += `: ${retrospective.skipReason}`;
	}
	return line;
}
