<script lang="ts" setup>
import { onMounted, onUnmounted, ref } from 'vue';
import wasmUrl from '@/assets/main.wasm?url';

const loading = ref(true); // wasm loaded into webpage?
const fetching = ref(false); // price data fetched?
const items = ref<any[]>([]);
const lastUpdate = ref<string | null>(null);
const lastUpdateTimestamp = ref<number | null>(null);
const timeAgo = ref<string>('');

const CACHE_KEY = 'tarkov_prices_data';
const CACHE_TIME_KEY = 'tarkov_prices_timestamp';
const FIVE_MINUTES = 5 * 60 * 1000;

const updateRelTime = () => {
  if (!lastUpdateTimestamp.value) { timeAgo.value = 'Never'; return; }

  const now = Date.now();
  const diffInSec = Math.floor((now - lastUpdateTimestamp.value) / 1000);

  if (diffInSec < 60) {
    timeAgo.value = 'Just Now';
  } else {
    const minutes = Math.floor(diffInSec / 60);
    timeAgo.value = `${minutes}m ago`;
  }
};

const fetchPrices = async (force = false) => {
  const now = Date.now();
  const cachedData = localStorage.getItem(CACHE_KEY);
  const cacheTime = localStorage.getItem(CACHE_TIME_KEY);

  // Check cache
  if (!force && cachedData && cacheTime && (now - Number(cacheTime) < FIVE_MINUTES)) {
    console.log("Loading from browser cache");
    items.value = JSON.parse(cachedData);
    lastUpdate.value = new Date(Number(cacheTime)).toLocaleTimeString();
    lastUpdateTimestamp.value = Number(cacheTime);
    loading.value = false;
    updateRelTime();
    return;
  }

  fetching.value = true;
  try {
    console.log("Fetching from API");
    const rawData = await (window as any).getTarkovPrices();
    items.value = JSON.parse(rawData);

    // Set/update cache
    localStorage.setItem(CACHE_KEY, rawData);
    localStorage.setItem(CACHE_TIME_KEY, now.toString());

    lastUpdate.value = new Date(Number(now)).toLocaleTimeString();
    lastUpdateTimestamp.value = now;
    updateRelTime();
  } catch (error) {
    console.error("API fetch failed:", error);
  } finally {
    fetching.value = false;
  }
};

const loadWasmModule = async () => {
  try {
    // 1. Initialize the Go runner from the script you just found
    const go = new (window as any).Go();

    const response = await fetch(wasmUrl);

    // 2. Use go.importObject (this provides all the Go-specific "wiring")
    const { instance } = await WebAssembly.instantiateStreaming(
      response,
      go.importObject
    );

    // 3. Run the Go program (this executes main() and sets up your functions)
    go.run(instance);
    console.log("Wasm Ready!");

    await fetchPrices();
  } catch (error) {
    console.error("Wasm load failed:", error);
  } finally {
    loading.value = false;
  }
};

let timerInterval: any = null;
onMounted(async () => {
  await loadWasmModule();
  timerInterval = setInterval(updateRelTime, 30000); // Update "ago" every 30 sec
});
// Timer cleanup
onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval);
});
</script>

<template>
  <div>
    <div style="display:flex; align-items: center; gap: 15px;">
      <h1>Tarkov Prices (PvE)</h1>
    </div>

    <div class="status-bar" :class="{ 'is-fetching': fetching }">
      <div v-if="fetching" class="spinner"></div>
      <span class="status-text">
        {{ fetching ? 'Syncing data...' : `Last sync: ${timeAgo}` }}
      </span>
      <button v-if="!loading" @click="fetchPrices(true)" :disabled="fetching" class="refresh-button">
        Refresh
      </button>
    </div>

    <ul>
      <li v-for="item in items" :key="item.shortName" class="item-row">
        <img :src="item.iconLink" class="item-icon" />
        {{ item.shortName }}
        <span class="price">{{ item.bestPrice.toLocaleString() }} &#x20BD;</span>
      </li>
    </ul>
    <!-- <h3>Raw Data Debug:</h3>
      <pre>{{ JSON.stringify(items, null, 2) }}</pre> -->
  </div>
</template>

<style scoped>
.status-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 30px;
  /* Keeps layout stable */
  margin-top: -10px;
  margin-bottom: 15px;
  padding: 0 10px;
  border-left: 3px solid #555;
  /* Default inactive border */
  transition: all 0.3s ease;
}

/* Change style when fetching */
.status-bar.is-fetching {
  border-left-color: #9a8866;
  /* Tarkov gold/tan */
  background: rgba(154, 136, 102, 0.1);
}

.status-text {
  flex-grow: 1;
  font-size: 0.85rem;
  color: #888;
  font-family: 'Courier New', Courier, monospace;
  /* Terminal feel */
}

.is-fetching .status-text,
.price {
  color: #9a8866;
  font-weight: bold;
}

.spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(154, 136, 102, 0.2);
  border-top-color: #9a8866;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.refresh-btn {
  background: transparent;
  border: 1px solid #444;
  color: #888;
  padding: 2px 8px;
  cursor: pointer;
  font-size: 0.7rem;
  text-transform: uppercase;
}

.refresh-btn:hover:not(:disabled) {
  border-color: #9a8866;
  color: #9a8866;
}

.item-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
  background: #1a1a1a;
  /* Dark Tarkov-style theme */
  padding: 8px;
  border-radius: 4px;
  color: #fff;
}

.item-icon {
  width: 36px;
  height: 36px;
  object-fit: contain;
  background: #333;
  border: 1px solid #555;
}
</style>