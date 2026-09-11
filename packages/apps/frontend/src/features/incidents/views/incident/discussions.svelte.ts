import type { Editor } from "@tiptap/core";
import type { JSONContent } from "@tiptap/core";
import { createMutation, useQueryClient } from "@tanstack/svelte-query";
import {
	createDiscussionThreadMutation,
	listDiscussionCommentsOptions,
	listDiscussionThreadsOptions,
} from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";

export const discussionCommentContent = (content: string): JSONContent => {
	return JSON.parse(content) as JSONContent;
};

export const createDiscussionCommentsController = (discussionId: () => string) =>
	createPaginatedQuery({
		source: "local",
		queryOptions: () =>
			listDiscussionCommentsOptions({ path: { id: discussionId() }, query: { page: 1, pageSize: 1 } }),
	});

export const createDiscussionsController = (retrospectiveId: () => string) => {
	const queryClient = useQueryClient();
	const paginatedQuery = createPaginatedQuery({
		source: "local",
		queryOptions: (pagination) =>
			listDiscussionThreadsOptions({ query: { retrospectiveId: retrospectiveId(), ...pagination } }),
	});
	const queryOptions = () =>
		listDiscussionThreadsOptions({ query: { retrospectiveId: retrospectiveId() } });
	const createDiscussion = createMutation(() => ({
		...createDiscussionThreadMutation(),
		onSuccess: () => {
			draft.clear(true);
			const { queryKey } = queryOptions();
			queryClient.invalidateQueries({ queryKey });
		},
	}));

	return {
		get query() {
			return paginatedQuery.query;
		},
		createDiscussion,
		saveDraft: (editor: Editor) =>
			createDiscussion.mutate({
				body: {
					attributes: {
						retrospectiveId: retrospectiveId(),
						kind: "comment",
						initialMessage: JSON.stringify(editor.getJSON()),
					},
				},
			}),
	};
};
export type DiscussionsController = ReturnType<typeof createDiscussionsController>;

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
