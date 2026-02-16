<script lang="ts">
	import { page } from "$app/state";
	import UserView from "$features/users/views/user";
</script>

<UserView id={page.params.id ?? ""} />