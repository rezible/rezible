/**
 * Format a date to a human-readable string (YYYY-MM-DD)
 */
export function formatDate(date: Date): string {
  return date.toISOString().split('T')[0];
}

/**
 * Format a time to a human-readable string (HH:MM)
 */
export function formatTime(date: Date): string {
  return date.toTimeString().substring(0, 5);
}

/**
 * Format a date and time to a human-readable string (YYYY-MM-DD HH:MM)
 */
export function formatDateTime(date: Date): string {
  return `${formatDate(date)} ${formatTime(date)}`;
}
