<script lang="ts">

	type Props = {
		title: string;
		value: any;
		description: string;
	}

	const { title, value, description }: Props = $props();

</script>

<div>
	<div class="flex flex-col justify-between mb-1">
		<span class="text-surface-content">{title}</span>
		<span class="font-medium text-xl">{value}</span>
	</div>
	<div class="text-sm text-gray-500 mt-1">{description}</div>
</div>