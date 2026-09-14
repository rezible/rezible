<script lang="ts">
	import type { Task } from "$lib/api";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { Button } from "$components/ui/button";
	import * as Select from "$components/ui/select";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import ErrorAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { initTaskDetailController } from "./controller.svelte";

	type Props = { taskId: string };
	let { taskId }: Props = $props();

	const controller = initTaskDetailController(() => taskId);

	const states: Task["attributes"]["state"][] = ["open", "completed", "cancelled"];
	registerPageDescriptor(() => ({
		title: controller.task?.attributes.name ?? "Task",
	}));

	const task = $derived(controller.task);
	const incident = $derived(controller.incident);
	const incidentHref = $derived(`/incidents/${incident?.attributes.slug}`);
</script>

<div class="min-h-0 flex-1 overflow-y-auto p-6">
	<div class="mx-auto max-w-3xl space-y-6">
		<LoadingQueryWrapper query={controller.taskQuery} feedbackOnly />

		{#if task}
			<section class="space-y-6 rounded-lg border border-border bg-card p-6">
				<header>
					<p class="text-sm text-muted-foreground">Incident follow-up</p>
					<h1 class="mt-1 text-2xl font-semibold">{task.attributes.name}</h1>
					<p class="mt-2 text-sm text-muted-foreground">{task.attributes.description}</p>
				</header>
				<dl class="grid gap-4 border-t pt-5 text-sm sm:grid-cols-2">
					{#if incident}
						<div>
							<dt class="text-xs text-muted-foreground">Source incident</dt>
							<dd>
								<a class="text-primary underline" href={incidentHref}>
									{incident.attributes.title}</a
								>
							</dd>
						</div>
					{/if}
					<div>
						<dt class="text-xs text-muted-foreground">Owner</dt>
						<dd>
							{#if task.attributes.ownerId}
								<LoadingQueryWrapper query={controller.ownerQuery} feedbackOnly />{controller
									.ownerQuery.data?.data.attributes.name ?? "Loading owner…"}
							{:else}
								Unassigned
							{/if}
						</dd>
					</div>
					<div>
						<dt class="text-xs text-muted-foreground">Due date</dt>
						<dd>
							{#if task.attributes.dueAt}
								<time datetime={task.attributes.dueAt}
									>{new Date(task.attributes.dueAt).toLocaleString()}</time
								>
							{:else}
								No due date
							{/if}
						</dd>
					</div>
					<div>
						<dt class="text-xs text-muted-foreground">Current status</dt>
						<dd class="capitalize">{task.attributes.state}</dd>
					</div>
				</dl>

				{#if task.attributes.originEntryId}
					<section class="space-y-2 border-t pt-4" aria-label="Originating finding">
						<h2 class="text-sm font-medium">Originating finding</h2>
						<LoadingQueryWrapper query={controller.findingQuery} feedbackOnly />
						{#if controller.findingQuery.data}
							<h3 class="text-sm">{controller.findingQuery.data.data.attributes.title}</h3>
							<p class="whitespace-pre-wrap text-sm text-muted-foreground">
								{controller.findingQuery.data.data.attributes.body}
							</p>
						{/if}
					</section>
				{/if}
				<form
					class="space-y-3 border-t pt-5"
					onsubmit={(event) => {
						event.preventDefault();
						controller.save();
					}}
				>
					<label for="task-status" class="text-sm font-medium">Status</label>
					<Select.Root
						type="single"
						value={controller.draftState}
						onValueChange={(value) => controller.setState(value as Task["attributes"]["state"])}
						disabled={controller.update.isPending}
						><Select.Trigger id="task-status" class="w-full capitalize"
							>{controller.draftState}</Select.Trigger
						><Select.Content>
							{#each states as state (state)}
								<Select.Item value={state} class="capitalize">{state}</Select.Item>
							{/each}
						</Select.Content></Select.Root
					>
					<Button type="submit" disabled={controller.update.isPending}
						>{controller.update.isPending ? "Saving…" : "Save status"}</Button
					>

					{#if controller.update.error}
						<div role="alert">
							<ErrorAlert error={controller.update.error} />
							<p class="mt-2 text-sm">Your selected status is retained.</p>
						</div>
					{/if}

					{#if controller.update.isSuccess}
						<p role="status" class="text-sm">
							Status saved: {task.attributes.state}.
						</p>
					{/if}
				</form>
			</section>
		{/if}
	</div>
</div>
