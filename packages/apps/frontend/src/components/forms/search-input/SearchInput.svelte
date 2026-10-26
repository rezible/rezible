<script lang="ts">
	import { Input } from "$components/ui/input";
	import { useDebounce } from "runed";

	type Props = {
		value: string | undefined;
		label?: string;
	};
	let { value = $bindable(), label = "Search" }: Props = $props();

	const id = $props.id();

	const commit = useDebounce((text: string) => {
		const trimmed = text.trim();
		if (trimmed) {
			value = trimmed;
		} else {
			value = undefined;
		}
	}, 300);
</script>

<div class="min-w-0">
	<label for={`${id}-search`} class="mb-1 block text-xs text-muted-foreground">{label}</label>
	<Input
		id={`${id}-search`}
		type="search"
		value={value ?? ""}
		oninput={(event) => commit(event.currentTarget.value)}
	/>
</div>
