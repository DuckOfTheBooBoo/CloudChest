<script setup lang="ts">
import { useRoute, useRouter } from "vue-router";
import { watch, Ref, ref, onBeforeMount, onBeforeUnmount, onMounted } from "vue";
import { type CloudChestFile } from "../../models/file";
import { getFilesFromCode } from "../../utils/filesApi";
import { getFolderList } from "../../utils/foldersApi";
import File from "../File.vue";
import Folder from "../Folder.vue";
import FolderModel from "../../models/folder";
import { useEventEmitterStore } from "../../stores/eventEmitterStore";
import { FILE_UPDATED, FOLDER_UPDATED } from "../../constants";
import type FolderHierarchy from "../../models/folderHierarchy";

const fileList: Ref<CloudChestFile[]> = ref([] as CloudChestFile[]);
const folderList: Ref<FolderModel[]> = ref([] as FolderModel[]);
const folderHierarchies: Ref<FolderHierarchy[]> = ref([] as FolderHierarchy[]);

const eventEmitter = useEventEmitterStore();
const route = useRoute();
const router = useRouter();
const folderCode = ref('');
const isFoldersLoading = ref<boolean>(false);
const isFilesLoading = ref<boolean>(false);

eventEmitter.eventEmitter.on(FILE_UPDATED, () => {
  fetchFiles(folderCode.value)
})

eventEmitter.eventEmitter.on(FOLDER_UPDATED, () => {
  fetchFolders(folderCode.value)
})

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

async function fetchFolders(folderCode: string): Promise<void> {
  isFoldersLoading.value = true;
  const response = await getFolderList(folderCode);
  folderList.value = response.folders;
  folderHierarchies.value = response.hierarchies;
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
  <v-container class="tw-flex tw-flex-col tw-gap-6">
    <nav>
      <span>
        <v-btn variant="text" rounded="xl" class="text-h6" @click="router.push({ name: 'explorer-files' })">Home</v-btn>
        <v-icon>mdi-menu-right</v-icon>
      </span>
      <span v-for="hierarchy in folderHierarchies" :key="hierarchy.code">
        <span v-if="hierarchy.name !== '/'">
          <v-btn variant="text" rounded="xl" class="text-h6" @click="router.push({ name: 'explorer-files-code', params: { code: hierarchy.code } })">{{ hierarchy.name }}</v-btn>
          <v-icon>mdi-menu-right</v-icon>
        </span>
      </span>
      
    </nav>
    <div>
      <h1 class="tw-mb-3 tw-text-3xl">Folders</h1>
      <div class="tw-min-h-1">
        <v-progress-linear v-if="isFoldersLoading" :indeterminate="true" color="primary"></v-progress-linear>
      </div>
      <v-item-group multiple>
        <v-container>
          <v-row>
            <v-col v-for="folder in folderList" :key="folder" :cols="2">
              <v-item v-slot="{ isSelected, toggle }">
                <Folder :folder="folder" :parent-path="decodeURIComponent(folderCode)"
                  @folder-code:change="handleFolderCodeChange" :is-selected="isSelected" @click="toggle" />
              </v-item>
            </v-col>
          </v-row>
        </v-container>
      </v-item-group>
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
