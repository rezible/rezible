<script lang="ts">
	import { page } from "$app/state";
	import UserView from "$features/user/view/UserView.svelte";
</script>

<UserView userId={page.params.id} />