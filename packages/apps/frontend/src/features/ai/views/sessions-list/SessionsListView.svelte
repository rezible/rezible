<script lang="ts">
	import { Button } from "$components/ui/button";
	import { initSessionsListController } from "./controller.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { resolve } from "$app/paths";

	const controller = initSessionsListController();
	const formatTime = (value: string) => new Date(value).toLocaleString();

	const query = $derived(controller.paginatedSessionsQuery.query);
</script>

<div class="p-4">
	<header class="mb-4">
		<h1 class="text-xl font-semibold">AI Sessions</h1>
		<p class="text-sm text-muted-foreground">Persisted agent runs and linked analyses.</p>
	</header>

	<PaginatedQueryListBox {...controller.paginatedSessionsQuery}>
		{#if query.isPending}
			<p class="text-muted-foreground">Loading sessions…</p>
		{:else if query.error}
			<div class="text-destructive">
				Could not load sessions.
				<Button variant="outline" onclick={() => query.refetch()}>Retry</Button>
			</div>
		{:else if !controller.sessions.length}
			<p class="text-muted-foreground">No sessions found.</p>
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
							{@const attrs = session.attributes}
							<tr class="border-t border-border">
								<td class="p-2">
									<a
										class="font-medium underline"
										href={resolve(`/ai/sessions/${session.id}`)}
									>
										{attrs.agentName}
									</a>
									<div class="font-mono text-xs text-muted-foreground">
										{session.id}
									</div>
								</td>
								<td class="p-2 text-xs">
									<div>{formatTime(attrs.createdAt)}</div>
									<div class="text-muted-foreground">
										{formatTime(attrs.updatedAt)}
									</div>
								</td>
								<td class="p-2">
									{attrs.permissionScopes.length
										? attrs.permissionScopes.join(", ")
										: "None"}
								</td>
								<td class="p-2">{attrs.systemAnalysisId ? "Linked" : "—"}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</PaginatedQueryListBox>
</div>
