import { getSchema } from "@tiptap/core";
import { TiptapTransformer } from "@hocuspocus/transformer";

import {
  configureBaseExtensions,
  configureUserMentionExtension,
  configureAnnotationExtension,
  getHandoverExtensions,
} from "@rezible/tiptap-extensions";
import type { MarkSpec, NodeSpec, SchemaSpec } from "@tiptap/pm/model";

export const extensions = [
	...configureBaseExtensions(false),
	configureAnnotationExtension(),
	configureUserMentionExtension(),
];

export const handoverSchema = getSchema(getHandoverExtensions());
export const schema = getSchema(extensions);

export const documentTransformer = TiptapTransformer.extensions(extensions);

type transformedSpec = {
	marks: Object;
	nodes: Object;
	topNode: string;
}
export const transformSchemaSpec = (spec: SchemaSpec): transformedSpec => {
	let marks = new Map<string, MarkSpec>();
	let nodes = new Map<string, NodeSpec>();
	
	if (typeof spec.marks?.forEach == "function") {
		spec.marks?.forEach((name, val) => marks.set(name, val));
	}
	if (typeof spec.nodes?.forEach == "function") {
		spec.nodes?.forEach((name, val) => nodes.set(name, val));
	}
	const topNode = spec.topNode ?? "doc";
	const tsSpec: transformedSpec = {
		marks: Object.fromEntries(marks), 
		nodes: Object.fromEntries(nodes), 
		topNode,
	};
	return tsSpec;
}