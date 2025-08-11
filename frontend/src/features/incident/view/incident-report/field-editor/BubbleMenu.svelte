<script lang="ts" module>
	export type AnnotationType = "draft-comment" | "service";
</script>

<script lang="ts">
	import { Editor } from "$components/tiptap-editor/TiptapEditor.svelte";
	import Button from "$components/button/Button.svelte";
	import { mdiComment, mdiMarker } from "@mdi/js";
	import { PluginKey, type Selection } from "@tiptap/pm/state";
	import { BubbleMenuPlugin } from "@tiptap/extension-bubble-menu";
	import { onMount } from "svelte";

	interface Props {
		editor: Editor;
		field: string;
		onCreate: (t: AnnotationType) => void;
	}
	let { editor, field, onCreate }: Props = $props();

	const pluginKey = $derived(new PluginKey(`bubble-menu-${field}`));

	let hideBubbleMenuFn = $state<VoidFunction>();
	let bubbleMenuElement = $state<HTMLElement>();

	const hideBubbleMenu = () => {
		if (hideBubbleMenuFn) hideBubbleMenuFn();
	};

	const commentIsContainedInSelection = (selection: Selection) => {
		return false;
	};

	// type ShouldShowProps = {
	// 	editor: TipTapEditor;
	// 	view: EditorView;
	// 	state: EditorState;
	// 	oldState?: EditorState | undefined;
	// 	from: number;
	// 	to: number;
	// };
	// const shouldShowBubbleMenu = ({ editor, view, state, from, to }: ShouldShowProps) => {
	// 	if (!editor.isEditable) return false;

	// 	const { doc, selection } = state;
	// 	if (selection.empty) return false;

	// 	const isChildOfMenu = !!bubbleMenuElement && bubbleMenuElement.contains(document.activeElement);
	// 	const hasEditorFocus = view.hasFocus() || isChildOfMenu;
	// 	if (!hasEditorFocus) return false;

	// 	const isEmptyTextBlock = !doc.textBetween(from, to).length && isTextSelection(state.selection);
	// 	if (isEmptyTextBlock) return false;

	// 	return true;
	// };

	const registerBubbleMenu = (editor: Editor, element: HTMLElement) => {
		editor.registerPlugin(
			BubbleMenuPlugin({
				pluginKey,
				editor,
				element,
				tippyOptions: {
					duration: 100,
					placement: "right",
					hideOnClick: true,
					onShow: (inst) => {
						hideBubbleMenuFn = () => {
							if (inst) inst.hide();
						};
					},
					onHidden: () => {
						hideBubbleMenuFn = undefined;
					},
				},
				// shouldShow: shouldShowBubbleMenu,
			})
		);
		return () => {
			if (editor) editor.unregisterPlugin(pluginKey);
		};
	};

	onMount(() => {
		if (!editor || editor.isDestroyed || !bubbleMenuElement) {
			console.log("invalid editor for bubble menu");
			return;
		}
		return registerBubbleMenu(editor, bubbleMenuElement);
	});

	const onCommentButtonClicked = (event: Event) => {
		onCreate("draft-comment");
		hideBubbleMenu();
	};

	const onAnnotateButtonClicked = (event: Event) => {
		onCreate("service");
		hideBubbleMenu();
	};
</script>

<div>
	<div
		bind:this={bubbleMenuElement}
		style="visibility: hidden"
		class="flex flex-col gap-2 bg-surface-300 rounded-full border"
	>
		<Button
			iconOnly
			icon={mdiComment}
			classes={{ root: "bg-surface-200 hover:bg-primary" }}
			on:click={onCommentButtonClicked}
		/>

		<Button
			iconOnly
			icon={mdiMarker}
			classes={{ root: "bg-surface-200 hover:bg-primary" }}
			on:click={onAnnotateButtonClicked}
		/>
	</div>
</div>
