import {
	getTaskOptions,
	updateTaskMutation,
	getIncidentOptions,
	getUserOptions,
	getSystemAnalysisEntryOptions,
	listInboxItemsQueryKey,
	type Task,
} from "$lib/api";
import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch, type Getter } from "runed";

export class TaskDetailController {
	private queryClient = useQueryClient();

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

	private ownerId = $derived(this.task?.attributes.ownerId ?? "");
	ownerQuery = createQuery(() => ({
		...getUserOptions({ path: { id: this.ownerId } }),
		enabled: !!this.ownerId,
	}));

	private findingId = $derived(this.task?.attributes.originEntryId ?? "");
	findingQuery = createQuery(() => ({
		...getSystemAnalysisEntryOptions({ path: { id: this.findingId } }),
		enabled: !!this.findingId,
	}));

	draftState = $state<Task["attributes"]["state"]>("open");
	private initializedFor = "";

	update = createMutation(() => ({
		...updateTaskMutation(),
		onSuccess: (response) => {
			this.queryClient.setQueryData(
				getTaskOptions({ path: { id: response.data.id } }).queryKey,
				response
			);
			this.draftState = response.data.attributes.state;
			void this.queryClient.invalidateQueries({ queryKey: listInboxItemsQueryKey() });
		},
	}));

	constructor(idFn: Getter<string>) {
		watch(idFn, (id) => {
			this.taskId = id;
		});
		watch(
			() => this.task,
			(task) => {
				if (!task || this.initializedFor === task.id) return;
				this.initializedFor = task.id;
				this.draftState = task.attributes.state;
			}
		);
	}

	setState = (state: Task["attributes"]["state"]) => {
		this.draftState = state;
	};

	save = () => {
		if (!this.task || this.update.isPending) return;
		this.update.mutate({ path: { id: this.task.id }, body: { attributes: { state: this.draftState } } });
	};
}

const ctx = new Context<TaskDetailController>("TaskDetailController");
export const initTaskDetailController = (idFn: Getter<string>) => ctx.set(new TaskDetailController(idFn));
