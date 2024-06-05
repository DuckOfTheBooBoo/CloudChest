<script setup lang="ts">
import { useRoute, useRouter } from "vue-router";
import { watch, Ref, ref, onBeforeMount, onBeforeUnmount, onMounted } from "vue";
import { MinIOFile } from "../../models/file";
import { getFilesFromPath } from "../../utils/filesApi";
import File from "../File.vue";
import Folder from "../Folder.vue";
import FolderModel from "../../models/folder";
import { useEventEmitterStore } from "../../stores/eventEmitterStore";
import { FILE_UPDATED } from "../../constants";

const fileList: Ref<MinIOFile[]> = ref([] as MinIOFile[]);
const folderList: Ref<FolderModel[]> = ref([] as FolderModel[]);

const eventEmitter = useEventEmitterStore();
const route = useRoute();
const router = useRouter();
const path = ref('/');
const isLoading = ref<boolean>(false);

eventEmitter.eventEmitter.on(FILE_UPDATED, () => {
  fetchFiles(path.value)
})

// Handle back and forward navigation by watching route changes
watch(route, (newRoute, _) => {
  const newDecodedPath = decodeURIComponent(newRoute.query.path as string);
  fetchFiles(newDecodedPath);
})

onBeforeUnmount(() => {
  window.removeEventListener('popstate', () => {});
});

onBeforeMount(async () => {
  path.value = decodeURIComponent(route.query.path as string);
  await fetchFiles(path.value);
});

onMounted(() => {
  console.log(route.fullPath)
})

async function makeRequest(pathParam: string): Promise<void> {
  path.value = pathParam
  await fetchFiles(path.value)
  router.push({ path: '/explorer/files', query: { path: encodeURIComponent(path.value) }})
}

async function fetchFiles(pathParam: string): Promise<void> {
  isLoading.value = true;
  const response = await getFilesFromPath(pathParam);
  fileList.value = response.files;
  folderList.value = response.folders;
  isLoading.value = false;
}
</script>

<template>
  <v-progress-linear v-if="isLoading" :indeterminate="true" color="primary"></v-progress-linear>
  <v-container>
    <v-row>
      <v-col v-for="folder in folderList" :key="folder" :cols="2">
        <Folder :folder="folder" :parent-path="decodeURIComponent(path)" :make-request="makeRequest"/>
      </v-col>
      <v-col v-for="file in fileList" :key="file" :cols="2">
        <File :file="file" />
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped></style>
