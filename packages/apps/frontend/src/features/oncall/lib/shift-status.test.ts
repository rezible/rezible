import { expect, test } from "bun:test";
import type { OncallShift } from "$lib/api";
import { shiftStatus } from "./shift-status";

test("shift status follows the shift's window", () => {
	const shift = {
		id: "shift",
		attributes: { startAt: "2026-05-14T09:00:00Z", endAt: "2026-05-14T17:00:00Z" },
	} as unknown as OncallShift;

	expect(shiftStatus(shift, Date.parse("2026-05-14T08:59:00Z")).label).toBe("Upcoming");
	expect(shiftStatus(shift, Date.parse("2026-05-14T09:00:00Z")).label).toBe("On call");
	expect(shiftStatus(shift, Date.parse("2026-05-14T17:00:00Z")).label).toBe("Ended");
});
