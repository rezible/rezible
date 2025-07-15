<script lang="ts">
	type Props = {
		slug: string;
	};
	const { slug }: Props = $props();
</script>

<span>{slug}</span>