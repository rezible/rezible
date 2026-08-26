<script lang="ts">
	import { Button } from "$components/ui/button";
	import { initSessionsListController } from "./controller.svelte";

	const controller = initSessionsListController();
	const formatTime = (value: string) => new Date(value).toLocaleString();
</script>

<div class="p-4">
	<header class="mb-4">
		<h1 class="text-xl font-semibold">AI Sessions</h1>
		<p class="text-sm text-muted-foreground">Persisted agent runs and linked analyses.</p>
	</header>
	{#if controller.query.isPending}
		<p class="text-muted-foreground">Loading sessions…</p>
	{:else if controller.query.error}
		<div class="text-destructive">
			Could not load sessions.
			<Button variant="outline" onclick={() => controller.query.refetch()}>Retry</Button>
		</div>
	{:else if !controller.sessions.length}
		<p class="text-muted-foreground">No AI sessions found.</p>
	{:else}
		<div class="overflow-x-auto border border-border">
			<table class="w-full text-sm">
				<thead class="bg-muted text-left">
					<tr>
						<th class="p-2">Agent / session</th>
						<th class="p-2">Created / updated</th>
						<th class="p-2">Permissions</th>
						<th class="p-2">Analysis</th>
					</tr>
				</thead>
				<tbody>
					{#each controller.sessions as session (session.id)}
						<tr class="border-t border-border">
							<td class="p-2">
								<a class="font-medium underline" href={`/ai/sessions/${session.id}`}>
									{session.attributes.agentName}
								</a>
								<div class="font-mono text-xs text-muted-foreground">
									{session.id}
								</div>
							</td>
							<td class="p-2 text-xs">
								<div>{formatTime(session.attributes.createdAt)}</div>
								<div class="text-muted-foreground">
									{formatTime(session.attributes.updatedAt)}
								</div>
							</td>
							<td class="p-2">
								{session.attributes.permissionScopes.length
									? session.attributes.permissionScopes.join(", ")
									: "None"}
							</td>
							<td class="p-2">{session.attributes.systemAnalysisId ? "Linked" : "—"}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
		<footer class="mt-3 flex items-center justify-between text-sm">
			<span>{controller.start}–{controller.end} of {controller.total}</span>
			<div class="flex gap-2">
				<Button
					variant="outline"
					disabled={controller.page === 0 || controller.query.isFetching}
					onclick={controller.previous}>Previous</Button
				>
				<Button
					variant="outline"
					disabled={controller.end >= controller.total || controller.query.isFetching}
					onclick={controller.next}>Next</Button
				>
			</div>
		</footer>
	{/if}
</div>
