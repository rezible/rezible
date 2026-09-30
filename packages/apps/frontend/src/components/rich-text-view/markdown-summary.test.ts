import { expect, test } from "bun:test";
import { markdownSummary } from "./markdown-summary";

test("uses the first paragraph after headings", () => {
	const report = [
		"# Investigation Report: Search Query Queue Backup",
		"",
		"## Summary of Situation",
		"",
		'At 2026-05-14T04:45:00Z, an alert titled **"Search query queue backing up"** was triggered.',
		"The queue depth exceeded 5,000 messages.",
		"",
		"## Incident Details",
		"- **Alert**: queue depth",
	].join("\n");

	expect(markdownSummary(report)).toBe(
		'At 2026-05-14T04:45:00Z, an alert titled "Search query queue backing up" was triggered. The queue depth exceeded 5,000 messages.'
	);
});

test("skips leading lists, code fences, tables and rules", () => {
	const markdown = [
		"- first item",
		"- second item",
		"",
		"```",
		"code block",
		"",
		"still code",
		"```",
		"",
		"| a | b |",
		"| - | - |",
		"",
		"---",
		"",
		"The real paragraph.",
	].join("\n");

	expect(markdownSummary(markdown)).toBe("The real paragraph.");
});

test("returns undefined when there is no paragraph", () => {
	expect(markdownSummary("# Title\n\n## Subtitle")).toBeUndefined();
	expect(markdownSummary("")).toBeUndefined();
});

test("strips links, images, emphasis, code and quotes", () => {
	const markdown =
		"> See [the dashboard](https://example.com) ![chart](c.png) for _latency_ and `p99` ~~old~~ values.";
	expect(markdownSummary(markdown)).toBe("See the dashboard for latency and p99 old values.");
});

test("cuts at a word boundary", () => {
	const text = "alpha beta gamma delta epsilon";
	expect(markdownSummary(text, 100)).toBe(text);
	expect(markdownSummary(text, 15)).toBe("alpha beta…");
});
