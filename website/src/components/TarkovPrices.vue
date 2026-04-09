<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref } from 'vue';
import wasmUrl from '@/assets/main.wasm?url';

const loading = ref(true); // wasm loaded into webpage?
const fetching = ref(false); // price data fetched?
const items = ref<any[]>([]);
const recipeItems = ref<any[]>([]);
const lastUpdateTimestamp = ref<number | null>(null);
const timeAgo = ref<string>('');

const CACHE_KEY = 'tarkov_prices_data';
const CACHE_TIME_KEY = 'tarkov_prices_timestamp';

const FIVE_MINUTES = 5 * 60 * 1000;
const ONE_MINUTE = 1 * 60 * 1000;
const THIRTY_SEC = 30 * 1000;

// Display in status bar
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

const autoRefresh = () => {
  if (!lastUpdateTimestamp.value || fetching.value) return;

  const now = Date.now();
  const diff = now - lastUpdateTimestamp.value;

  if (diff >= FIVE_MINUTES) {
    console.log("Auto refreshing...");
    fetchData(true);
  }
}

const fetchData = async (force = false) => {
  const now = Date.now();
  const cachedData = localStorage.getItem(CACHE_KEY);
  const cacheTime = localStorage.getItem(CACHE_TIME_KEY);

  // Check cache
  if (!force && cachedData && cacheTime && (now - Number(cacheTime) < FIVE_MINUTES)) {
    console.log("Loading from browser cache");
    const parsed = JSON.parse(cachedData);
    items.value = parsed.prices;
    recipeItems.value = parsed.recipes;
    lastUpdateTimestamp.value = Number(cacheTime);
    updateRelTime();
    return;
  }

  fetching.value = true;
  try {
    console.log("Fetching from API");
    const pricesRawData = await (window as any).getTarkovPrices();
    const recipesRawData = await (window as any).getTarkovRecipes();
    const parsedPrices = JSON.parse(pricesRawData);
    const parsedRecipes = JSON.parse(recipesRawData);
    items.value = parsedPrices;
    recipeItems.value = parsedRecipes;

    const combinedData = {
      prices: parsedPrices,
      recipes: parsedRecipes
    };

    // Set/update cache
    localStorage.setItem(CACHE_KEY, JSON.stringify(combinedData));
    localStorage.setItem(CACHE_TIME_KEY, now.toString());

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

    await fetchData();
  } catch (error) {
    console.error("Wasm load failed:", error);
  } finally {
    loading.value = false;
  }
};

let timerInterval: any = null;
let autoRefreshInterval: any = null;
onMounted(async () => {
  await loadWasmModule();
  localStorage.clear(); // Clear cached data
  timerInterval = setInterval(updateRelTime, THIRTY_SEC); // Update "ago" every 30 sec
  autoRefreshInterval = setInterval(autoRefresh, ONE_MINUTE); // autorefresh data every 1 min
});
// Timer cleanup
onUnmounted(() => {
  clearInterval(timerInterval);
  clearInterval(autoRefreshInterval);
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
        <template v-if="fetching">Syncing data...</template>
        <template v-else-if="timeAgo">{{ `Last sync: ${timeAgo}` }}</template>
        <template v-else>Initializing...</template>
      </span>
      <button v-if="!loading" @click="fetchData(true)" :disabled="fetching" class="refresh-button">
        Refresh
      </button>
    </div>

    <div v-if="!loading" class="list-container">
      <ul>
        <h3>Item Watchlist</h3>
        <li v-for="item in items" :key="item.shortName" class="item-row">
          <img :src="item.iconLink" class="item-icon" />
          {{ item.shortName }}
          <span class="price">{{ item.bestPrice.toLocaleString() }} &#x20BD;</span>
        </li>
      </ul>

      <ul>
        <h3>Recipe Products</h3>
        <li v-for="recipe in recipeItems" :key="recipe.id" class="recipe-card">
          <div class="item-row">
            <img :src="recipe.iconLink" class="item-icon">
            <strong>{{ recipe.name }}</strong>
          </div>

          <div v-for="(barter, index) in recipe.bartersFor" :key="index" class="requirements-list"
            :class="{ 'is-best-price': barter.isBestOption }">
            <div v-for="req in barter.requiredItems" :key="req.item.id" class="requirement-row">
              <img :src="req.item.iconLink" class="req-icon" />
              <span class="qty">{{ req.quantity }}x</span>
              <span class="name">{{ req.item.shortName }}</span>
            </div>

            <div class="total-row">
              <span>Total Cost: </span>
              <span class="price">
                {{ barter.totalCost.toLocaleString() }} &#x20BD;
              </span>
            </div>
          </div>
        </li>
      </ul>
    </div>

    <!-- <h3>Raw Data Debug:</h3> -->
    <!-- <pre>{{ JSON.stringify(items, null, 2) }}</pre> -->
    <!-- <pre>{{ JSON.stringify(recipeItems, null, 2) }}</pre> -->
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

.refresh-button {
  background: transparent;
  border: 1px solid #444;
  color: #888;
  padding: 2px 8px;
  cursor: pointer;
  font-size: 0.7rem;
  text-transform: uppercase;
}

.refresh-button:hover:not(:disabled) {
  border-color: #9a8866;
  color: #9a8866;
}

.list-container {
  display: flex;
  justify-content: center;
  gap: 40px;
  margin: 0 auto;
  list-style-type: none;
}

ul {
  padding: 0;
  margin: 0;
}

.item-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 6px;
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

.recipe-card {
  list-style: none;
  margin-bottom: 20px;
  background: #1a1a1a;
  padding: 10px;
  border-radius: 4px;
}

.requirements-list {
  margin-top: 8px;
  padding-left: 20px;
  border-left: 1px solid #444;

  &.is-best-price {
    border-left-color: #4caf50;
    background: rgba(76, 175, 80, 0.05);
  }
}

.requirement-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.8rem;
  color: #bbb;
  margin-top: 4px;
}

.req-icon {
  width: 20px;
  height: 20px;
}

.qty {
  color: #9a8866;
  font-weight: bold;
}

.total-row {
  display: flex;
  justify-content: space-between;
  margin-top: 10px;
  padding-top: 5px;
  border-top: 1px dashed #444;
  font-weight: bold;
  font-size: 0.85rem;
}
</style>