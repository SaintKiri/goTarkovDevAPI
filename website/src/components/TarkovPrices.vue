<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import wasmUrl from '@/assets/main.wasm?url';

const loading = ref(true);
const items = ref<any[]>([]);

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

    loading.value = false;
    console.log("Wasm Ready!");

    const rawData = await (window as any).getTarkovPrices();
    items.value = JSON.parse(rawData);
  } catch (error) {
    console.error("Wasm load failed:", error);
    loading.value = false;
  }
};

onMounted(loadWasmModule);
</script>

<template>
  <div>
    <h2>Tarkov Prices</h2>

    <div v-if="loading">
      <p v-if="!items.length">Loading</p>
    </div>

    <div v-else>
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
  </div>
</template>

<style scoped>
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

.price {
  color: #9a8866;
  /* Gold/Tan color like the Flea Market */
  font-weight: bold;
}
</style>