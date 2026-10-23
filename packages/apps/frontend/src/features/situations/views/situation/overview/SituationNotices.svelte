<script lang="ts">
	import RiMessage2Line from "remixicon-svelte/icons/message-2-line";

	import { Button } from "$components/ui/button";
	import type { NoticeLine } from "$features/situations/lib/notices";

	import type { SituationNoticesController } from "./situationNotices.svelte";

	type Props = { notices: SituationNoticesController };
	const { notices }: Props = $props();
</script>

{#snippet channel(line: NoticeLine)}
	{#if line.href}
		<a
			class="underline underline-offset-4 hover:text-foreground"
			href={line.href}
			target="_blank"
			rel="noreferrer"
		>
			{line.channel}
		</a>
	{:else}
		<span>{line.channel}</span>
	{/if}
{/snippet}

{#if notices.lines.length > 0}
	<ul aria-label="Chat notices" class="flex flex-col gap-1 text-xs text-muted-foreground">
		<!-- Two installations can post to the same channel, so lines are keyed by position. -->
		{#each notices.lines as line, index (index)}
			<li class="flex flex-wrap items-center gap-x-1.5 gap-y-1">
				<RiMessage2Line class="size-3.5 shrink-0" aria-hidden="true" />
				{#if line.kind === "posted"}
					<span>Posted in</span>
					{@render channel(line)}
					{#if line.text}
						<span aria-hidden="true">·</span>
						<span>{line.text}</span>
					{/if}
				{:else}
					<span>{line.text}</span>
				{/if}
				{#if line.retryable}
					<span aria-hidden="true">·</span>
					<Button
						variant="link"
						size="inline"
						class="text-xs"
						disabled={notices.retrying}
						onclick={notices.retry}
					>
						Retry
					</Button>
				{/if}
			</li>
		{/each}
		{#if notices.ratings}
			<li class="pl-5">{notices.ratings}</li>
		{/if}
	</ul>
{/if}
