<script lang="ts">
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";

	setPageBreadcrumbs(() => [
		{ label: "Users", href: "/users" },
		{label: "Name", href: "/users/foo"},
	]);
</script>

<span>todo user</span>
