import type { SituationNotice, SituationNoticeRatings } from "$lib/api";

/** One chat notice of the situation, as its line on the situation page shows it. */
export type NoticeLine = {
	kind: "posted" | "posting" | "not-posted";
	/** The channel as people know it: `#name`, or its ID when the name is not known. */
	channel: string;
	/** Where the posted message is, when its link is known. */
	href?: string;
	/**
	 * What the line says while posting, or why the notice was not posted. On a posted line, why its last
	 * update failed, if it did.
	 */
	text?: string;
	/** Whether delivery can be tried again. A notice skipped by the hourly limit cannot. */
	retryable: boolean;
};

function channelLabel(notice: SituationNotice) {
	if (!notice.channelName) {
		return notice.channelId;
	}
	return `#${notice.channelName}`;
}

function noticeLine(notice: SituationNotice): NoticeLine {
	const line: NoticeLine = {
		kind: "posting",
		channel: channelLabel(notice),
		retryable: false,
	};

	if (notice.status === "posted") {
		line.kind = "posted";
		line.href = notice.permalink || undefined;
		return line;
	}

	if (notice.status === "pending") {
		line.text = `Posting to ${line.channel}…`;
		return line;
	}

	// A notice skipped by the hourly limit stays skipped.
	line.retryable = notice.status === "failed";

	// A failed notice with a link was posted, then an update to it failed.
	if (notice.status === "failed" && notice.permalink) {
		line.kind = "posted";
		line.href = notice.permalink;
		line.text = withReason("couldn't update", notice.error);
		return line;
	}

	line.kind = "not-posted";
	line.text = withReason(`Couldn't post to ${line.channel}`, notice.error);
	return line;
}

function withReason(text: string, error: string | undefined) {
	if (!error) {
		return text;
	}
	return `${text}: ${error}`;
}

export function noticeLines(notices: SituationNotice[] | undefined): NoticeLine[] {
	return (notices ?? []).map(noticeLine);
}

/** The ratings people gave the notice, or undefined before anyone has rated it. */
export function noticeRatingsText(ratings: SituationNoticeRatings | undefined) {
	if (!ratings || ratings.useful + ratings.notUseful === 0) {
		return undefined;
	}
	return `Rated useful by ${ratings.useful}, not useful by ${ratings.notUseful}`;
}
