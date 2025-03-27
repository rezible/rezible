<script lang="ts">
	import { mdiChevronRight, mdiAccount } from "@mdi/js";
	import { Header, Icon } from "svelte-ux";
	import Avatar from "$components/avatar/Avatar.svelte";
	import type { User } from "../types";

	const makeFakeUsers = (): User[] => {
		// todo
		return [];
	}
	const users: User[] = makeFakeUsers();
</script>

<div class="flex flex-col gap-2 w-full">
	<div class="flex-1 flex flex-col gap-1 overflow-y-auto">
		{#each users as usr}
			<a href="/users/{usr.id}" class="block w-full">
				<div class="flex items-center bg-surface-100 hover:bg-accent-800/40 p-2 w-full">
					<div class="flex items-center gap-2">
						<Avatar kind="user" size={20} id={usr.id} />
						<span class="font-medium">{usr.attributes.name}</span>
					</div>
					<div class="flex-1 grid justify-items-end">
						<Icon data={mdiChevronRight} />
					</div>
				</div>
			</a>
		{:else}
			<div class="text-surface-600 italic p-2">No users assigned to this roster</div>
		{/each}
	</div>
</div>