<script setup lang="ts">
import { Ref, ref, onBeforeMount } from "vue";
import { CloudChestFile } from "../../models/file";
import File from "../File.vue";
import { getFavoriteFiles } from "../../utils/filesApi";
import { useEventEmitterStore } from "../../stores/eventEmitterStore";
import { FILE_UPDATED } from "../../constants";

const fileList: Ref<CloudChestFile[]> = ref([] as CloudChestFile[]);

const isLoading = ref<boolean>(false);


const eventEmitter = useEventEmitterStore();

eventEmitter.eventEmitter.on(FILE_UPDATED, fetchFavoriteFiles);

onBeforeMount(fetchFavoriteFiles);

async function fetchFavoriteFiles(): Promise<void> {
  isLoading.value = true;
  const response = await getFavoriteFiles();
  fileList.value = response.files;
  isLoading.value = false;
}
</script>

<template>
  <div class="tw-min-h-1">
    <v-progress-linear
      v-if="isLoading"
      :indeterminate="true"
      color="primary"
    ></v-progress-linear>
  </div>
  <v-container>
    <v-row>
      <v-col v-for="file in fileList" :key="file" :cols="2">
        <File :file="file" />
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped></style>
