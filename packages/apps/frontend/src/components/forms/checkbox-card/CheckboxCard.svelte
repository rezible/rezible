<script lang="ts">
	import { Checkbox } from "$src/components/ui/checkbox";
	import { Label } from "$src/components/ui/label";

    type Props = {
        checked: boolean;
        label: string;
        description?: string;
    };
    let {
        checked = $bindable(),
        label,
        description,
    }: Props = $props();
    const id = $props.id();
</script>

<Label class="hover:bg-accent/50 flex items-start gap-3 rounded-lg border p-3 has-[[aria-checked=true]]:border-accent-600 has-[[aria-checked=true]]:bg-accent-50 dark:has-[[aria-checked=true]]:border-accent-900 dark:has-[[aria-checked=true]]:bg-accent-950">
    <Checkbox {id} 
        bind:checked
        class="data-[state=checked]:border-accent data-[state=checked]:bg-accent data-[state=checked]:text-white dark:data-[state=checked]:border-primary dark:data-[state=checked]:bg-primary"
    />
    <div class="grid gap-1.5 font-normal">
        <p class="text-sm leading-none font-medium">{label}</p>
        <p class="text-muted-foreground text-sm">{description}</p>
    </div>
</Label>