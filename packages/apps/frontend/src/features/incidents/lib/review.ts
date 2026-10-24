import { formatDate } from "date-fns";
import type { RetrospectiveAttributes } from "$lib/api";
import type { StatusPresentation } from "$components/common/status-badge/status";
import RiDraftLine from "remixicon-svelte/icons/draft-line";
import RiEditLine from "remixicon-svelte/icons/edit-line";
import RiEyeLine from "remixicon-svelte/icons/eye-line";
import RiCheckboxCircleLine from "remixicon-svelte/icons/checkbox-circle-line";
import RiSkipForwardLine from "remixicon-svelte/icons/skip-forward-line";

export function reviewPanelModel(retrospective?: RetrospectiveAttributes, currentUserId?: string) {
	const reviews = retrospective?.reviews ?? [];
	const owner = retrospective?.owner;
	const closed = retrospective?.state === "closed";
	const isOwner = !!currentUserId && owner?.id === currentUserId;
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
		explanation = "The owner is writing the retrospective.";
	} else if (retrospective?.state === "in_review") {
		status = { ...status, label: "In review", tone: "warning", icon: RiEyeLine };
		const noun = reviews.length === 1 ? "reviewer" : "reviewers";
		explanation = `Waiting on ${waiting.length} of ${reviews.length} ${noun}.`;
	} else if (closed) {
		const closer = retrospective.closedBy?.attributes?.name ?? "a member";
		const date = retrospective.closedAt ? formatDate(retrospective.closedAt, "d MMM") : "";
		status = { ...status, label: "Closed without review", icon: RiCheckboxCircleLine };
		explanation = `Closed by ${closer} on ${date} without review.`;
		if (retrospective.outcome === "reviewed") {
			status = { ...status, label: "Reviewed", tone: "success" };
			explanation = `Closed by ${closer} on ${date}, after review.`;
		} else if (retrospective.outcome === "skipped") {
			status = { ...status, label: "Skipped", icon: RiSkipForwardLine };
			explanation = `Skipped by ${closer} on ${date}.`;
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
			setOwner: isOwner && !closed,
			requestReview: isOwner && retrospective?.state === "draft",
			changeReviewers: isOwner && retrospective?.state === "in_review",
			decide: !!ownReview && retrospective?.state === "in_review",
			close: isOwner && !closed,
			canClose: isOwner && !closed && waiting.length === 0,
			skip: isOwner && !closed,
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
		line += ` · ${formatDate(retrospective.closedAt, "EEE d MMM, HH:mm")}`;
	}
	if (retrospective.outcome === "skipped" && retrospective.skipReason) {
		line += `: ${retrospective.skipReason}`;
	}
	return line;
}
