<script lang="ts">
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
	<span>user picker todo</span>
</div>
