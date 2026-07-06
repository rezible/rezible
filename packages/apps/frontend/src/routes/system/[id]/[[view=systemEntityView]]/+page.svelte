<script lang="ts">
	import { page } from "$app/state";
	import SystemEntityView from "$src/features/system/views/system-entity";
</script>

<SystemEntityView id={page.params.id || ""} />
