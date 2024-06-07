<script setup lang="ts">
import { Ref, ref, onBeforeMount } from "vue";
import { MinIOFile } from "../../models/file";
import File from "../File.vue";
import { getTrashCan } from "../../utils/filesApi";
import { useEventEmitterStore } from "../../stores/eventEmitterStore";
import { FILE_UPDATED } from "../../constants";

const fileList: Ref<MinIOFile[]> = ref([] as MinIOFile[]);

const eventEmitter = useEventEmitterStore();

eventEmitter.eventEmitter.on(FILE_UPDATED, fetchTrashCan);

const isLoading = ref<boolean>(false);

onBeforeMount(fetchTrashCan);

async function fetchTrashCan(): Promise<void> {
  isLoading.value = true;
  const response = await getTrashCan();
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
