<script lang="ts">
	import { page } from "$app/state";
	import { AlertView, setAlertViewState } from "$features/alert";

	setAlertViewState(() => (page.params.id || ""));
</script>

<AlertView />