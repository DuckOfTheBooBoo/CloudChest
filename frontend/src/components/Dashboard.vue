<script setup lang="ts">
import axios from "axios";
import { ref, mergeProps, inject } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useEventEmitterStore } from "../stores/eventEmitterStore"
import { FILE_UPDATED } from "../constants"
import { createNewFolder } from "../utils/foldersApi";
import { CloudChestFile } from "../models/file";
import { downloadFile } from "../utils/filesApi";
import AxiosManager from "./AxiosManager.vue";
import {useAxiosManagerStore} from "../stores/axiosManagerStore";

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

const eventEmitter = useEventEmitterStore();

const uploadFileDialog = ref<boolean>(false);
const upFileDialogActivator = ref(undefined);

const newFolderDialog = ref<boolean>(false);
const newFolderDialogActivator = ref(undefined);

const axiosManager = useAxiosManagerStore();

async function logout(): Promise<void> {
  try {
    const response = await axios.post("/api/users/logout");

    if (response.status == 201) {
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
  const newFile: File = file.value as File;
  axiosManager.addUploadRequest(newFile, folderCode);
  file.value = null;
  uploadFileDialog.value = false;
}

async function newFolder(_: Event): Promise<void> {
  const folderCode: string = route.params.code ? route.params.code as string : 'root';

  await createNewFolder(folderCode, newFolderName.value as string);
  eventEmitter.eventEmitter.emit(FILE_UPDATED);
  newFolderDialog.value = false; // end
  newFolderName.value = null;
}

const rules = {
  required: (value: string) => !!value || 'Field is required',
};

async function getFileURL(): Promise<string> {
  const resp = await downloadFile(selectedFile.value!.ID)
  if (resp) {
    const downloadFileUrl: string = `${resp.Scheme}://${resp.Host}${resp.Path}?${resp.RawQuery}`;
    return downloadFileUrl
  }

  return '';
}

function handleFileChange(file: CloudChestFile): void {
  overlayVisible.value = true;
  selectedFile.value = file;
  previewable.value = file.FileType.includes('image/') || file.FileType.includes('video/') && file.IsPreviewable;

  if (selectedFile.value.FileType.includes('image/')) {
    getFileURL().then(url => fileURL.value = url)
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
</script>

<template>
  <v-layout class="rounded rounded-md tw-relative">
    <AxiosManager v-if="axiosManager.ongoingRequests.length > 0" class="tw-fixed tw-box-border tw-bottom-0 tw-right-0 tw-z-10" />
    <v-overlay v-model="overlayVisible" scroll-strategy="block">
      <v-toolbar class="tw-w-screen" density="comfortable">
        <v-toolbar-title>{{ selectedFile?.FileName }}</v-toolbar-title>

        <v-spacer></v-spacer>

        <v-btn @click="handlePreviewClose" icon>
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-toolbar>

      <div class="tw-py-6 tw-flex tw-justify-center tw-items-center tw-drop-shadow-xl">
        <v-img v-if="previewable && selectedFile?.FileType.includes('image/')" :src="fileURL"
          class="tw-h-[calc(100dvh-100px)]">
          <template v-slot:placeholder>
            <div class="d-flex align-center justify-center fill-height">
              <v-progress-circular color="grey-lighten-4" indeterminate></v-progress-circular>
            </div>
          </template>
        </v-img>
        <media-controller class="tw-h-[calc(100dvh-100px)]" v-else-if="previewable && selectedFile?.FileType.includes('video/')">
          <hls-video :src="`/api/hls/${selectedFile?.FileCode}/masterPlaylist`" slot="media"
            crossorigin muted></hls-video>
          <media-loading-indicator slot="centered-chrome" noautohide></media-loading-indicator>
          <media-control-bar>
            <media-play-button></media-play-button>
            <media-seek-backward-button></media-seek-backward-button>
            <media-seek-forward-button></media-seek-forward-button>
            <media-mute-button></media-mute-button>
            <media-volume-range></media-volume-range>
            <media-time-range></media-time-range>
            <media-time-display showduration remaining></media-time-display>
            <media-playback-rate-button></media-playback-rate-button>
            <media-fullscreen-button></media-fullscreen-button>
          </media-control-bar>
        </media-controller>
        <p v-else class="tw-text-2xl">This file is does not have a preview.</p>
      </div>

    </v-overlay>

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
