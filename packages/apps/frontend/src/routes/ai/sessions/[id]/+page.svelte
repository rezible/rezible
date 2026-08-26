<script lang="ts">
	import { page } from "$app/state";
	import { SessionDetailView } from "$features/ai/views/session-detail";
</script>

<SessionDetailView sessionId={page.params.id ?? ""} />
