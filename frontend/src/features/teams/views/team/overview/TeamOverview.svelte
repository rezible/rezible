<script lang="ts">
	import { mdiSlack } from "@mdi/js";
	import { SelectField } from "svelte-ux";
	import TeamUsersTable from "./TeamUsersTable.svelte";
	import TeamRostersTable from "./TeamRostersTable.svelte";

	const slackChannelOptions = [{ label: "team-channel", value: "foo" }];
	const slackChannel = slackChannelOptions[0].value;
</script>

<div class="flex flex-col gap-2">
	<div class="flex flex-col gap-2 w-fit p-2">
		<SelectField
			label="Slack Channel"
			labelPlacement="top"
			icon={mdiSlack}
			readonly
			options={slackChannelOptions}
			value={slackChannel}
			clearable={false}
		/>
	</div>

	<div class="flex gap-2">
		<TeamUsersTable />

		<TeamRostersTable />
	</div>
</div>
