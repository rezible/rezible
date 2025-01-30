import type { Editor, Content } from "@tiptap/core";
import { Editor as SvelteEditor } from "svelte-tiptap";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import Bold from "@tiptap/extension-bold";
import Italic from "@tiptap/extension-italic";
import { session } from "$lib/auth.svelte";

import { RezUserSuggestion } from "./user-suggestions/user-suggestion.svelte";
import { configureUserMentionExtension } from "@rezible/documents/tiptap-extensions";

export const createReplyEditor = (content: Content, editable?: boolean) => {
  const userMentions = configureUserMentionExtension(RezUserSuggestion);
  return new SvelteEditor({
    content,
    editable,
    autofocus: editable,
    extensions: [Document, Paragraph, Text, Bold, Italic, userMentions],
    editorProps: {
      attributes: {
        class: "focus:outline-none",
      },
    },
  });
};

const createActiveDiscussion = () => {
  let value = $state<string>();

  return {
    get id() {
      return value;
    },
    set: (id?: string) => {
      value = id;
    },
  };
};
export const activeDiscussion = createActiveDiscussion();

type Draft = {
  editor: Editor;
};
const createDraft = () => {
  let value = $state<Draft>();

  const set = (val?: Draft) => {
    value = val;
  };

  const clear = (navigate: boolean) => {
    if (value) {
      if (navigate) value.editor.commands.navigateToDraftDiscussion();
      value.editor.commands.clearDraftDiscussion();
    }
    set();
  };

  const create = (editor: Editor) => {
    if (!session.user) return;
    clear(false);
    set({ editor });
    editor.commands.draftDiscussion();
  };

  return {
    get open() {
      return value !== undefined;
    },
    get editor() {
      return value?.editor;
    },
    set,
    create,
    clear,
  };
};
export const draft = createDraft();
