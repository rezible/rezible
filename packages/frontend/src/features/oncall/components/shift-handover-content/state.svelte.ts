import { debounce } from "$lib/utils.svelte";
import type { OncallShiftHandover, OncallShiftHandoverSection } from "$lib/api";
import { SvelteMap } from "svelte/reactivity";
import { watch } from "runed";

import type { Content, Editor } from "@tiptap/core";
import { Editor as SvelteEditor } from "$components/tiptap-editor/TiptapEditor.svelte";
import { createHandoverEditor } from "$components/tiptap-editor/editors";

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

export class ShiftHandoverEditorState {
	handover = $state<OncallShiftHandover>();
	private allowEditing = $state(false);
	isSent = $derived(!!this.handover && (new Date(this.handover.attributes.sentAt).valueOf()) > 0);
	editable = $derived(this.allowEditing && !this.isSent);

	isEmpty = $state(true);
	activeEditor = $state<Editor>();
	sections = $state<HandoverSection[]>([]);

	canSend = $derived(!this.isSent && !this.isEmpty);

	constructor(handoverFn: () => (OncallShiftHandover | undefined), allowEditing: boolean) {
		this.allowEditing = allowEditing;

		watch(handoverFn, h => {
			if (this.handover?.id === h?.id) return;
			this.handover = h;
			this.setup();
		});
	}

	setup() {
		this.isEmpty = true;
		this.activeEditor = undefined;

		const handoverSections = this.handover?.attributes.content ?? [];
		this.sections = handoverSections.map((sec, idx) => {
			const { kind, header } = sec;
			if (kind !== "regular") return { header, kind }

			const content = !!sec.jsonContent ? JSON.parse(sec.jsonContent) : undefined;
			const { editor, activeStatus, contentEmpty } = this.createEditor(idx, content);
			if (this.isSent && contentEmpty) {
				editor.destroy();
				return { header, kind };
			}
			return { header, kind, editor, activeStatus };
		});
	};

	createEditor(idx: number, content?: Content) {
		const activeStatus = new SvelteMap<string, boolean>();
		const updateStatusFn = debounce((e: Editor) => {
			activeStatus.set("bold", e.isActive("bold"));
			activeStatus.set("bulletList", e.isActive("bulletList"));
			activeStatus.set("heading", e.isActive("heading"));
		}, 100);

		const editor = createHandoverEditor({
			content,
			editable: this.editable,
			autofocus: (this.editable && idx === 0) ? "end" : false,
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
