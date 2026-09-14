import { Context } from "runed";
import { createQuery } from "@tanstack/svelte-query";
import {
	listActivityOptions,
	listInboxItemsOptions,
	listIncidentsOptions,
	listSituationsOptions,
	type ErrorModel,
	type DiscussionThread,
	type DiscussionComment,
} from "$lib/api";
import { timestamp } from "./model";

type SubmissionMapValue = {
	pending: boolean;
	error?: ErrorModel;
	result?: DiscussionComment | DiscussionThread;
};

class HomeController {
	drafts = $state<Record<string, string>>({});
	submissions = $state<Record<string, SubmissionMapValue>>({});

	incidentsQuery = createQuery(() =>
		listIncidentsOptions({
			query: { page: 1, pageSize: 3, statuses: ["started", "mitigated"] },
		})
	);

	situationsQuery = createQuery(() =>
		listSituationsOptions({
			query: { page: 1, pageSize: 3, status: "active" },
		})
	);

	inboxQuery = createQuery(() =>
		listInboxItemsOptions({
			query: { page: 1, pageSize: 5, state: "open" },
		})
	);

	activityQuery = createQuery(() =>
		listActivityOptions({
			query: { page: 1, pageSize: 9 },
		})
	);
	activityGroups = $derived(
		["Today", "Earlier"]
			.map((label) => ({
				label,
				items:
					this.activityQuery.data?.data.filter(
						(item) => timestamp(item.attributes.occurredAt).day === label
					) ?? [],
			}))
			.filter((group) => group.items.length)
	);
}

const ctx = new Context<HomeController>("HomeController");
export const initHomeController = () => ctx.set(new HomeController());
export const useHomeController = () => ctx.get();
