<script lang="ts">
	import type { Team, User } from "$lib/api";

	type Props = {
		user: User;
		teams: Team[];
	};
	const { user, teams }: Props = $props();
</script>

<div class="">
	<h2>Team Membership</h2>

	{#if teams.length === 0}
		<div class="">
			<p>Not a member of any teams</p>
		</div>
	{:else}
		<ul class="">
			{#each teams as team}
				{@const attr = team.attributes}
				<li>
					<a href="/teams/{team.id}" class="">
						<div class="" style="background-color: #4299E1">
							{attr.name.substring(0, 2).toUpperCase()}
						</div>
						<div class="">
							<div class="">{attr.name}</div>
						</div>
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</div>
