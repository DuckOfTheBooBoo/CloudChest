<script setup lang="ts">
import { ref } from "vue";
import Folder from "../models/folder";
import Filename from "./Filename.vue";
import { patchFolder } from "../utils/foldersApi";
import { type FolderPatchRequest } from "../models/requestModel";

const menuVisible = ref<boolean>(false);
const isHover = ref<boolean>(false);
const renameFolderPlaceholder = ref<string | undefined>();
const folderDetailDialog = ref<boolean>(false);

const props = defineProps<{
  folder: Folder;
  isSelected: boolean
}>();

const emit = defineEmits<{
  (e: "folderCode:change", newFolderCode: string): void
}>();

const rules = {
  required: (value: string) => !!value || 'Folder name cannot be empty',
};

async function renameFolder(): Promise<void> {
  const request: FolderPatchRequest = {
    folder_name: renameFolderPlaceholder.value
  }
  await patchFolder(props.folder.Code, request)
}
</script>

<template>
  <v-tooltip :text="folder.Name" location="bottom">
    <template v-slot:activator="{ props: tltpProps }">
      <div
        class="tw-flex tw-flex-col tw-items-center tw-max-w-15rem hover:tw-cursor-pointer hover:tw-bg-[#424242] tw-rounded-xl tw-pb-2 tw-transition-[background-color]"
        @dblclick="emit('folderCode:change', folder.Code)" @mouseover="isHover = true" @mouseleave="isHover = false"
        :class="{ 'tw-bg-[#424242]': isSelected }" v-bind="tltpProps">
        <v-icon class="tw-text-9xl !important">mdi-folder</v-icon>
        <div class="tw-w-full tw-flex tw-flex-row tw-justify-around">
          <Filename :filename="folder.Name" />
          <v-menu location="bottom" :attach="true" close-delay="0" :no-click-animation="true">
            <template v-slot:activator="{ props }">
              <v-btn density="compact" icon="mdi-dots-vertical" variant="text"
                size="small" v-bind="props"></v-btn>
            </template>
            <v-list>

              <!-- RENAME FILE -->
              <v-list-item @click="() => {
                renameFolderPlaceholder = folder.Name }">
                <v-icon>mdi-pencil</v-icon> Rename

                <v-dialog activator="parent" max-width="500px">
                  <template v-slot:default="{ isActive }">
                    <v-card>
                      <v-card-title title>
                        Rename {{ folder.Name }}
                      </v-card-title>

                      <v-card-item>
                        <v-text-field v-model="renameFolderPlaceholder" label="New filename" single-line
                          :rules="[rules.required]"></v-text-field>
                      </v-card-item>

                      <v-card-actions>
                        <v-btn @click="isActive.value = false">Cancel</v-btn>
                        <v-btn variant="outlined" @click="() => {
                          isActive.value = false
                          renameFolder()
                        }">Rename</v-btn>
                      </v-card-actions>
                    </v-card>
                  </template>
                </v-dialog>

              </v-list-item>

              <!-- DOWNLOAD -->
              <v-list-item @click="() => {}">
                <v-icon>mdi-download</v-icon> Download
              </v-list-item>

              <!-- MOVE FILE -->
              <v-list-item @click="() => {}">
                <v-icon>mdi-folder-arrow-right</v-icon> <span class="tw-ml-1">Move to</span>
              </v-list-item>

              <!-- FILE DETAILS DIALOG -->
              <v-list-item @click="() => { }">
                <!-- DETAILS DIALOG -->
                <v-dialog activator="parent" max-width="30rem" v-model="folderDetailDialog">
                  <template v-slot:default="{ isActive: _ }">
                    <v-card>
                      <v-card-title>
                        <div class="tw-flex tw-flex-row tw-justify-between tw-items-center tw-px-2">
                          <p>File details</p>
                          <v-btn icon="mdi-close" variant="flat" @click="folderDetailDialog = false"></v-btn>
                        </div>
                      </v-card-title>
                      <v-card-text>
                      </v-card-text>
                    </v-card>
                  </template>
                </v-dialog>

                <v-icon>mdi-information-outline</v-icon> Details
              </v-list-item>

              <!-- TOGGLE FAVORITE FILES -->
              <v-list-item @click="() => {}" v-if="!folder.DeletedAt">
              </v-list-item>

              <!-- DELETE FILE -->
              <v-list-item v-if="!folder.DeletedAt" @click="() => {}">
                <v-icon>mdi-trash-can</v-icon> Delete
              </v-list-item>

              <!-- RESTORE DELETED FILE -->
              <v-list-item v-else @click="() => {}">
                <v-icon>mdi-delete-restore</v-icon> Restore
              </v-list-item>

              <!-- PERMANENTLY DELETE FILE -->
              <v-list-item v-if="folder.DeletedAt" @click="() => { }">
                <v-icon>mdi-delete-forever</v-icon> Prune
                <v-dialog activator="parent" max-width="340">
                  <template v-slot:default="{ isActive }">
                    <v-card prepend-icon="mdi-alert" text="Pruning will delete the file permanently, are you sure?"
                      title="Confirmation Dialog">
                      <template v-slot:actions>
                        <v-btn class="" @click="isActive.value = false">Cancel</v-btn>
                        <v-btn class="text-red" variant="outlined" @click="() => {
                          isActive.value = false
                        }">Delete</v-btn>
                      </template>
                    </v-card>
                  </template>
                </v-dialog>
              </v-list-item>
            </v-list>
          </v-menu>
        </div>
      </div>
    </template>
  </v-tooltip>
</template>

<script setup></script>
