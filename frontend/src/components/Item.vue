<script setup lang="ts">
const props = defineProps<{class: string}>()
</script>

<template>
  <v-card max-width="10rem" class="pa-2 rounded-lg" elevation="10" v-bind="props">
    <!-- Upper part (file name and menu) -->
    <div
      class="tw-flex tw-flex-row tw-h-full tw-mb-3 tw-w-full tw-items-center tw-flex-wrap"
    >
      <p class="text-body-2 tw-grow">File name.jpg</p>
      <v-menu>
        <template v-slot:activator="{ props }">
          <v-btn
            density="compact"
            icon="mdi-dots-vertical"
            variant="plain"
            v-bind="props"
          ></v-btn>
        </template>
        <v-list>
          <v-list-item>
            <v-list-item-title>Details</v-list-item-title>
            <v-list-item-title>Star</v-list-item-title>
            <v-list-item-title>Delete</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </div>

    <div
      class="tw-flex tw-justify-center tw-items-center tw-mb-2 tw-w-full tw-h-16 tw-rounded-lg bg-grey-darken-3"
    >
      <v-icon icon="mdi-trash-can"></v-icon>
    </div>

    <!-- Bottom part (date) -->
    <p class="text-caption">1 day ago</p>
  </v-card>
</template>

<script setup></script>
