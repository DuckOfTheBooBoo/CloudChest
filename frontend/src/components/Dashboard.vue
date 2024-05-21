<script setup lang="ts">
import { ref } from "vue";

// import Item from "./Item.vue";
const uploadFileDialog = ref<boolean>(false);
const upFileDialogActivator = ref(undefined);

const newFolderDialog = ref<boolean>(false);
const newFolderDialogActivator = ref(undefined);

</script>

<template>
  <v-layout class="rounded rounded-md">
    <v-app-bar>
      <v-app-bar-title>Halo Arajdian Altaf!</v-app-bar-title>
    </v-app-bar>

    <v-navigation-drawer :width="200">
      <v-list-item>
        <v-menu>
          <template v-slot:activator="{ props }">
            <v-btn class="rounded-lg mt-2" variant="tonal" block v-bind="props">
              <v-icon>mdi-plus</v-icon>
              New
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="()=>{}" ref="upFileDialogActivator">
              <v-icon>mdi-upload</v-icon> Upload file
            </v-list-item>
            <v-divider></v-divider>
            <v-list-item @click="()=>{}" ref="newFolderDialogActivator">
              <v-icon>mdi-folder-plus</v-icon> New folder
            </v-list-item>
            <v-list-item @click="()=>{}">
              <v-icon>mdi-folder-upload</v-icon> Upload folder
            </v-list-item>
          </v-list>
        </v-menu>
      </v-list-item>
      <v-divider class="my-2"></v-divider>
      <v-list-item link>
        <v-icon class="mr-2">mdi-folder</v-icon>
        Files
      </v-list-item>
      <v-list-item link>
        <v-icon class="mr-2">mdi-star-outline</v-icon>
        Favorite
      </v-list-item>
      <v-list-item link>
        <v-icon class="mr-2">mdi-trash-can</v-icon>
        Trash
      </v-list-item>
    </v-navigation-drawer>

    <!-- UPLOAD FILE DIALOG -->
    <v-dialog v-model="uploadFileDialog" :activator="upFileDialogActivator" max-width="30rem" persistent>
      <template v-slot:default="{ isActive }">
        <v-card title="Upload file">
          <v-card-text>
            <v-file-input variant="outlined" accept="image/*" label="File input" counter show-size></v-file-input>
          </v-card-text>
          <v-card-actions>
            <v-btn @click="uploadFileDialog=false">Cancel</v-btn>
            <v-btn variant="tonal" color="blue" @click="uploadFileDialog=false">Upload</v-btn>
          </v-card-actions>
        </v-card>
      </template>
    </v-dialog>

    <!-- NEW FOLDER DIALOG -->
    <v-dialog v-model="newFolderDialog" :activator="newFolderDialogActivator" max-width="30rem" persistent>
      <template v-slot:default="{ isActive }">
        <v-card title="Create new folder">
          <v-card-text>
            <v-text-field label="Folder name" variant="outlined"></v-text-field>
          </v-card-text>
          <v-card-actions>
            <v-btn @click="newFolderDialog=false">Cancel</v-btn>
            <v-btn variant="tonal" color="blue" @click="newFolderDialog=false">Create</v-btn>
          </v-card-actions>
        </v-card>
      </template>
    </v-dialog>

    <v-main>
      <v-item-group selected-class="bg-primary">
        <v-container>
          <v-row>
            <v-col v-for="n in 10" :key="n" :cols="2">
              <v-item v-slot="{ isSelected, selectedClass, toggle }">
                <v-card
                  max-width="10rem"
                  :class="['pa-2 rounded-lg', selectedClass]"
                  elevation="5"
                  @click="toggle"
                >
                  <!-- Upper part (file name and menu) -->
                  <div
                    class="tw-flex tw-flex-row tw-h-full tw-mb-3 tw-w-full tw-items-center tw-flex-wrap"
                  >
                    <p class="text-body-2 tw-grow">
                      {{ isSelected ? "Selected" : "Not Selected" }}
                    </p>
                    <v-menu>
                      <template v-slot:activator="{ props }">
                        <v-btn
                          density="compact"
                          icon="mdi-dots-vertical"
                          variant="plain"
                          v-bind="props"
                        ></v-btn>
                      </template>
                      <v-list>
                        <v-list-item @click="console.log('Download')">
                          <v-icon>mdi-download</v-icon> Download
                        </v-list-item>
                        <v-list-item @click="console.log('Details')">
                          <v-icon>mdi-information-outline</v-icon> Details
                        </v-list-item>
                        <v-list-item @click="console.log('Mark as favorite')">
                          <v-icon>mdi-star-outline</v-icon> Mark as favorite
                        </v-list-item>
                        <v-list-item @click="console.log('Delete')">
                          <v-icon>mdi-trash-can</v-icon> Delete
                        </v-list-item>
                      </v-list>
                    </v-menu>
                  </div>

                  <div
                    class="tw-flex tw-justify-center tw-items-center tw-mb-2 tw-w-full tw-h-16 tw-rounded-lg bg-grey-darken-3"
                  >
                    <v-icon icon="mdi-trash-can"></v-icon>
                  </div>

                  <!-- Bottom part (date) -->
                  <p class="text-caption">1 day ago</p>
                </v-card>
              </v-item>
            </v-col>
          </v-row>
        </v-container>
      </v-item-group>
    </v-main>
  </v-layout>
</template>

<style scoped></style>
