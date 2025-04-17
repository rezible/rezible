import type { Content, Editor } from "@tiptap/core";
import { Editor as SvelteEditor } from "$components/tiptap-editor/TiptapEditor.svelte";
import { debounce } from "$lib/utils.svelte";
import { getHandoverExtensions } from "@rezible/documents/tiptap-extensions";
import type {
	OncallShiftHandover,
	OncallShiftHandoverSection,
} from "$lib/api";
import { SvelteMap } from "svelte/reactivity";
import { onMount } from "svelte";
import { watch } from "runed";

export type HandoverEditorSection = {
	header: string;
	editor?: SvelteEditor;
	activeStatus?: Map<string, boolean>;
	kind: "regular";
};

type HandoverAnnotationsSection = {
	header: string;
	kind: "annotations";
};

export type HandoverSection = HandoverEditorSection | HandoverAnnotationsSection;

export class HandoverEditorState {
	sent = $state(false);
	isEmpty = $state(true);
	activeEditor = $state<Editor>();
	sections = $state<HandoverSection[]>([]);

	constructor(handoverFn: () => OncallShiftHandover) {
		watch(handoverFn, h => this.setupHandover(h))
		onMount(() => {
			return () => this.destroyEditors();
		})
	}

	reset() {
		this.sent = false;
		this.isEmpty = true;
		this.destroyEditors();
		this.activeEditor = undefined;
		this.sections = [];
	};

	setupHandover(handover: OncallShiftHandover) {
		this.reset();
		this.sent = (new Date(handover.attributes.sentAt).valueOf()) > 0;

		this.sections = handover.attributes.content.map((sec, idx) => {
			const { kind, header } = sec;
			if (kind !== "regular") return { header, kind }

			const content = !!sec.jsonContent ? JSON.parse(sec.jsonContent) : undefined;
			const { editor, activeStatus, contentEmpty } = this.createEditor(idx, content);
			if (this.sent && contentEmpty) {
				editor.destroy();
				return { header, kind };
			}
			return { header, kind, editor, activeStatus };
		});
	};

	destroyEditors() {
		this.sections.forEach((s) => {
			if (s.kind === "regular") s.editor?.destroy();
		});
	};

	createEditor(idx: number, content?: Content) {
		const activeStatus = new SvelteMap<string, boolean>();
		const updateStatusFn = debounce((e: Editor) => {
			activeStatus.set("bold", e.isActive("bold"));
			activeStatus.set("bulletList", e.isActive("bulletList"));
			activeStatus.set("heading", e.isActive("heading"));
		}, 100);

		const editor = new SvelteEditor({
			content,
			editable: !this.sent,
			extensions: getHandoverExtensions(),
			autofocus: (!this.sent && idx === 0) ? "end" : false,
			editorProps: {
				attributes: {
					class: "max-w-none focus:outline-none list-disc",
				},
			},
			onTransaction: ({ editor }) => {
				updateStatusFn(editor);
			},
			onUpdate: ({ editor }) => {
				this.isEmpty = editor.isEmpty;
			},
			onFocus: ({ editor }) => {
				if (this.activeEditor != editor) this.activeEditor = editor;
			},
		});

		const contentEmpty = editor.getText().trim() == "";

		return { editor, activeStatus, contentEmpty };
	};

	setSent() {
		this.sent = true;
		this.sections.forEach((s) => {
			if (s.kind === "regular") s.editor?.setEditable(false);
		});
		this.activeEditor = undefined;
	};

	setEditorFocus(i: number, focus: boolean) {
		if (i >= this.sections.length || this.sections[i].kind != "regular") return;
		const editor = this.sections[i].editor;
		if (!editor) return;
		if (focus && !editor.isFocused) editor.commands.focus();
	};

	getSectionContent(): OncallShiftHandoverSection[] {
		return this.sections.map((s) => {
			const jsonContent = s.kind === "regular" ? JSON.stringify(s.editor?.getJSON()) : undefined;
			return { header: s.header, kind: s.kind, jsonContent };
		});
	};
}
