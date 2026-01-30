<script lang="ts">
	import { Field } from 'svelte-ux';
	import type { ConfigComponentProps } from '../types';
	import { onMount } from "svelte";

    const { integration, configured, onConfigChange }: ConfigComponentProps = $props();

    onMount(() => {

    });

    let inputEl = $state<HTMLInputElement>(null!);
    const onFileInputChange = async () => {
        if (!inputEl.files) return;
        for (let i = 0; i < inputEl.files.length; i++) {
            const file = inputEl.files.item(i);
            if (!file) continue;
            const data = await file.text();
            onConfigChange("ServiceAccountCredentials", data);
        }
    }
</script>

<span>google config</span>

<Field label="Service Account Credentials" let:id>
  <input {id}
    bind:this={inputEl}
    onchange={onFileInputChange}
    max={1}
    type="file"
    class="w-full outline-none bg-surface-100"
    accept=".json"
  />
</Field>
