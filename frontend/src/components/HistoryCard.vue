<template>
  <div id="container">
    <div id="header">
      <p>ID:{{ id }},</p>
      <p>Time: {{ new Date(time).toLocaleString() }}</p>
    </div>
    <div id="content">
      <p>{{ text }}</p>

    </div>
    <button id="copy-button" @click="copyToClipboard(text)">Copy to Clipboard</button>
  </div>

</template>

<script setup lang="ts">
import { defineProps } from 'vue';



const props = defineProps<{
  id: number,
  text: string,
  time: string,
}>();


function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text);
}


</script>

<style scoped>
p {
  margin: 0;
  text-align: left;
}

#container {
  margin: 0 10px;
  padding: 5px;
  border-radius: 5px;
  border: 1px solid #ccc;
  margin-bottom: 10px;

}

#header {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  margin-bottom: 10px;
}

#content {
  margin-bottom: 10px;
}

#copy-button {
  padding: 8px 16px;
}
</style>