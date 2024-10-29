<script setup lang="ts">
// Libraries
import axios from "axios";

// Vue & Vue Router
import { ref, mergeProps, provide } from "vue";
import { useRouter, useRoute } from "vue-router";
// Pinia
import { useEventEmitterStore } from "../stores/eventEmitterStore"
import {useAxiosManagerStore} from "../stores/axiosManagerStore";

// Local imports
import { FILE_UPDATED } from "../constants"
import isFolder from "../utils/isFolder";

// API
import { createNewFolder, patchFolder } from "../utils/foldersApi";
import { patchFile } from "../utils/filesApi";

// Models & type
import { CloudChestFile } from "../models/file";
import type Folder from "../models/folder";

// Components
import AxiosManager from "./AxiosManager.vue";
import FolderListNavigator from "./FolderListNavigator.vue";
import Previewer from "./Previewer.vue";
import { FilePatchRequest, FolderPatchRequest } from "../models/requestModel";

const selectedNav = ref(0);
const router = useRouter();
const route = useRoute();
const file = ref<File | null>(null);
const newFolderName = ref<string | null>(null);
const overlayVisible = ref<boolean>(false);
const selectedFile = ref<CloudChestFile | null>(null);
const fileURL = ref<string | null>(null);
const previewable = ref<boolean>(false);
const lastFolderCode = ref<string>("root");
const fileListNavDialog = ref<boolean>(false);
const moveObjectPlaceholder = ref<CloudChestFile | Folder | null>(null);
const blacklistedFolder = ref<Folder | undefined>();

const eventEmitter = useEventEmitterStore();

const uploadFileDialog = ref<boolean>(false);
const upFileDialogActivator = ref(undefined);

const newFolderDialog = ref<boolean>(false);
const newFolderDialogActivator = ref(undefined);

const axiosManager = useAxiosManagerStore();

async function logout(): Promise<void> {
  try {
    const response = await axios.post("/api/auth/logout");

    if (response.status == 200) {
      localStorage.removeItem("token");
      axios.defaults.headers.common["Authorization"] = "";

      router.push("/login");
    }
  } catch (error) {
    console.error(error);

    // Token might be invalid already
    router.push("/login");
  }
}

async function uploadFile(_: Event): Promise<void> {
  const folderCode: string = route.params.code ? route.params.code as string : 'root';
  if (file.value) {
    axiosManager.addUploadRequest(file.value, folderCode);
    file.value = null;
  }
  uploadFileDialog.value = false;
}

async function newFolder(_: Event): Promise<void> {
  const folderCode: string = route.params.code ? route.params.code as string : 'root';
  await createNewFolder(folderCode, newFolderName.value as string);
  newFolderDialog.value = false; // end
  newFolderName.value = null;
}

const rules = {
  required: (value: string) => !!value || 'Field is required',
};


function handleFileChange(file: CloudChestFile): void {
  overlayVisible.value = true;
  selectedFile.value = file;
  previewable.value = file.FileType.includes('image/') || file.FileType.includes('video/') && file.IsPreviewable;

  if (selectedFile.value.FileType.includes('image/')) {
    fileURL.value = `/api/files/${selectedFile.value.FileCode}/download`;
  }
}

function handlePreviewClose(): void {
  overlayVisible.value = false;
  selectedFile.value = null;
  fileURL.value = null;
  previewable.value = false;
}

function handleFolderSelect(newFolderCode: string): void {
  lastFolderCode.value = newFolderCode;
}

async function handleFileNavigatorSelect(folder: Folder | null): Promise<void> {
  console.log(folder);
  if (folder && moveObjectPlaceholder.value) {
    fileListNavDialog.value = false;

    // variable contains CloudChestFile
    if (moveObjectPlaceholder.value instanceof CloudChestFile) {
      const moveRequest: FilePatchRequest = {
        folder_code: folder.Code,
      }
  
      await patchFile(moveObjectPlaceholder.value, moveRequest);
    } else if (isFolder(moveObjectPlaceholder.value)) {
      // variable contains object that implements Folder interface
      const moveRequest: FolderPatchRequest = {
        parent_folder_code: folder.Code,
      }
  
      await patchFolder(blacklistedFolder.value!.Code, moveRequest);
    }
  }

  moveObjectPlaceholder.value = null;
  blacklistedFolder.value = undefined;
}

function handleFileNavigatorCancel(): void {
  fileListNavDialog.value = false;
  moveObjectPlaceholder.value = null;
  blacklistedFolder.value = undefined;
}

const showFileNavigatorDialog = (file: CloudChestFile | Folder): void => {
  fileListNavDialog.value = true;
  moveObjectPlaceholder.value = file;

  if (isFolder(file)) {
    blacklistedFolder.value = file;
  }
}

provide('showFileNavigatorDialog', showFileNavigatorDialog);
</script>

<template>
  <!-- File Navigator Dialog -->
  <v-dialog
    v-model="fileListNavDialog"
    scrollable 
    persistent
    max-width="500px"
    max-height="90%"
    transition="dialog-transition"
  >
    <FolderListNavigator :blacklistedFolder="blacklistedFolder" @nav:cancel="handleFileNavigatorCancel" @nav:move="handleFileNavigatorSelect" />
  </v-dialog>

  <v-layout class="rounded rounded-md tw-relative">
    <AxiosManager v-if="axiosManager.ongoingRequests.length > 0" class="tw-fixed tw-box-border tw-bottom-0 tw-right-0 tw-z-10" />
    <Previewer :visible="overlayVisible" :file="selectedFile" @on:close="handlePreviewClose" />

    <v-app-bar>
      <v-menu>
        <template v-slot:activator="{ props: menu }">
          <v-tooltip location="top">
            <template v-slot:activator="{ props: tooltip }">
              <v-btn icon="mdi-account" v-bind="mergeProps(menu, tooltip)"></v-btn>
            </template>
            <span>Account menu</span>
          </v-tooltip>
        </template>
        <v-list>
          <v-list-item prepend-icon="mdi-logout-variant" @click="logout">Log out</v-list-item>
        </v-list>
      </v-menu>
      <v-app-bar-title>Halo Arajdian Altaf!</v-app-bar-title>
    </v-app-bar>

    <v-navigation-drawer :width="200">
      <v-list-item>
        <v-menu>
          <template v-slot:activator="{ props }">
            <v-btn class="rounded-lg mt-2" variant="tonal" block v-bind="props"
              :disabled="(route.name != 'explorer-files' && route.name != 'explorer-files-code')">
              <v-icon>mdi-plus</v-icon>
              New
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="() => { }" ref="upFileDialogActivator">
              <v-icon>mdi-upload</v-icon> Upload file
            </v-list-item>
            <v-divider></v-divider>
            <v-list-item @click="() => { }" ref="newFolderDialogActivator">
              <v-icon>mdi-folder-plus</v-icon> New folder
            </v-list-item>
            <v-list-item @click="() => { }">
              <!-- TODO: WIP -->
              <v-icon>mdi-folder-upload</v-icon> Upload folder
            </v-list-item>
          </v-list>
        </v-menu>
      </v-list-item>
      <v-divider class="my-2"></v-divider>
      <v-list v-model:selected="selectedNav">
        <v-list-item v-if="lastFolderCode === 'root'" link :value="0" to="/explorer/files">
          <v-icon class="mr-2">mdi-folder</v-icon>
          Files
        </v-list-item>
        <v-list-item v-else link :value="0" :to="`/explorer/files/${lastFolderCode}`">
          <v-icon class="mr-2">mdi-folder</v-icon>
          Files
        </v-list-item>
        <v-list-item link :value="1" to="/explorer/favorite">
          <v-icon class="mr-2">mdi-star-outline</v-icon>
          Favorite
        </v-list-item>
        <v-list-item link :value="2" to="/explorer/trash">
          <v-icon class="mr-2">mdi-trash-can</v-icon>
          Trash
        </v-list-item>
      </v-list>
    </v-navigation-drawer>

    <!-- UPLOAD FILE DIALOG -->
    <v-dialog v-model="uploadFileDialog" :activator="upFileDialogActivator" max-width="30rem" persistent>
      <template v-slot:default="{ isActive: _ }">
        <form @submit.prevent="uploadFile">
          <v-card title="Upload file">
            <v-card-text>
              <v-file-input variant="outlined" accept="*" label="File input" v-model="file" counter show-size
                name="file"></v-file-input>
            </v-card-text>
            <v-card-actions>
              <v-btn @click="uploadFileDialog = false">Cancel</v-btn>
              <v-btn variant="tonal" color="blue" type="submit">Upload</v-btn>
            </v-card-actions>
          </v-card>
        </form>
      </template>
    </v-dialog>

    <!-- NEW FOLDER DIALOG -->
    <v-dialog v-model="newFolderDialog" :activator="newFolderDialogActivator" max-width="30rem" persistent>
      <template v-slot:default="{ isActive: _ }">
        <form @submit.prevent="newFolder">
          <v-card title="Create new folder">
            <v-card-text>
              <v-text-field label="Folder name" v-model="newFolderName" variant="outlined"
                :rules="[rules.required]"></v-text-field>
            </v-card-text>
            <v-card-actions>
              <v-btn @click="newFolderDialog = false">Cancel</v-btn>
              <v-btn variant="tonal" color="blue" @click="newFolderDialog = false" type="submit">Create</v-btn>
            </v-card-actions>
          </v-card>
        </form>
      </template>
    </v-dialog>

    <v-main>
      <RouterView v-slot="{ Component }">
        <component :is="Component" @file:select="handleFileChange" @folder:select="handleFolderSelect" />
      </RouterView>
    </v-main>
  </v-layout>
</template>

<style scoped></style>
