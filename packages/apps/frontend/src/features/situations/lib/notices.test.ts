import { expect, test } from "bun:test";
import type { SituationNotice } from "$lib/api";
import { noticeLines, noticeRatingsText } from "./notices";

const notice = (attributes: Partial<SituationNotice>): SituationNotice => ({
	channelId: "C0AAA1",
	channelName: "incidents",
	status: "pending",
	...attributes,
});

test("a posted notice links to its message", () => {
	const permalink = "https://example.slack.com/archives/C0AAA1/p1";

	const [line] = noticeLines([notice({ status: "posted", permalink })]);

	expect(line).toMatchObject({ kind: "posted", channel: "#incidents", href: permalink, retryable: false });
});

test("a posted notice without a known link is not a link", () => {
	const [line] = noticeLines([notice({ status: "posted" })]);

	expect(line.href).toBeUndefined();
});

test("a pending notice is being posted", () => {
	const [line] = noticeLines([notice({ status: "pending" })]);

	expect(line).toMatchObject({ kind: "posting", text: "Posting to #incidents…", retryable: false });
});

test("a failed notice gives its error and can be retried", () => {
	const lines = noticeLines([notice({ status: "failed", error: "not_in_channel" })]);

	expect(lines[0]).toMatchObject({
		kind: "not-posted",
		text: "Couldn't post to #incidents: not_in_channel",
		retryable: true,
	});
});

test("a failed update of a posted notice keeps its link and can be retried", () => {
	const permalink = "https://example.slack.com/archives/C0AAA1/p1";

	const [line] = noticeLines([notice({ status: "failed", error: "ratelimited", permalink })]);

	expect(line).toMatchObject({
		kind: "posted",
		channel: "#incidents",
		href: permalink,
		text: "couldn't update: ratelimited",
		retryable: true,
	});
});

test("a skipped notice is shown like a failed one, without retry", () => {
	const lines = noticeLines([notice({ status: "skipped", error: "hourly notice limit reached" })]);

	expect(lines[0]).toMatchObject({
		kind: "not-posted",
		text: "Couldn't post to #incidents: hourly notice limit reached",
		retryable: false,
	});
});

test("a channel whose name is not known shows its ID", () => {
	const [line] = noticeLines([notice({ channelName: undefined })]);

	expect(line.channel).toBe("C0AAA1");
});

test("ratings are shown once anyone has rated", () => {
	expect(noticeRatingsText(undefined)).toBeUndefined();
	expect(noticeRatingsText({ useful: 0, notUseful: 0 })).toBeUndefined();
	const text = noticeRatingsText({ useful: 2, notUseful: 1 });
	expect(text).toContain("useful by 2");
	expect(text).toContain("not useful by 1");
});
