import { expect, test } from "bun:test";
import { formatDuration, formatTime } from "./time";

const now = Date.parse("2026-05-14T12:00:00Z");
const before = (ms: number) => new Date(now - ms).toISOString();

const SECOND = 1000;
const MINUTE = 60 * SECOND;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

test("formatTime treats missing, empty, unparsable and zero times as unavailable", () => {
	for (const value of [undefined, null, "", "not a date", "0001-01-01T00:00:00Z"]) {
		const formatted = formatTime(value, now);
		expect(formatted.iso).toBeUndefined();
		expect(formatted.epochMs).toBeUndefined();
		expect(formatted.absolute).toBe("Time unavailable");
		expect(formatted.relative).toBe("Time unavailable");
		expect(formatted.clock).toBe("—");
	}
});

test("formatTime returns ISO and epoch values for valid input", () => {
	const formatted = formatTime("2026-05-14T04:45:00Z", now);
	expect(formatted.iso).toBe("2026-05-14T04:45:00.000Z");
	expect(formatted.epochMs).toBe(Date.parse("2026-05-14T04:45:00Z"));
	expect(formatted.absolute).not.toBe("Time unavailable");
});

test("formatTime relative thresholds", () => {
	expect(formatTime(before(44 * SECOND), now).relative).toBe("just now");
	expect(formatTime(before(5 * MINUTE), now).relative).toContain("5 minutes");
	expect(formatTime(before(59 * MINUTE), now).relative).toContain("59 minutes");
	expect(formatTime(before(3 * HOUR), now).relative).toContain("3 hours");
	expect(formatTime(before(4 * DAY), now).relative).toContain("4 days");
	expect(formatTime(before(40 * DAY), now).relative).not.toContain("ago");
});

test("formatDuration units", () => {
	const start = "2026-05-14T00:00:00Z";
	const at = (ms: number) => new Date(Date.parse(start) + ms).toISOString();

	expect(formatDuration(start, at(45 * SECOND))).toBe("45 s");
	expect(formatDuration(start, at(23 * MINUTE))).toBe("23 min");
	expect(formatDuration(start, at(4 * HOUR + 5 * MINUTE))).toBe("4 h 5 min");
	expect(formatDuration(start, at(4 * HOUR))).toBe("4 h");
	expect(formatDuration(start, at(3 * DAY + 2 * HOUR))).toBe("3 d 2 h");
	expect(formatDuration(start, at(3 * DAY))).toBe("3 d");
});

test("formatDuration rejects invalid ranges and defaults end to now", () => {
	expect(formatDuration(undefined, undefined, now)).toBeUndefined();
	expect(formatDuration("2026-05-14T12:00:00Z", "2026-05-14T11:00:00Z")).toBeUndefined();
	expect(formatDuration(before(30 * MINUTE), null, now)).toBe("30 min");
});
