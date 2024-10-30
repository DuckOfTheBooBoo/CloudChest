<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";

const router = useRouter();
const route = useRoute();
</script>

<template>
  <main>
    <RouterView />
  </main>
</template>

<style scoped></style>
