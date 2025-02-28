<template>
  <div class="container">

    <label for="groq-key-input">Groq Key:</label>
    <input type="text" v-model="groqKey" id="groq-key-input" placeholder="Enter your GROQ key" />
    <button @click="saveAPIKey">Save API Key</button>

    <button @click="ClearRecordingsDir">Clear Recordings</button>

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
    <button class="record-button" :class="{ 'recording': isRecording }" @click="toggleRecording" :disabled="!groqKey">
      {{ isRecording ? 'Stop Recording' : 'Start Recording' }}
    </button>

    <div v-if="recordingStatus" class="status-message">
      {{ recordingStatus }}
    </div>
    <div v-if="!groqKey" class="error-message">
      Please enter your Groq API key to start recording.
    </div>
  </div>

  <div class="transcription-area">
    <h2>Current Transcription:</h2>
    <textarea v-model="currentTranscription" rows="5" cols="50" readonly></textarea>
    <button @click="copyToClipboard" :disabled="!currentTranscription">Copy to Clipboard</button>
  </div>

  <div class="history-area">
    <h2>Transcription History:</h2>
    <ul>
      <li v-for="(transcription, index) in transcriptionHistory" :key="index">
        {{ transcription }}
      </li>
    </ul>

  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted } from 'vue';
import { GetAudioDevices, SetSelectedDevice, StartRecordingMicrophone, StopRecordingMicrophone, IsRecording, ClearRecordingsDir, GetGroqAPIKey, SetGroqAPIKey, GetTranscriptionHistory } from '../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';

interface AudioDevice {
  id: string;
  name: string;
}

const model = ref('distil-whisper-large-v3-en');
const groqKey = ref('');
const audioDevices = ref<AudioDevice[]>([]);
const selectedDevice = ref('');
const isRecording = ref(false);
const recordingStatus = ref('');
const currentTranscription = ref('');
const transcriptionHistory = ref<string[]>([]);

onMounted(async () => {
  try {
    // Load API key from config
    groqKey.value = await GetGroqAPIKey();

    const devices = await GetAudioDevices();
    audioDevices.value = devices;
    if (devices.length > 0) {
      selectedDevice.value = devices[0].id;
      await changeDevice();
    }

    // Check initial recording state
    isRecording.value = await IsRecording();

    // Load transcription history
    transcriptionHistory.value = await GetTranscriptionHistory();

    // Listen for recording events
    EventsOn('recording-started', (result) => {
      isRecording.value = true;
      recordingStatus.value = result as string;
    });

    EventsOn('recording-stopped', (result) => {
      isRecording.value = false;
      recordingStatus.value = result as string;
      currentTranscription.value = result as string;
      updateTranscriptionHistory();
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

async function saveAPIKey() {
  try {
    const result = await SetGroqAPIKey(groqKey.value);
    recordingStatus.value = result;
  } catch (error) {
    console.error('Error saving API key:', error);
    recordingStatus.value = 'Failed to save API key';
  }
}

async function copyToClipboard() {
  try {
    await navigator.clipboard.writeText(currentTranscription.value);
    recordingStatus.value = 'Transcription copied to clipboard';
  } catch (error) {
    console.error('Error copying to clipboard:', error);
    recordingStatus.value = 'Failed to copy to clipboard';
  }
}

async function updateTranscriptionHistory() {
  try {
    transcriptionHistory.value = await GetTranscriptionHistory();
  } catch (error) {
    console.error('Error getting transcription history:', error);
  }
}
</script>

<style>
.container {
  display: flex;
  flex-direction: column;
  padding: 20px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  max-width: 500px;
  margin: 0 auto;
  gap: 0;
}


h1 {
  font-size: 24px;
  margin-bottom: 15px;
}

.settings-row {
  display: flex;
  align-items: center;
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

.transcription-area,
.history-area {
  margin-top: 20px;
  border-top: 1px solid #eee;
  padding-top: 20px;
}

textarea {
  width: 100%;
  padding: 8px;
  border-radius: 4px;
  border: 1px solid #ccc;
  margin-bottom: 10px;
}

ul {
  list-style-type: none;
  padding: 0;
}

li {
  margin-bottom: 5px;
}

.error-message {
  color: red;
  margin-top: 10px;
}
</style>