<script lang="ts">
  import type { User } from "$lib/api/oapi.gen/types.gen";
  import { formatDate } from "$lib/utils/date";
  import UserAvatar from "./UserAvatar.svelte";
  
  export let user: User;
</script>

<div class="user-profile">
  <div class="profile-header">
    <UserAvatar {user} size="large" />
    <div class="user-info">
      <h1>{user.attributes.name}</h1>
      <div class="user-meta">
        <div class="meta-item">
          <span class="icon">✉️</span>
          <a href="mailto:{user.attributes.email}">{user.attributes.email}</a>
        </div>
        {#if user.attributes.timezone}
          <div class="meta-item">
            <span class="icon">🌐</span>
            <span>{user.attributes.timezone}</span>
            <span class="timezone-time">
              {new Date().toLocaleTimeString([], { timeZone: user.attributes.timezone, hour: '2-digit', minute: '2-digit' })}
            </span>
          </div>
        {/if}
        {#if user.attributes.chatId}
          <div class="meta-item">
            <span class="icon">💬</span>
            <span>Chat ID: {user.attributes.chatId}</span>
          </div>
        {/if}
        {#if user.attributes.createdAt}
          <div class="meta-item">
            <span class="icon">📅</span>
            <span>Member since {formatDate(new Date(user.attributes.createdAt))}</span>
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>

<style>
  .user-profile {
    background-color: var(--color-bg-card);
    border-radius: 8px;
    padding: 1.5rem;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .profile-header {
    display: flex;
    gap: 1.5rem;
    align-items: flex-start;
  }

  .user-info {
    flex: 1;
  }

  h1 {
    margin: 0 0 0.5rem 0;
    font-size: 1.8rem;
    font-weight: 600;
  }

  .user-meta {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    color: var(--color-text-secondary);
  }

  .meta-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .icon {
    width: 1.5rem;
    display: inline-flex;
    justify-content: center;
  }

  .timezone-time {
    margin-left: 0.5rem;
    font-weight: 500;
    color: var(--color-text-primary);
    background: var(--color-bg-subtle);
    padding: 0.1rem 0.4rem;
    border-radius: 4px;
  }
</style>
