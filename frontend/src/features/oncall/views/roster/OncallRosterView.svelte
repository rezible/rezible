<script lang="ts">
	import { type User, type OncallRoster } from "$lib/api";
	import Avatar from "$src/components/avatar/Avatar.svelte";
	import { mdiChevronRight } from "@mdi/js";
	import { Header, Icon } from "svelte-ux";

	type Props = { roster: OncallRoster };
	const { roster }: Props = $props();

	const attr = $derived(roster.attributes);

	const users = $derived<User[]>([]);
</script>


<div class="grid grid-cols-3 gap-2 h-full max-h-full min-h-0 overflow-hidden">
	<div class="flex flex-col gap-1 h-full min-h-0">
		{@render rosterDetails()}
	</div>

	<div class="col-span-2 flex flex-col gap-1 h-full min-h-0 border rounded-lg p-2">
		<Header title="Stats" />
		
	</div>
</div>

{#snippet rosterDetails()}
	<div class="grid grid-cols-2 grid-rows-5 gap-2 auto-rows-min h-full">
		<div class="row-span-2 flex flex-col gap-1 border rounded-lg p-2">
			<Header title="Users" />

			{#each users as usr}
				<a href="/users/{usr.id}" class="flex-1">
					<div class="flex items-center gap-4 bg-surface-100 hover:bg-accent-800/40 h-full p-2">
						<Avatar kind="user" size={32} id={usr.id} />
						<div class="flex flex-col">
							<span class="text-lg">{usr.attributes.name}</span>
						</div>
						<div class="flex-1 grid justify-items-end">
							<Icon data={mdiChevronRight} />
						</div>
					</div>
				</a>
			{/each}
		</div>

		<div class="row-span-2 flex flex-col gap-1 border rounded-lg p-2">
			<Header title="Services" />

			<!--a href="/oncall/rosters/{roster.id}" class="flex-1">
				<div class="flex items-center gap-4 bg-surface-100 hover:bg-accent-800/40 h-full p-2">
					<Avatar kind="roster" size={32} id={roster.id} />
					<span class="text-lg">{roster.attributes.name}</span>
					<div class="flex-1 grid justify-items-end">
						<Icon data={mdiChevronRight} />
					</div>
				</div>
			</a-->
		</div>

		<div class="border row-span-3 col-span-2 row-start-3 rounded-lg p-2 flex flex-col gap-1">
			<Header title="Recent Shifts" />

		</div>
	</div>
{/snippet}