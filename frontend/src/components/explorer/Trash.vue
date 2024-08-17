<script setup lang="ts">
import { Ref, ref, onBeforeMount } from "vue";
import { CloudChestFile } from "../../models/file";
import File from "../File.vue";
import { getTrashCan, emptyTrashCan } from "../../utils/filesApi";
import { useEventEmitterStore } from "../../stores/eventEmitterStore";
import { FILE_UPDATED } from "../../constants";

const fileList: Ref<CloudChestFile[]> = ref([] as CloudChestFile[]);
const confirmDialogVisible: Ref<boolean> = ref<boolean>(false);

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

async function pruneAllFiles(): Promise<void> {
  await emptyTrashCan();
}
</script>

<template>
  <div class="tw-min-h-1">
    <v-progress-linear v-if="isLoading" :indeterminate="true" color="primary"></v-progress-linear>
  </div>
  <v-dialog v-model="confirmDialogVisible" max-width="540">
    <template v-slot:default="{ isActive }">
      <v-card prepend-icon="mdi-alert" text="Pruning will delete the file permanently, are you sure?"
        title="Confirmation Dialog">
        <template v-slot:actions>
          <v-btn class="" @click="isActive.value = false">Cancel</v-btn>
          <v-btn class="text-red" variant="outlined" @click="() => {
            isActive.value = false
            pruneAllFiles()
          }">Delete</v-btn>
        </template>
      </v-card>
    </template>
  </v-dialog>
  <v-container>
    <div
      class="tw-w-full tw-h-10 tw-flex tw-justify-between tw-items-center tw-px-4 tw-mb-2 bg-grey-lighten-2 rounded-lg">
      <span class="tw-text-left tw-font-semibold">All files in trash will be deleted permanently in 30 days.</span>
      <v-tooltip text="Delete all files permanently" location="bottom">
        <template v-slot:activator="{ props: tltpProps }">
          <v-btn icon="mdi-trash-can-outline" color="error" density="comfortable" variant="text" v-bind="tltpProps"
            @click="confirmDialogVisible = true"></v-btn>
        </template>
      </v-tooltip>
    </div>
    <v-row>
      <v-col v-for="file in fileList" :key="file" :cols="2">
        <File :file="file" />
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped></style>
