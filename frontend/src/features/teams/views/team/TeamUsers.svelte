<script lang="ts">
	import type { User } from "$lib/api";
	import Avatar from "$components/avatar/Avatar.svelte";

	interface Props {
		users: User[];
	}
	let { users }: Props = $props();
</script>

<div class="flex flex-col gap-2 w-fit">
	{#each users as user (user.id)}
		<a href="/users/{user.id}">
			<div
				class="flex items-center gap-2 px-2 hover:bg-surface-200 p-1 rounded-lg"
			>
				<Avatar kind="user" id={user.attributes.name} size={32} />
				<span>{user.attributes.name}</span>
			</div>
		</a>
	{/each}
</div>
