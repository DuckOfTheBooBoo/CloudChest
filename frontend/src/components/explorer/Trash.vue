<script setup lang="ts">
import { Ref, ref, onBeforeMount } from "vue";
import { CloudChestFile } from "../../models/file";
import File from "../File.vue";
import { getTrashCan, emptyTrashCan } from "../../utils/filesApi";
import { getDeletedFolders } from "../../utils/foldersApi";
import { useEventEmitterStore } from "../../stores/eventEmitterStore";
import Folder from "../Folder.vue";
import type FolderModel from "../../models/folder";

const deletedFilesList: Ref<CloudChestFile[]> = ref([] as CloudChestFile[]);
const deletedFoldersList: Ref<FolderModel[]> = ref([] as FolderModel[]);

// const eventEmitter = useEventEmitterStore();
const folderCode = ref('root');
const isFoldersLoading = ref<boolean>(false);
const isFilesLoading = ref<boolean>(false);
const confirmDialogVisible: Ref<boolean> = ref<boolean>(false);

const evStore = useEventEmitterStore();

/**
 * If file is permanently deleted or restored, remove it from the list of deleted files and deleted folders in Trash.vue
 */
evStore.getEventEmitter.on("FOLDER_DELETED_PERM", (deletedObjects) => {
  deletedFoldersList.value = deletedFoldersList.value.filter((folder: FolderModel) => !deletedObjects.deleted_folders.includes(folder.Code));
  deletedFilesList.value = deletedFilesList.value.filter((file: CloudChestFile) => !deletedObjects.deleted_files.includes(file.FileCode));
});

// FOLDER_UPDATED only occurs on Trash.vue if folder is restored, thus remove it from the list
evStore.getEventEmitter.on("FOLDER_UPDATED", (updatedfolder: FolderModel) => {
  deletedFoldersList.value = deletedFoldersList.value.filter((folder: FolderModel) => folder.Code !== updatedfolder.Code);
})

const isLoading = ref<boolean>(false);

onBeforeMount(fetchTrashCan);

function fetchTrashCan(): void {
  fetchDeletedFiles();
  fetchDeletedFolders();
}

async function fetchDeletedFiles(): Promise<void> {
  isFilesLoading.value = true;
  const response = await getTrashCan();
  deletedFilesList.value = response.files;
  isFilesLoading.value = false;
}

async function fetchDeletedFolders(): Promise<void> {
  isFoldersLoading.value = true;
  const response = await getDeletedFolders();
  deletedFoldersList.value = response.folders;
  isFoldersLoading.value = false;
}

async function pruneAllFiles(): Promise<void> {
  await emptyTrashCan();
}

// If folder is permanently deleted or restored, remove it from the list
function handlePatchedFolder(patchedFolder: FolderModel) {
  deletedFoldersList.value = deletedFoldersList.value.filter((folder: FolderModel) => folder.Code !== patchedFolder.Code)
}

// If file is permanently deleted or restored, remove it from the list
function handlePatchedFile(patchedFile: CloudChestFile) {
  deletedFilesList.value = deletedFilesList.value.filter((file: CloudChestFile) => file.FileCode !== patchedFile.FileCode)
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
    <div>
      <h1 class="tw-mb-3 tw-text-3xl">Deleted Folders</h1>
      <div class="tw-min-h-1">
        <v-progress-linear v-if="isFoldersLoading" :indeterminate="true" color="primary"></v-progress-linear>
      </div>
      <v-item-group multiple>
        <v-container>
          <v-row>
            <v-col v-for="folder in deletedFoldersList" :key="folder" :cols="2">
              <v-item v-slot="{ isSelected, toggle }">
                <Folder :folder="folder" :parent-path="decodeURIComponent(folderCode)" :is-selected="isSelected" @click="toggle" @folder-state:update="handlePatchedFolder" />
              </v-item>
            </v-col>
          </v-row>
        </v-container>
      </v-item-group>
    </div>
    <div>
      <h1 class="tw-mb-3 tw-text-3xl">Deleted Files</h1>
      <div class="tw-min-h-1">
        <v-progress-linear v-if="isFilesLoading" :indeterminate="true" color="primary"></v-progress-linear>
      </div>
      <v-row>
        <v-col v-for="file in deletedFilesList" :key="file" :cols="2">
          <File :file="file" @file-state:update="handlePatchedFile" />
        </v-col>
      </v-row>
    </div>
  </v-container>
</template>

<style scoped></style>
