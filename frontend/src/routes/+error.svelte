<script>
	import { page } from "$app/state";
	import ErrorView from "$features/app/views/error-boundary/ErrorBoundary.svelte";

	console.log("error page", page.error);
</script>

<ErrorView />