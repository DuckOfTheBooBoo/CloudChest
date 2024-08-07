<script setup lang="ts">
import axios from "axios";
import { ref, mergeProps } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useEventEmitterStore } from "../stores/eventEmitterStore"
import { FILE_UPDATED } from "../constants"
import { createNewFolder } from "../utils/foldersApi";

const selectedNav = ref(0);
const router = useRouter();
const route = useRoute();
const file = ref<File | null>(null);
const newFolderName = ref<string | null>(null);

const eventEmitter = useEventEmitterStore();

const uploadFileDialog = ref<boolean>(false);
const upFileDialogActivator = ref(undefined);

const newFolderDialog = ref<boolean>(false);
const newFolderDialogActivator = ref(undefined);

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
  try {
    await axios.postForm('/api/files?multiple=false', {
      file: file.value as File,
      path: decodeURIComponent(route.query.path as string)
    });
    eventEmitter.eventEmitter.emit(FILE_UPDATED);
  } catch (error) {
    console.error(error);
  } finally {
    uploadFileDialog.value = false; // end
    file.value = null;
  }
}

async function newFolder(_: Event): Promise<void> {
  const folderCode: string = route.params.code ? route.params.code as string : '';
  
  await createNewFolder(folderCode, newFolderName.value as string);
  eventEmitter.eventEmitter.emit(FILE_UPDATED);
  newFolderDialog.value = false; // end
  newFolderName.value = null;
}

const rules = {
  required: (value: string) => !!value || 'Field is required',
};
console.log(route)
</script> 

<template>
  <v-layout class="rounded rounded-md">
    <v-app-bar>
      <v-menu>
        <template v-slot:activator="{ props: menu }">
          <v-tooltip location="top">
            <template v-slot:activator="{ props: tooltip }">
              <v-btn
                icon="mdi-account"
                v-bind="mergeProps(menu, tooltip)"
              ></v-btn>
            </template>
            <span>Account menu</span>
          </v-tooltip>
        </template>
        <v-list>
          <v-list-item prepend-icon="mdi-logout-variant" @click="logout"
            >Log out</v-list-item
          >
        </v-list>
      </v-menu>
      <v-app-bar-title>Halo Arajdian Altaf!</v-app-bar-title>
    </v-app-bar>

    <v-navigation-drawer :width="200">
      <v-list-item>
        <v-menu>
          <template v-slot:activator="{ props }">
            <v-btn class="rounded-lg mt-2" variant="tonal" block v-bind="props" :disabled="(route.name != 'explorer-files' && route.name != 'explorer-files-code')">
              <v-icon>mdi-plus</v-icon>
              New
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="() => {}" ref="upFileDialogActivator">
              <v-icon>mdi-upload</v-icon> Upload file
            </v-list-item>
            <v-divider></v-divider>
            <v-list-item @click="() => {}" ref="newFolderDialogActivator">
              <v-icon>mdi-folder-plus</v-icon> New folder
            </v-list-item>
            <v-list-item @click="() => {}">
              <!-- TODO: WIP -->
              <v-icon>mdi-folder-upload</v-icon> Upload folder
            </v-list-item>
          </v-list>
        </v-menu>
      </v-list-item>
      <v-divider class="my-2"></v-divider>
      <v-list v-model:selected="selectedNav">
        <v-list-item link :value="0" to="/explorer/files">
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
    <v-dialog
      v-model="uploadFileDialog"
      :activator="upFileDialogActivator"
      max-width="30rem"
      persistent
    >
      <template v-slot:default="{ isActive:_ }">
        <form @submit.prevent="uploadFile">
          <v-card title="Upload file">
            <v-card-text>
              <v-file-input
                variant="outlined"
                accept="*"
                label="File input"
                v-model="file"
                counter
                show-size
                name="file"
              ></v-file-input>
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
    <v-dialog
      v-model="newFolderDialog"
      :activator="newFolderDialogActivator"
      max-width="30rem"
      persistent
    >
      <template v-slot:default="{ isActive:_ }">
        <form @submit.prevent="newFolder">
          <v-card title="Create new folder">
            <v-card-text>
              <v-text-field label="Folder name" v-model="newFolderName" variant="outlined" :rules="[rules.required]"></v-text-field>
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
      <RouterView />
    </v-main>
  </v-layout>
</template>

<style scoped></style>
