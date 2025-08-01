<script lang="ts">
	import { page } from "$app/state";
	import { UserView, setUserViewState } from "$features/user";

	setUserViewState(() => page.params.id);
</script>

<UserView />