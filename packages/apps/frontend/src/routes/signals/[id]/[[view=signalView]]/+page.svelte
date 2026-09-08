<script lang="ts">
	import { page } from "$app/state";
	import AlertView from "$features/signals/views/signal";
</script>

<AlertView id={page.params.id || ""} />
