import { describe, expect, test } from "bun:test";
import type { MarkSpec, NodeSpec, SchemaSpec } from "@tiptap/pm/model";
import { emptyDocument, transformSchemaSpec } from "./transformer";

const createSpec = (input: {
	nodes?: Record<string, NodeSpec>;
	marks?: Record<string, MarkSpec>;
	topNode?: string;
}) => ({
	nodes: {
		forEach: (cb: (name: string, spec: NodeSpec) => void) => {
			Object.entries(input.nodes ?? {}).forEach(([name, spec]) => cb(name, spec));
		},
	},
	marks: {
		forEach: (cb: (name: string, spec: MarkSpec) => void) => {
			Object.entries(input.marks ?? {}).forEach(([name, spec]) => cb(name, spec));
		},
	},
	topNode: input.topNode,
}) as unknown as SchemaSpec;

describe("transformSchemaSpec", () => {
	test("converts schema nodes and marks into plain objects", () => {
		const doc = { content: "block+" };
		const paragraph = { content: "inline*" };
		const bold = {};

		expect(transformSchemaSpec(createSpec({
			nodes: { doc, paragraph },
			marks: { bold },
			topNode: "doc",
		}))).toEqual({
			nodes: { doc, paragraph },
			marks: { bold },
			topNode: "doc",
		});
	});

	test("defaults the top node to doc", () => {
		expect(transformSchemaSpec(createSpec({})).topNode).toBe("doc");
	});
});

describe("emptyDocument", () => {
	test("is encoded as a non-empty Yjs update", () => {
		expect(emptyDocument).toBeInstanceOf(Uint8Array);
		expect(emptyDocument.length).toBeGreaterThan(0);
	});
});
