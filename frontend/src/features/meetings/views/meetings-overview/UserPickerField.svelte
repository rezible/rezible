<script lang="ts">
	import { Field } from "svelte-ux";

	type Attendee = {
		type: "person" | "team";
		id: string;
	};
	type Props = {
		value: Attendee[];
	};
	let { value = $bindable() }: Props = $props();
</script>

<div>
	<Field label="Attendees">
		<span>todo</span>
	</Field>
</div>
