<script lang="ts">
	import Avatar from "$components/common/entity-avatar/EntityAvatar.svelte";
	import PaginatedQueryListBox from "$components/layout/paginated-query-listbox/PaginatedQueryListBox.svelte";
	import { resolve } from "$app/paths";
	import { useTeamOverviewController } from "./controller.svelte";

	const controller = useTeamOverviewController();
</script>

<div class="flex flex-col max-h-full max-w-xl border p-2">
	<span class="uppercase font-semibold text-surface-content/90">Users</span>

	<PaginatedQueryListBox dense {...controller.paginatedUsersQuery}>
		{#each controller.users as user (user.id)}
			<a
				class="flex gap-2 items-center rounded border border-surface-content/10 p-2"
				href={resolve(`/users/${user.id}`)}
			>
				<Avatar kind="user" size={20} id={user.id} />
				<span>{user.attributes.name}</span>
			</a>
		{/each}
	</PaginatedQueryListBox>
</div>
