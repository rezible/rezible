<script lang="ts">
	import { appShell } from "$features/app";

	import Avatar from "$components/avatar/Avatar.svelte";
	import OncallStats from "./OncallStats.svelte";
	import Header from "$components/header/Header.svelte";
	import { initUserViewController } from "./controller.svelte";

	const { id }: { id: string } = $props();

	const view = initUserViewController(() => id);

	appShell.setPageBreadcrumbs(() => [
		{ label: "Users", href: "/users" },
		{ label: view.userName, href: `/users/${view.userId}`, avatar: {kind: "user", id: view.userId}},
	]);
</script>

<div class="grid grid-cols-3 gap-2 h-full">
	<div class="flex flex-col gap-2">
		<div class="border p-2">
			<Header title="Information" classes={{title: "text-xl"}} />
		
			<div class="">
				<div class="">
					<span class="">🌐</span>
					<span>{view.timeZone}</span>
					<span class="text-surface-content">({view.userLocalTime})</span>
				</div>
				<div class="">
					<span class="">✉️</span>
					<a href="mailto:{view.user?.attributes.email}">{view.user?.attributes.email}</a>
				</div>
			</div>
		</div>

		<div class="flex-1 grid grid-cols-2 gap-2">
			<div class="flex flex-col p-2 border">
				<Header title="Teams" classes={{title: "text-xl"}} />
			
				<div class="flex flex-col gap-2">
					{#if !view.teams}
						<span>loading</span>
					{:else}
						{#each view.teams as team}
							{@const attr = team.attributes}
							<a href="/teams/{team.id}" class="rounded-lg border bg-neutral p-2">
								<div class="flex items-center gap-2">
									<Avatar kind="team" id={team.id} />
									<span class="text-lg">{attr.name}</span>
								</div>
							</a>
						{:else}
							<p>Not a member of any teams</p>
						{/each}
					{/if}
				</div>
			</div>

			<div class="flex flex-col p-2 border">
				<Header title="Rosters" classes={{title: "text-xl"}} />
			
				<div class="flex flex-col gap-2">
					{#if !view.rosters}
						<span>loading</span>
					{:else}
						{#each view.rosters as roster}
							{@const attr = roster.attributes}
							<a href="/rosters/{roster.id}" class="rounded-lg border bg-neutral p-2">
								<div class="flex items-center gap-2">
									<Avatar kind="roster" id={roster.id} />
									<span class="text-lg">{attr.name}</span>
								</div>
							</a>
						{:else}
							<p>Not a member of any oncall roster</p>
						{/each}
					{/if}
				</div>
			</div>
		</div>
	</div>

	<div class="col-span-2 border">
		<OncallStats />
	</div>
</div>
