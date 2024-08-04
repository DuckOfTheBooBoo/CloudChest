<script setup lang="ts">
import { useRoute, useRouter } from "vue-router";
import { watch, Ref, ref, onBeforeMount, onBeforeUnmount, onMounted } from "vue";
import { CloudChestFile } from "../../models/file";
import { getFilesFromCode } from "../../utils/filesApi";
import { getFolderList } from "../../utils/foldersApi";
import File from "../File.vue";
import Folder from "../Folder.vue";
import FolderModel from "../../models/folder";
import { useEventEmitterStore } from "../../stores/eventEmitterStore";
import { FILE_UPDATED } from "../../constants";

const fileList: Ref<CloudChestFile[]> = ref([] as CloudChestFile[]);
const folderList: Ref<FolderModel[]> = ref([] as FolderModel[]);

const eventEmitter = useEventEmitterStore();
const route = useRoute();
const router = useRouter();
const folderCode = ref('');
const isFoldersLoading = ref<boolean>(false);
const isFilesLoading = ref<boolean>(false);

eventEmitter.eventEmitter.on(FILE_UPDATED, () => {
  fetchFiles(folderCode.value)
})

// Handle back and forward navigation by watching route changes
// watch(route, (newRoute, _) => {
//   const newDecodedPath = decodeURIComponent(newRoute.query.folderCode as string);
//   folderCode.value = newDecodedPath
//   fetchFiles(newDecodedPath);
// })

// onBeforeUnmount(() => {
//   window.removeEventListener('popstate', () => {});
// });

// onBeforeMount(async () => {
//   console.log(route)
//   folderCode.value = decodeURIComponent(route.query.folderCode as string);
//   await fetchFiles(folderCode.value);
// });

watch(() => route.params.code, async () => {
  folderCode.value = route.params.code ? route.params.code as string : '';
  fetchFiles(folderCode.value);
  fetchFolders(folderCode.value);
}, { immediate: true })


onMounted(async () => {
  folderCode.value = route.params.code ? route.params.code as string : '';
  fetchFiles(folderCode.value);
  fetchFolders(folderCode.value);
})

// async function makeRequest(pathParam: string): Promise<void> {
//   folderCode.value = pathParam
//   await fetchFiles(folderCode.value)
//   router.push({ path: '/explorer/files', query: { path: encodeURIComponent(folderCode.value) }})
// }

async function fetchFolders(folderCode: string): Promise<void> {
  isFoldersLoading.value = true;
  folderList.value = await getFolderList(folderCode);
  isFoldersLoading.value = false;
}

async function fetchFiles(folderCode: string): Promise<void> {
  isFilesLoading.value = true;
  fileList.value = await getFilesFromCode(folderCode);
  isFilesLoading.value = false;
}

function handleFolderCodeChange(newFolderCode: string) {
  router.push({ name: 'explorer-files-code', params: { code: newFolderCode } })
}
</script>

<template>
  <!-- <div class="tw-min-h-1">
    <v-progress-linear v-if="isLoading" :indeterminate="true" color="primary"></v-progress-linear>
  </div> -->
  <v-container class="tw-flex tw-flex-col tw-gap-6">
    <div>
      <h1 class="tw-mb-3 tw-text-3xl">Folders</h1>
      <div class="tw-min-h-1">
        <v-progress-linear v-if="isFoldersLoading" :indeterminate="true" color="primary"></v-progress-linear>
      </div>
      <v-row>
        <v-col v-for="folder in folderList" :key="folder" :cols="2">
          <Folder :folder="folder" :parent-path="decodeURIComponent(folderCode)"
            @folder-code:change="handleFolderCodeChange" />
        </v-col>
      </v-row>
    </div>
    <div>
      <h1 class="tw-mb-3 tw-text-3xl">Files</h1>
      <div class="tw-min-h-1">
        <v-progress-linear v-if="isFilesLoading" :indeterminate="true" color="primary"></v-progress-linear>
      </div>
      <v-row>
        <v-col v-for="file in fileList" :key="file" :cols="2">
          <File :file="file" />
        </v-col>
      </v-row>
    </div>
  </v-container>
</template>

<style scoped></style>
