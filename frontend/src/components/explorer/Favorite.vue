<script setup lang="ts">
import { useRoute, useRouter } from "vue-router";
import { watch, Ref, ref, onMounted } from "vue";
import { type CloudChestFile } from "../../models/file";
import { getFavoriteFiles } from "../../utils/filesApi";
import { getFavoriteFolders } from "../../utils/foldersApi";
import File from "../File.vue";
import Folder from "../Folder.vue";
import FolderModel from "../../models/folder";
// import { useEventEmitterStore } from "../../stores/eventEmitterStore";
// import { FILE_UPDATED, FOLDER_UPDATED } from "../../constants";

const emit = defineEmits<{
  (e: "file:select", file: CloudChestFile): void,
  (e: "folder:select", folderCode: string): void
}>();

const favoriteFilesList: Ref<CloudChestFile[]> = ref([] as CloudChestFile[]);
const favoriteFoldersList: Ref<FolderModel[]> = ref([] as FolderModel[]);

// const eventEmitter = useEventEmitterStore();
const route = useRoute();
const router = useRouter();
const folderCode = ref('root');
const isFoldersLoading = ref<boolean>(false);
const isFilesLoading = ref<boolean>(false);

// eventEmitter.eventEmitter.on(FILE_UPDATED, () => {
//   fetchFavoriteFiles()
// })

// eventEmitter.eventEmitter.on(FOLDER_UPDATED, () => {
//   fetchFavoriteFolders()
// })

watch(() => route.params.code, async () => {
  folderCode.value = route.params.code ? route.params.code as string : 'root';
  fetchFavoriteFiles();
  fetchFavoriteFolders();
}, { immediate: true })

onMounted(async () => {
  folderCode.value = route.params.code ? route.params.code as string : 'root';
  fetchFavoriteFiles();
  fetchFavoriteFolders();
})

async function fetchFavoriteFolders(): Promise<void> {
  isFoldersLoading.value = true;
  const response = await getFavoriteFolders();
  favoriteFoldersList.value = response.folders;
  isFoldersLoading.value = false;
}

async function fetchFavoriteFiles(): Promise<void> {
  isFilesLoading.value = true;
  const resp = await getFavoriteFiles();
  favoriteFilesList.value = resp.files;
  isFilesLoading.value = false;
}

function handleFolderCodeChange(newFolderCode: string) {
  emit('folder:select', newFolderCode)
  router.push({ name: 'explorer-files-code', params: { code: newFolderCode } })
}

function handlePatchedFolder(patchedFolder: FolderModel) {
  const index: number = favoriteFoldersList.value.findIndex((folder: FolderModel) => folder.Code === patchedFolder.Code);
  favoriteFoldersList.value.splice(index, 1, patchedFolder)
}

function handlePatchedFile(patchedFile: CloudChestFile) {
  const index: number = favoriteFilesList.value.findIndex((file: CloudChestFile) => file.FileCode === patchedFile.FileCode);
  favoriteFilesList.value.splice(index, 1, patchedFile)
}
</script>

<template>
  <v-container class="tw-flex tw-flex-col tw-gap-6">
    <div>
      <h1 class="tw-mb-3 tw-text-3xl">Favorite Folders</h1>
      <div class="tw-min-h-1">
        <v-progress-linear v-if="isFoldersLoading" :indeterminate="true" color="primary"></v-progress-linear>
      </div>
      <v-item-group multiple>
        <v-container>
          <v-row>
            <v-col v-for="folder in favoriteFoldersList" :key="folder" :cols="2">
              <v-item v-slot="{ isSelected, toggle }">
                <Folder :folder="folder" :parent-path="decodeURIComponent(folderCode)"
                  @folder-code:change="handleFolderCodeChange" :is-selected="isSelected" @click="toggle" @folder-state:update="handlePatchedFolder" />
              </v-item>
            </v-col>
          </v-row>
        </v-container>
      </v-item-group>
    </div>
    <div>
      <h1 class="tw-mb-3 tw-text-3xl">Favorite Files</h1>
      <div class="tw-min-h-1">
        <v-progress-linear v-if="isFilesLoading" :indeterminate="true" color="primary"></v-progress-linear>
      </div>
      <v-row>
        <v-col v-for="file in favoriteFilesList" :key="file" :cols="2">
          <File :file="file" @dblclick="emit('file:select', file)" @file-state:update="handlePatchedFile" />
        </v-col>
      </v-row>
    </div>
  </v-container>
</template>

<style scoped></style>
