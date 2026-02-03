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
  } catch (error) {
    console.error("Wasm load failed:", error);
  }
};

const handleGetPrices = async () => {
  // 4. In Go, you likely used js.Global().Set("getTarkovPrices", ...)
  // So we call it directly from the window.
  try {
    const rawData = await (window as any).getTarkovPrices();
    items.value = JSON.parse(rawData);
  } catch (err) {
    console.error("Error fetching prices:", err);
  }
};

onMounted(loadWasmModule);
</script>

<template>
  <div>
    <h2>Tarkov Prices (Go + Wasm)</h2>
    <p v-if="loading">Initializing Go Runtime...</p>
    <div v-else>
      <button @click="handleGetPrices">Fetch Prices</button>
      <ul>
        <li v-for="item in items" :key="item.name">
          {{ item.name }}: {{ item.price }} RUB
        </li>
      </ul>
    </div>
  </div>
</template>