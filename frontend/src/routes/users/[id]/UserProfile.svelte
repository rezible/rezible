<script lang="ts">
	import type { User } from "$lib/api";
	import Avatar from "$components/avatar/Avatar.svelte";
	import { getLocalTimeZone } from "@internationalized/date";

	type Props = {
		user: User;
	};
	const { user }: Props = $props();

	const timezone = $derived(getLocalTimeZone());
</script>

<div class="">
	<div class="">
		<Avatar kind="user" id={user.id} />
		<div class="">
			<h1>{user.attributes.name}</h1>
			<div class="">
				<div class="">
					<span class="">✉️</span>
					<a href="mailto:{user.attributes.email}">{user.attributes.email}</a>
				</div>
				<div class="">
					<span class="">🌐</span>
					<span>{timezone}</span>
					<span class="timezone-time">
						{new Date().toLocaleTimeString([], {
							timeZone: timezone,
							hour: "2-digit",
							minute: "2-digit",
						})}
					</span>
				</div>
				<div class="">
					<span class="">💬</span>
					<span>Chat ID: {user.id}</span>
				</div>
			</div>
		</div>
	</div>
</div>
