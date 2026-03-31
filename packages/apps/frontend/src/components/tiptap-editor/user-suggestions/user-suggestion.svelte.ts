import type { Content } from "@tiptap/core";
import type { SuggestionKeyDownProps, SuggestionOptions, SuggestionProps } from "@tiptap/suggestion";
import { mount, unmount } from "svelte";
import SuggestionPopup from "./UserSuggestionPopup.svelte";

const userList = [
	"Lea Thompson",
	"Cyndi Lauper",
	"Tom Cruise",
	"Madonna",
	"Jerry Hall",
	"Joan Collins",
	"Winona Ryder",
	"Christina Applegate",
	"Alyssa Milano",
	"Molly Ringwald",
	"Ally Sheedy",
	"Debbie Harry",
	"Olivia Newton-John",
	"Elton John",
	"Michael J. Fox",
	"Axl Rose",
	"Emilio Estevez",
	"Ralph Macchio",
	"Rob Lowe",
	"Jennifer Grey",
	"Mickey Rourke",
	"John Cusack",
	"Matthew Broderick",
	"Justine Bateman",
	"Lisa Bonet",
];

const getClientRect = (fn: () => DOMRect | null) => {
	return () => {
		const fnRes = fn();
		return fnRes || new DOMRect(0, 0, 0, 0);
	};
};

export const RezUserSuggestion: Partial<SuggestionOptions<string, any>> = {
	allowSpaces: true,

	command: ({ editor, range, props }) => {
		const nodeAfter = editor.view.state.selection.$to.nodeAfter;
		const overrideSpace = nodeAfter?.text?.startsWith(" ");

		if (overrideSpace) {
			range.to += 1;
		}

		const content: Content[] = [
			{
				type: "rez-user-mention",
				attrs: props,
			},
			{
				type: "text",
				text: " ",
			},
		];

		editor.chain().focus().insertContentAt(range, content).run();
		window.getSelection()?.collapseToEnd();
	},

	items: async ({ query }) => {
		return userList.filter((item) => item.toLowerCase().startsWith(query.toLowerCase())).slice(0, 5);
	},

	render: () => {
		let componentProps = $state<SuggestionProps<string, any>>();
		let componentOnKeyDown: (props: SuggestionKeyDownProps) => boolean;
		let unmountComponent: () => void;
		// let popup: TippyInstance<TippyProps>[];

		return {
			onStart: (props) => {
				const target = document.createElement("div");
				if (props.decorationNode) {
					props.decorationNode.appendChild(target);
				}
				componentProps = props;
				const component = mount(SuggestionPopup, {
					target,
					props: componentProps,
				});
				componentOnKeyDown = component.onKeyDown;
				unmountComponent = () => unmount(component);

				if (!props.clientRect) return;

				/*
				popup = tippy("body", {
					getReferenceClientRect: getClientRect(props.clientRect),
					appendTo: () => document.body,
					content: target,
					showOnCreate: true,
					interactive: true,
					trigger: "manual",
					placement: "bottom-start",
				});
				*/
			},

			onUpdate(props) {
				if (componentProps) {
					componentProps.command = props.command;
					componentProps.items = props.items;
				}
				if (!props.clientRect) return;
				// popup[0].setProps({ getReferenceClientRect: getClientRect(props.clientRect) });
			},

			onKeyDown(props) {
				if (props.event.key === "Escape") {
					// popup[0].hide();
					return true;
				}

				return componentOnKeyDown(props);
			},

			onExit() {
				// popup[0].destroy();
				unmountComponent();
			},
		};
	},
};
