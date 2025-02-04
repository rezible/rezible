import type { Content, Editor, HTMLContent, JSONContent } from "@tiptap/core";
import { Editor as SvelteEditor } from "svelte-tiptap";
import { debounce } from "$lib/utils.svelte";
import {
	configureBaseExtensions,
	getHandoverExtensions,
} from "@rezible/documents/tiptap-extensions";
import type {
	OncallShiftHandover,
	OncallShiftHandoverSection,
	OncallShiftHandoverTemplate,
	OncallShiftHandoverTemplateSection,
} from "$lib/api";
import { SvelteMap } from "svelte/reactivity";

type HandoverEditorSection = {
	header: string;
	editor: SvelteEditor;
	activeStatus: Map<string, boolean>;
	kind: "regular";
};

type HandoverAnnotationsSection = {
	header: string;
	kind: "annotations";
};

type HandoverIncidentsSection = {
	header: string;
	kind: "incidents";
};

export type HandoverSection =
	| HandoverEditorSection
	| HandoverAnnotationsSection
	| HandoverIncidentsSection;

const createHandoverState = () => {
	let sent = $state(false);
	let isEmpty = $state(true);
	let activeEditor = $state<Editor>();
	let sections = $state<HandoverSection[]>([]);

	const resetState = () => {
		activeEditor = undefined;
		sections = [];
	};

	const setupTemplate = async (template: OncallShiftHandoverTemplate) => {
		resetState();
		template.attributes.sections.forEach((sec, idx) => {
			if (sec.type === "regular") {
				let content = sec.list ? "<ul><li></li></ul>" : "";
				const { editor, activeStatus } = createEditor(idx, content);
				sections.push({
					header: sec.header,
					kind: sec.type,
					editor: editor,
					activeStatus: activeStatus,
				});
			} else {
				sections.push({
					header: sec.header,
					kind: sec.type,
				});
			}
		});
	};

	const restoreExisting = (handover: OncallShiftHandover) => {
		resetState();
		handover.attributes.content.forEach((sec, idx) => {
			if (sec.kind === "regular") {
				const { editor, activeStatus } = createEditor(
					idx,
					!!sec.jsonContent ? JSON.parse(sec.jsonContent) : undefined
				);
				sections.push({
					header: sec.header,
					kind: sec.kind,
					editor: editor,
					activeStatus: activeStatus,
				});
			} else if (sec.kind == "annotations" || sec.kind == "incidents") {
				sections.push({
					header: sec.header,
					kind: sec.kind,
				});
			}
		});
	};

	const setSent = () => {
		sent = true;
		sections.forEach((s) => {
			if (s.kind === "regular") s.editor.setEditable(false);
			activeEditor = undefined;
		});
	};

	const createEditor = (idx: number, content?: Content) => {
		const activeStatus = $state(new SvelteMap<string, boolean>());
		const updateStatusFn = debounce((e: Editor) => {
			activeStatus.set("bold", e.isActive("bold"));
			activeStatus.set("bulletList", e.isActive("bulletList"));
			activeStatus.set("heading", e.isActive("heading"));
		}, 100);

		const extensions = getHandoverExtensions();

		const editor = new SvelteEditor({
			content,
			editable: !sent,
			extensions,
			autofocus: idx === 0 ? "end" : false,
			editorProps: {
				attributes: {
					class: "m-3 max-w-none focus:outline-none prose prose-sm",
				},
			},
			onTransaction: ({ editor }) => {
				updateStatusFn(editor);
			},
			onUpdate({ editor }) {
				isEmpty = editor.isEmpty;
			},
			onFocus({ editor }) {
				if (activeEditor != editor) activeEditor = editor;
			},
			onBlur({ editor }) {
				// activeStatus.clear();
				// if (activeEditor === editor) activeEditor = undefined;
			},
		});

		return { editor, activeStatus };
	};

	const setEditorFocus = (i: number, focus: boolean) => {
		if (i >= sections.length || sections[i].kind != "regular") return;
		const editor = sections[i].editor;
		if (!editor) return;
		if (focus && !editor.isFocused) editor.commands.focus();
	};

	const getSectionContent = (): OncallShiftHandoverSection[] => {
		return sections.map((s) => {
			const jsonContent =
				s.kind === "regular"
					? JSON.stringify(s.editor.getJSON())
					: undefined;
			return { header: s.header, kind: s.kind, jsonContent };
		});
	};

	const destroy = () => {
		sections.forEach((s) => {
			if (s.kind === "regular") s.editor?.destroy();
		});
	};

	return {
		setupTemplate,
		restoreExisting,
		destroy,
		setEditorFocus,
		setSent,
		getSectionContent,
		get sections() {
			return sections;
		},
		get activeEditor() {
			return activeEditor;
		},
		get sent() {
			return sent;
		},
		get isEmpty() {
			return isEmpty;
		},
	};
};
export const handoverState = createHandoverState();
