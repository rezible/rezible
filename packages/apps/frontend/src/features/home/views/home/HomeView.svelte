<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { Button } from "$components/ui/button";
	import { initHomeController } from "./controller.svelte";
	import IncidentsSection from "./IncidentsSection.svelte";
	import SituationsSection from "./SituationsSection.svelte";
	import EventsSection from "./EventsSection.svelte";

	const controller = initHomeController();
	registerPageDescriptor(() => ({ title: "Home" }));
</script>

<div class="min-h-0 flex-1 overflow-y-auto p-4">
	<div class="mb-3 flex items-center justify-between gap-3">
		<p class="text-sm text-muted-foreground">Latest operational activity</p>
		<Button
			variant="outline"
			size="sm"
			onclick={controller.refresh}
			disabled={controller.refreshing}
		>
			Refresh
		</Button>
	</div>
	<div class="grid items-start gap-4 lg:grid-cols-2">
		<IncidentsSection />
		<SituationsSection />
		<EventsSection />
	</div>
</div>
