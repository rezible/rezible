<script lang="ts">
	import { setPageBreadcrumbs } from "$lib/appState.svelte";

	setPageBreadcrumbs(() => [
		{ label: "Users", href: "/users" },
		{label: "Name", href: "/users/foo"},
	]);
</script>

<span>todo user</span>
