import { getTaskOptions, getIncidentOptions, getUserOptions } from "$lib/api";
import { createQuery } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";

export class TaskDetailController {
	private taskId = $state("");
	taskQuery = createQuery(() => ({
		...getTaskOptions({ path: { id: this.taskId } }),
		enabled: !!this.taskId,
	}));
	task = $derived(this.taskQuery.data?.data);

	private incidentId = $derived(this.task?.attributes.incidentId ?? "");
	incidentQuery = createQuery(() => ({
		...getIncidentOptions({ path: { id: this.incidentId } }),
		enabled: !!this.incidentId,
	}));
	incident = $derived(this.incidentQuery.data?.data);

	private ownerId = $derived(this.task?.attributes.ownerId?.id ?? "");
	ownerQuery = createQuery(() => ({
		...getUserOptions({ path: { id: this.ownerId } }),
		enabled: !!this.ownerId && !this.task?.attributes.ownerId?.attributes,
	}));

	ownerName = $derived.by(() => {
		const owner = this.task?.attributes.ownerId;
		if (!owner) return "Unassigned";
		return owner.attributes?.name ?? this.ownerQuery.data?.data.attributes.name ?? "Unavailable user";
	});
	tickets = $derived(
		(this.task?.attributes.externalTickets ?? []).map((ticket) => ({
			id: ticket.id,
			title: ticket.title || ticket.reference || "External ticket",
			href: ticket.url && /^https?:\/\//i.test(ticket.url) ? ticket.url : undefined,
		}))
	);

	constructor(idFn: Getter<string>) {
		this.taskId = idFn();
		watch(idFn, (id) => {
			this.taskId = id;
		});
	}
}

const ctx = new Context<TaskDetailController>("TaskDetailController");
export const initTaskDetailController = (idFn: Getter<string>) => ctx.set(new TaskDetailController(idFn));
