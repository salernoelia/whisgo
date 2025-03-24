<template>
  <div class="container">
    <div class="settings-panel">
      <h1>Whisgo</h1>

      <div class="settings-row">
        <label for="groq-key-input">Groq API Key:</label>
        <input type="text" v-model="groqKey" id="groq-key-input" placeholder="Enter your GROQ key" @blur="saveAPIKey" />
        <button @click="saveAPIKey">Save</button>
      </div>

      <div class="settings-row">
        <label for="model-selector">Model:</label>
        <select name="model" id="model-selector" v-model="selectedModel" @change="saveModel">
          <option value="distil-whisper-large-v3-en">distil-whisper-large-v3-en</option>
          <option value="whisper-large-v3">whisper-large-v3</option>
          <option value="whisper-large-v3-turbo">whisper-large-v3-turbo</option>
        </select>
      </div>

      <div class="settings-row">
        <label for="device-selector">Input Device:</label>
        <select name="device" id="device-selector" v-model="selectedDeviceId" @change="changeDevice">
          <option v-for="device in audioDevices" :key="device.deviceId" :value="device.deviceId">
            {{ device.label || `Microphone ${device.deviceId.slice(0, 5)}...` }}
          </option>
        </select>
        <button @click="refreshDevices">Refresh</button>
      </div>

      <div class="recording-controls">
        <button class="record-button" :class="{ 'recording': isRecording }" @click="toggleRecording"
          :disabled="!groqKey || isProcessing">
          {{ isRecording ? 'Stop Recording' : 'Start Recording' }}
        </button>

        <div v-if="recordingStatus" class="status-message">
          {{ recordingStatus }}
        </div>

        <div v-if="!groqKey" class="error-message">
          Please enter your Groq API key to start recording.
        </div>

        <div v-if="audioError" class="error-message">
          {{ audioError }}
        </div>
      </div>

      <div class="action-buttons">
        <button @click="copyToClipboard(currentTranscription)" :disabled="!currentTranscription">
          Copy Latest to Clipboard
        </button>
        <button @click="clearTranscriptionHistory">
          Clear History
        </button>
      </div>
    </div>

    <div class="history-panel">
      <h2>Transcription History</h2>
      <div class="history-list">
        <HistoryCard v-for="transcription in transcriptions" :key="transcription.id" :id="transcription.id"
          :text="transcription.text" :time="transcription.timestamp" @copy="copyToClipboard" />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime';
import HistoryCard from './components/HistoryCard.vue';
import { useGroq } from './composables/useGroq';
import { useAudioRecording } from './composables/useAudioRecording';
import { useTranscriptionHistory } from './composables/useTranscriptionHistory';

// Initialize composables
const {
  apiKey: groqKey,
  model: selectedModel,
  isProcessing,
  saveApiKey,
  saveModel: saveModelSetting, // Renamed to avoid conflict
  transcribeAudio
} = useGroq();

const {
  isRecording,
  audioData,
  error: audioError,
  availableDevices,
  selectedDevice: selectedDeviceId,
  initAudioDevices,
  selectDevice,
  startRecording,
  stopRecording
} = useAudioRecording();

const {
  history: transcriptions,
  addTranscription,
  clearHistory
} = useTranscriptionHistory();

// Local refs
const recordingStatus = ref('');
const currentTranscription = ref('');

// Computed for UI
const audioDevices = computed(() => {
  return availableDevices.value;
});

onMounted(async () => {
  try {
    // Initialize audio devices
    await initAudioDevices();

    // Listen for hotkey trigger event from Go backend
    EventsOn('hotkey-triggered', () => {
      toggleRecording();
    });
  } catch (error) {
    console.error('Error during initialization:', error);
    recordingStatus.value = 'Failed to initialize';
  }
});

onUnmounted(() => {
  EventsOff('hotkey-triggered');
});

async function changeDevice() {
  try {
    selectDevice(selectedDeviceId.value);
  } catch (error) {
    console.error('Error setting audio device:', error);
  }
}

async function refreshDevices() {
  try {
    await initAudioDevices();
    recordingStatus.value = 'Devices refreshed';
  } catch (error) {
    console.error('Error refreshing devices:', error);
    recordingStatus.value = 'Failed to refresh devices';
  }
}

async function toggleRecording() {
  try {
    if (isRecording.value) {
      recordingStatus.value = 'Stopping recording...';
      const audioBlob = await stopRecording();

      // Process the audio
      recordingStatus.value = 'Transcribing...';
      const transcription = await transcribeAudio(audioBlob);

      // Add to history and update UI
      addTranscription(transcription);
      currentTranscription.value = transcription;
      await copyToClipboard(transcription);
      recordingStatus.value = 'Transcription complete';
    } else {
      if (!groqKey.value) {
        recordingStatus.value = 'Please set your Groq API key';
        return;
      }

      await startRecording();
      recordingStatus.value = 'Recording started';
    }
  } catch (error) {
    console.error('Error during recording process:', error);
    recordingStatus.value = error instanceof Error ? error.message : 'Error with recording';
    isRecording.value = false;
  }
}

function saveAPIKey() {
  try {
    recordingStatus.value = saveApiKey(groqKey.value);
  } catch (error) {
    console.error('Error saving API key:', error);
    recordingStatus.value = 'Failed to save API key';
  }
}

function saveModel() {
  try {
    recordingStatus.value = saveModelSetting(selectedModel.value);
  } catch (error) {
    console.error('Error saving model:', error);
    recordingStatus.value = 'Failed to save model';
  }
}

async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text);
    recordingStatus.value = 'Copied to clipboard';
  } catch (error) {
    console.error('Error copying to clipboard:', error);
    recordingStatus.value = 'Failed to copy to clipboard';
  }
}

function clearTranscriptionHistory() {
  clearHistory();
  recordingStatus.value = 'History cleared';
}
</script>

<style scoped>
.container {
  display: flex;
  flex-direction: column;
  padding: 20px;
  gap: 20px;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  max-width: 800px;
  margin: 0 auto;
  height: calc(100vh - 40px);
  color: white;
}

h1 {
  font-size: 24px;
  margin-bottom: 15px;
}

h2 {
  font-size: 18px;
  margin-bottom: 10px;
}

.settings-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 15px;
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: 8px;
}

.settings-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

label {
  font-weight: 500;
  min-width: 100px;
}

input[type="text"],
select {
  flex: 1;
  font-size: 1em;
  padding: 8px;
  border-radius: 4px;
  border: 1px solid #444;
  background-color: #333;
  color: white;
}

.recording-controls {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  margin: 15px 0;
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

.record-button:hover:not(:disabled) {
  background-color: #45a049;
}

.record-button.recording {
  background-color: #f44336;
}

.record-button.recording:hover:not(:disabled) {
  background-color: #d32f2f;
}

.record-button:disabled {
  background-color: #555;
  cursor: not-allowed;
}

.status-message {
  margin-top: 5px;
  font-size: 14px;
  color: #aaa;
}

.error-message {
  color: #ff6b6b;
  margin-top: 5px;
  font-size: 14px;
}

.action-buttons {
  display: flex;
  gap: 10px;
  justify-content: space-between;
}

button {
  padding: 8px 16px;
  background-color: #3498db;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
}

button:hover:not(:disabled) {
  background-color: #2980b9;
}

button:disabled {
  background-color: #555;
  cursor: not-allowed;
}

.history-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.history-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-right: 5px;
}

/* Custom scrollbar */
.history-list::-webkit-scrollbar {
  width: 6px;
}

.history-list::-webkit-scrollbar-track {
  background: #333;
}

.history-list::-webkit-scrollbar-thumb {
  background-color: #666;
  border-radius: 6px;
}
</style>