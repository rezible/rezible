<script lang="ts">
	import { page } from "$app/state";
	import UserProfileView from "$features/users/views/user-profile/UserProfileView.svelte";
</script>

<UserProfileView userId={page.params.id} />