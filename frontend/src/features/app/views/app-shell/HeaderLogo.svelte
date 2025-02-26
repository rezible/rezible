<script lang="ts">
    type Props = {
        expandingHover: boolean;
    }
    const { expandingHover }: Props = $props();
    
</script>
<div class="h-16 flex items-center px-4">
    <a href="/" class="text-2xl flex items-center">
        <img src="/images/logo.svg" alt="logo" class="h-10 w-10 fill-neutral" />
        <span class="pl-3 {expandingHover ? 'hidden group-hover:inline' : ''}">Rezible</span>
    </a>
</div>