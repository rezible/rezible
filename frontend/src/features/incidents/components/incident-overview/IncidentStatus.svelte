<script lang="ts">
	import {
		mdiAccountDetails,
		mdiChartBox,
		mdiFileDocumentCheck,
		mdiFileDocumentEdit,
		mdiFire,
		mdiVideoAccount,
	} from "@mdi/js";

	import { Steps } from "svelte-ux";

	const stepsWithIcon = [
		{ label: "Incident Settled", completed: true, icon: mdiFire },
		{ label: "Debriefs", completed: true, icon: mdiAccountDetails },
		{
			label: "Retrospective",
			completed: false,
			icon: mdiFileDocumentCheck,
		},
		{
			label: "Review Meeting Scheduled",
			completed: false,
			icon: mdiVideoAccount,
		},
		{ label: "Report Available", completed: false, icon: mdiChartBox },
	];
</script>

<Steps data={stepsWithIcon} />
