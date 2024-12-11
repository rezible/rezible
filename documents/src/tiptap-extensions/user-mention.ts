import Mention from "@tiptap/extension-mention";
import { mergeAttributes } from "@tiptap/core";

export type SuggestionExtensionType = typeof Mention.options.suggestion;

export const RezUserMentionExtension = Mention.extend({
    name: "rez-user-mention",
}).configure({
    deleteTriggerWithBackspace: true,
    HTMLAttributes: {
        class: 'rez-user-mention',
        target: "_blank",
    },
    renderHTML({ options, node }) {
        return [
          'a',
          mergeAttributes({ href: '/users/' + node.attrs.id }, options.HTMLAttributes),
          `${options.suggestion.char}${node.attrs.label ?? node.attrs.id}`,
        ]
    },
})