<script lang="ts">
	import { useHomeController } from "./controller.svelte";
	import { inboxActions } from "./model";
	import DisplayTime from "./DisplayTime.svelte";
	import { Button } from "$components/ui/button";
	import { Skeleton } from "$components/ui/skeleton";
	import LoadingQueryWrapper from "$components/layout/loading-query-wrapper/LoadingQueryWrapper.svelte";
	import RiQuestionLine from "remixicon-svelte/icons/question-line";
	import RiChatQuoteLine from "remixicon-svelte/icons/chat-quote-line";
	import RiTaskLine from "remixicon-svelte/icons/task-line";
	import RiFileTextLine from "remixicon-svelte/icons/file-text-line";

	const controller = useHomeController();

	const query = $derived(controller.inboxQuery);

	const icons = {
		question: RiQuestionLine,
		annotation: RiChatQuoteLine,
		task: RiTaskLine,
		maintenance: RiFileTextLine,
	};
</script>

<section aria-labelledby="home-inbox" class="min-w-0">
	<header class="mb-3 flex items-center justify-between gap-3">
		<h2 id="home-inbox" class="text-lg font-semibold">
			Inbox <span class="ml-2 text-sm font-normal text-muted-foreground">
				{query.data?.pagination.total ?? "—"}
			</span>
		</h2>
		<a class="text-sm text-primary hover:underline focus-visible:outline-ring" href="/">View all</a>
	</header>
	<div class="overflow-hidden rounded-md border border-border bg-card">
		<LoadingQueryWrapper {query} feedbackOnly>
			{#snippet loading()}
				<div class="divide-y" aria-label="Loading Inbox">
					{#each [1, 2] as row (row)}
						<div class="flex gap-3 p-4">
							<Skeleton class="size-5" />
							<div class="flex-1 space-y-2">
								<Skeleton class="h-4 w-2/3" />
								<Skeleton class="h-3 w-5/6" />
							</div>
							<Skeleton class="h-8 w-24" />
						</div>
					{/each}
				</div>
			{/snippet}
		</LoadingQueryWrapper>

		{#if query.data}
			<ul class="divide-y divide-border">
				{#each query.data.data as item (item.id)}
					{@const Icon = icons[item.attributes.kind]}
					<li class="flex flex-wrap items-start gap-3 px-4 py-3">
						<Icon class="mt-1 size-5 shrink-0" aria-hidden="true" />
						<div class="min-w-40 flex-1 wrap-anywhere">
							<a
								class="text-sm font-medium hover:underline focus-visible:outline-ring"
								href={"/"}
							>
								{item.attributes.title}
							</a>
							<p class="mt-1 text-sm text-muted-foreground">{item.attributes.reason}</p>
						</div>
						<div class="ml-auto flex items-center gap-3 self-center">
							<Button variant="outline" size="sm" class="min-w-24" href={"/"}
							>
								{inboxActions[item.attributes.kind]}
							</Button>
							<DisplayTime value={item.attributes.occurredAt} />
						</div>
					</li>
				{:else}
					<li class="p-6 text-center text-sm text-muted-foreground">Inbox is clear.</li>
				{/each}
			</ul>
		{/if}
	</div>
</section>
