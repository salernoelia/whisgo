<template>
  <div class="history-card">
    <div class="card-header">
      <span class="timestamp">{{ formatTimestamp(time) }}</span>
    </div>
    <div class="card-content">
      <p>{{ text }}</p>
    </div>
    <div class="card-actions">
      <button @click="$emit('copy', text)" class="copy-button">
        Copy
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, defineEmits } from 'vue';

defineEmits(['copy']);

const props = defineProps<{
  id: number,
  text: string,
  time: string,
}>();

function formatTimestamp(timestamp: string): string {
  try {
    const date = new Date(timestamp);
    return date.toLocaleString();
  } catch (e) {
    return timestamp;
  }
}
</script>

<style scoped>
.history-card {
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  font-size: 0.8em;
  color: #aaa;
}

.timestamp {
  font-style: italic;
}

.card-content {
  font-size: 1em;
  color: white;
  line-height: 1.4;
}

.card-content p {
  margin: 0;
  word-break: break-word;
}

.card-actions {
  display: flex;
  justify-content: flex-end;
}

.copy-button {
  padding: 4px 12px;
  background-color: #3498db;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 0.9em;
  cursor: pointer;
  transition: background-color 0.2s;
}

.copy-button:hover {
  background-color: #2980b9;
}
</style>