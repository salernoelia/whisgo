<template>
  <div class="container">
    <div class="header">

      <div class="settings-row">
        <label for="model-selector">Model:</label>
        <select name="model" id="model-selector" v-model="model">
          <option value="distil-whisper-large-v3-en">distil-whisper-large-v3-en</option>
          <option value="whisper-large-v3">whisper-large-v3</option>
          <option value="whisper-large-v3-turbo">whisper-large-v3-turbo</option>
        </select>
      </div>
       
      <div class="settings-row">
        <label for="device-selector">Input Device:</label>
        <select name="device" id="device-selector" v-model="selectedDevice" @change="changeDevice">
          <option v-for="device in audioDevices" :key="device.id" :value="device.id">
            {{ device.name }}
          </option>
        </select>
      </div>
    </div>

    <div class="recording-controls">
      <button class="record-button" :class="{ 'recording': isRecording }" @click="toggleRecording">
        {{ isRecording ? 'Stop Recording' : 'Start Recording' }}
      </button>

      <div v-if="recordingStatus" class="status-message">
        {{ recordingStatus }}
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { GetAudioDevices, SetSelectedDevice, StartRecordingMicrophone, StopRecordingMicrophone, IsRecording } from '../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';

interface AudioDevice {
  id: string;
  name: string;
}

const model = ref('distil-whisper-large-v3-en');
const groqKey = ref('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');
const audioDevices = ref<AudioDevice[]>([]);
const selectedDevice = ref('');
const isRecording = ref(false);
const recordingStatus = ref('');

onMounted(async () => {
  try {
    const devices = await GetAudioDevices();
    audioDevices.value = devices;
    if (devices.length > 0) {
      selectedDevice.value = devices[0].id;
      await changeDevice();
    }

    // Check initial recording state
    isRecording.value = await IsRecording();

    // Listen for recording events
    EventsOn('recording-started', (result) => {
      isRecording.value = true;
      recordingStatus.value = result as string;
    });

    EventsOn('recording-stopped', () => {
      isRecording.value = false;
      recordingStatus.value = 'Recording stopped';
    });
  } catch (error) {
    console.error('Error in setup:', error);
    recordingStatus.value = 'Failed to initialize';
  }
});

// Clean up event listeners
onUnmounted(() => {
  EventsOff('recording-started');
  EventsOff('recording-stopped');
});

async function changeDevice() {
  try {
    await SetSelectedDevice(selectedDevice.value);
  } catch (error) {
    console.error('Error setting audio device:', error);
  }
}

async function toggleRecording() {
  try {
    if (isRecording.value) {
      recordingStatus.value = await StopRecordingMicrophone();
      isRecording.value = false;
    } else {
      recordingStatus.value = await StartRecordingMicrophone();
      isRecording.value = true;
    }
  } catch (error) {
    console.error('Error toggling recording:', error);
    recordingStatus.value = 'Error with recording';
    isRecording.value = false;
  }
}
</script>

<style>
.container {
  padding: 20px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  max-width: 500px;
  margin: 0 auto;
}

.header {
  margin-bottom: 20px;
}

h1 {
  font-size: 24px;
  margin-bottom: 15px;
}

.settings-row {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

label {
  font-weight: 500;
  width: 100px;
}

#model-selector,
#device-selector {
  flex: 1;
  font-size: 1em;
  padding: 8px;
  border-radius: 4px;
  border: 1px solid #ccc;
}

.recording-controls {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-top: 20px;
}

.record-button {
  padding: 12px 24px;
  font-size: 16px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.record-button:hover {
  background-color: #45a049;
}

.record-button.recording {
  background-color: #f44336;
}

.record-button.recording:hover {
  background-color: #d32f2f;
}

.status-message {
  margin-top: 10px;
  font-size: 14px;
  color: #666;
}
</style>