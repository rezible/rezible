<script lang="ts">
  import type { User } from "$lib/api/oapi.gen/types.gen";
  
  export let user: User;
  export let size: "small" | "medium" | "large" = "medium";
  
  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map(part => part.charAt(0))
      .join('')
      .toUpperCase()
      .substring(0, 2);
  };
  
  const getColorFromName = (name: string) => {
    const colors = [
      "#4299E1", "#48BB78", "#ED8936", "#9F7AEA", 
      "#F56565", "#38B2AC", "#ECC94B", "#667EEA"
    ];
    
    let hash = 0;
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }
    
    return colors[Math.abs(hash) % colors.length];
  };
  
  const initials = getInitials(user.attributes.name || "User");
  const bgColor = getColorFromName(user.attributes.name || "User");
  
  const sizeClass = {
    small: "w-8 h-8 text-xs",
    medium: "w-12 h-12 text-sm",
    large: "w-20 h-20 text-xl"
  }[size];
</script>

<div 
  class="avatar {sizeClass}" 
  style="background-color: {bgColor};"
  title="{user.attributes.name}"
>
  {initials}
</div>

<style>
  .avatar {
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    color: white;
    font-weight: 600;
    flex-shrink: 0;
  }
</style>
