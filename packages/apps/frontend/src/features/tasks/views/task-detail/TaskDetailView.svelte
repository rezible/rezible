<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import { initTaskDetailController } from "./controller.svelte";

	type Props = { taskId: string };
	let { taskId }: Props = $props();

	const controller = initTaskDetailController(() => taskId);

	registerPageDescriptor(() => ({
		title: controller.task?.attributes.title ?? "Task",
	}));

	const task = $derived(controller.task);
	const incident = $derived(controller.incident);
	const incidentHref = $derived(`/incidents/${incident?.attributes.slug}`);
</script>

<div class="min-h-0 flex-1 overflow-y-auto p-6">
	<div class="mx-auto max-w-3xl flex flex-col gap-6">
		<LoadingQueryWrapper query={controller.taskQuery} feedbackOnly />

		{#if task}
			<section class="flex flex-col gap-6 rounded-lg border border-border bg-card p-6">
				<header>
					<p class="text-sm text-muted-foreground">Incident follow-up</p>
					<h1 class="mt-1 text-2xl font-semibold">{task.attributes.title}</h1>
					<p class="mt-2 text-sm text-muted-foreground">{task.attributes.description}</p>
				</header>
				<dl class="grid gap-4 border-t pt-5 text-sm sm:grid-cols-2">
					{#if incident}
						<div>
							<dt class="text-xs text-muted-foreground">Source incident</dt>
							<dd>
								<a class="text-primary underline" href={incidentHref}>
									{incident.attributes.title}
								</a>
							</dd>
						</div>
					{/if}
					<div>
						<dt class="text-xs text-muted-foreground">Owner</dt>
						<dd>
							{controller.ownerName}
						</dd>
					</div>
					<div>
						<dt class="text-xs text-muted-foreground">Due date</dt>
						<dd>
							{#if task.attributes.dueAt}
								<time datetime={task.attributes.dueAt}>
									{new Date(task.attributes.dueAt).toLocaleString()}
								</time>
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

				{#if task.attributes.author}
					<p class="text-sm text-muted-foreground">
						Created by {task.attributes.author.attributes?.name ?? "Unknown user"}
					</p>
				{/if}
				{#if controller.tickets.length}
					<section class="flex flex-col gap-2 border-t pt-4" aria-label="External tickets">
						<h2 class="text-sm font-medium">External tickets</h2>
						{#each controller.tickets as ticket (ticket.id)}
							{#if ticket.href}
								<a
									class="text-sm underline"
									href={ticket.href}
									target="_blank"
									rel="noreferrer"
								>
									{ticket.title}
								</a>
							{:else}
								<p class="text-sm">{ticket.title}</p>
							{/if}
						{/each}
					</section>
				{/if}
			</section>
		{/if}
	</div>
</div>
