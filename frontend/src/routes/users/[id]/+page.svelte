<script lang="ts">
	import { page } from "$app/state";
	import UserProfileView from "$features/user/views/user-profile/UserProfileView.svelte";
</script>

<UserProfileView userId={page.params.id} />