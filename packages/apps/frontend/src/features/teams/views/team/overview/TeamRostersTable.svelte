<script lang="ts">
	import Avatar from "$components/common/entity-avatar/EntityAvatar.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { resolve } from "$app/paths";
	import { useTeamOverviewController } from "./controller.svelte";

	const controller = useTeamOverviewController();
</script>

<div class="flex flex-col max-h-full max-w-xl border p-2">
	<span class="text-sm uppercase font-semibold text-surface-content/90">Rosters</span>

	<PaginatedQueryListBox dense {...controller.paginatedRostersQuery}>
		{#each controller.rosters as roster (roster.id)}
			<a
				class="flex gap-2 items-center rounded border border-surface-content/10 p-2"
				href={resolve(`/oncall/rosters/${roster.attributes.slug}`)}
			>
				<Avatar kind="roster" size={20} id={roster.id} />
				<span>{roster.attributes.name}</span>
			</a>
		{/each}
	</PaginatedQueryListBox>
</div>
