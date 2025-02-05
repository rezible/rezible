<script lang="ts">
	import { AvatarMarble } from "svelte-boring-avatars";

	const avatarColors = ["#92A1C6", "#146A7C", "#F0AB3D", "#C271B4", "#C20D90"];

	type User = {
		name: string;
		role: string;
	};
	const users: User[] = [
		{ name: "Hart Hagerty", role: "admin" },
		{ name: "Foo Bar", role: "user" },
		{ name: "Foo Bar 2", role: "user" },
		{ name: "Foo Bar 3", role: "user" },
		{ name: "Foo Bar 4", role: "user" },
		{ name: "Foo Bar 5", role: "user" },
		{ name: "Foo Bar 6", role: "user" },
		{ name: "Foo Bar 7", role: "user" },
		{ name: "Foo Bar 8", role: "user" },
	];

	let curPage = $state(0);

	const sizeOptions = [10, 25, 50, 100];
	let pageSize = $state(10);

	const pages = $derived(Math.ceil(users.length / pageSize));
	const usersView = $derived(users.slice(pageSize * curPage, pageSize * (curPage + 1)));

	const editUser = (index: number) => {
		const userIndex = index + curPage * pageSize;
		alert(users[userIndex].name);
	};
</script>

<div class="flex items-center justify-between">
	<div>
		<span>Show </span>
		<select
			class="select select-ghost select-bordered select-sm py-0 mx-1 max-w-xs"
			bind:value={pageSize}
		>
			{#each sizeOptions as size}
				<option selected={size === pageSize}>{size}</option>
			{/each}
		</select>
		<span> users</span>
	</div>
	<div class="">
		<input
			id="search-input"
			type="text"
			placeholder="Search..."
			class="input input-bordered max-w-xs input-sm"
		/>
	</div>
</div>

<div class="overflow-x-auto my-2">
	<table class="table">
		<thead>
			<tr>
				<th></th>
				<th>Name</th>
				<th>Role</th>
				<th></th>
			</tr>
		</thead>
		<tbody>
			{#each usersView as user, i}
				<tr>
					<th>
						<label><input type="checkbox" class="checkbox" /></label>
					</th>
					<td>
						<div class="flex items-center space-x-3">
							<div class="avatar">
								<AvatarMarble size={20} name={user.name} colors={avatarColors} />
							</div>
							<div>
								<div class="font-bold">{user.name}</div>
							</div>
						</div>
					</td>
					<td>
						<span class="badge badge-ghost badge-sm">{user.role}</span>
					</td>
					<th>
						<button class="btn btn-ghost btn-xs" onclick={() => editUser(i)}> edit </button>
					</th>
				</tr>
			{/each}
		</tbody>
	</table>
</div>

<div class="flex items-center justify-between">
	<div>
		<span
			>Showing {curPage * pageSize + 1} - {curPage * pageSize + usersView.length} of {users.length}</span
		>
	</div>

	<div class="join">
		<button
			class="join-item btn"
			class:btn-disabled={curPage === 0}
			onclick={() => {
				curPage = curPage - 1;
			}}
		>
			«
		</button>
		{#each Array(pages).keys() as num}
			<button
				class="join-item btn"
				class:btn-active={curPage === num}
				onclick={() => {
					curPage = num;
				}}
			>
				{num + 1}
			</button>
		{/each}
		<button
			class="join-item btn"
			class:btn-disabled={curPage + 1 === pages}
			onclick={() => {
				curPage = curPage + 1;
			}}
		>
			»
		</button>
	</div>
</div>
