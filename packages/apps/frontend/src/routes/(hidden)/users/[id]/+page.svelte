<script lang="ts">
	import type { PageProps } from "./$types";
	import UserView from "$features/users/views/user";

	const { params }: PageProps = $props();
</script>

<UserView id={params.id} />
