import { describe, expect, test } from "bun:test";
import {
	acceptAttempt,
	applyArtifactChunk,
	applyModelChunk,
	closeAttempt,
	emptyOverlay,
	reconcileTurn,
} from "./model";
import type { AgentModelChunk } from "$lib/api";

const old = "2026-10-07T00:00:00.123456Z";
const current = "2026-10-07T00:00:00.123457Z";
const text = (value: string): AgentModelChunk => ({
	aggregated: false,
	index: 0,
	role: "model",
	parts: [{ kind: "text", text: value }],
});
const content = (overlay: ReturnType<typeof emptyOverlay>) =>
	overlay.models
		.get("turn:0")
		?.parts.map((part) => part.text)
		.join("");
const artifact = { name: "report", parts: [{ kind: "text" as const, text: "preview" }] };

describe("stream attempt ordering", () => {
	test("chunks before running survive the delayed status and keep their prefix", () => {
		let overlay = applyModelChunk(emptyOverlay(), "turn", old, text("old"));
		overlay = applyArtifactChunk(overlay, "turn", old, artifact);
		overlay = applyModelChunk(overlay, "turn", current, text("new "));
		expect(overlay.artifacts.size).toBe(0);
		overlay = applyArtifactChunk(overlay, "turn", current, artifact);
		overlay = acceptAttempt(overlay, "turn", current)!;
		overlay = applyModelChunk(overlay, "turn", current, text("output"));
		expect(content(overlay)).toBe("new output");
		expect(overlay.artifacts.size).toBe(1);
	});
	test("running before chunks clears only that turn's previous attempt", () => {
		let overlay = applyModelChunk(emptyOverlay(), "turn", old, text("old"));
		overlay = applyArtifactChunk(overlay, "other", old, artifact);
		overlay = acceptAttempt(overlay, "turn", current)!;
		expect(content(overlay)).toBeUndefined();
		expect(overlay.artifacts.size).toBe(1);
		overlay = applyModelChunk(overlay, "turn", current, text("new"));
		expect(content(overlay)).toBe("new");
	});
	test("old chunks and terminal status cannot change a newer attempt", () => {
		const overlay = applyModelChunk(emptyOverlay(), "turn", current, text("new"));
		expect(applyModelChunk(overlay, "turn", old, text("stale"))).toBe(overlay);
		expect(applyArtifactChunk(overlay, "turn", old, artifact)).toBe(overlay);
		expect(acceptAttempt(overlay, "turn", old)).toBeNull();
		expect(closeAttempt(overlay, "turn", old)).toBe(overlay);
		expect(reconcileTurn(overlay, "turn", old)).toBe(overlay);
	});
	test("terminal status suppresses late chunks even after reconciliation", () => {
		let overlay = closeAttempt(emptyOverlay(), "turn", current);
		overlay = reconcileTurn(overlay, "turn", current);
		overlay = acceptAttempt(overlay, "turn", current)!;
		expect(applyModelChunk(overlay, "turn", current, text("late"))).toBe(overlay);
		expect(applyArtifactChunk(overlay, "turn", current, artifact)).toBe(overlay);
	});
	test("a late refetch cannot reconcile away a newer attempt", () => {
		let overlay = closeAttempt(emptyOverlay(), "turn", old);
		overlay = applyArtifactChunk(overlay, "turn", current, artifact);
		expect(reconcileTurn(overlay, "turn", old)).toBe(overlay);
		expect(overlay.reconcilingTurns.has("turn")).toBe(false);
	});
	test("equivalent timestamp encodings identify the same attempt", () => {
		const overlay = applyModelChunk(emptyOverlay(), "turn", old, text("prefix"));
		expect(acceptAttempt(overlay, "turn", "2026-10-07T08:00:00.123456000+08:00")).toBe(overlay);
	});
});
