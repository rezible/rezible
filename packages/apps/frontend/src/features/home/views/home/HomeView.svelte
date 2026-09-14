<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initHomeController } from "./controller.svelte";
	import IncidentsSection from "./IncidentsSection.svelte";
	import SituationsSection from "./SituationsSection.svelte";
	import InboxSection from "./InboxSection.svelte";
	import ActivitySection from "./ActivitySection.svelte";
	import HomeViewPageActions from "./HomeViewPageActions.svelte";

	initHomeController();
	registerPageDescriptor(() => ({ 
		title: "Overview",
		parents: [{ label: "Home", path: "/" }],
		actions: {
			component: HomeViewPageActions,
		}
	}));
</script>

<div class="@container/dashboard min-h-0 flex-1 overflow-y-auto font-sans">
	<div
		class="grid grid-cols-1 gap-6 p-4 @min-[800px]/dashboard:p-6 @min-[1310px]/dashboard:grid-cols-[minmax(0,3fr)_minmax(0,1.85fr)]"
	>
		<div class="@container/main min-w-0 space-y-7">
			<IncidentsSection />
			<SituationsSection />
			<InboxSection />
		</div>
		<ActivitySection />
	</div>
</div>
