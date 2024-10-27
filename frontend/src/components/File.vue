<script setup lang="ts">
import { ref, inject, computed } from "vue";
import { formatDistance } from "date-fns";
import { fileDetailFormatter } from "../utils/fileDetailFormatter";
import Filename from "./Filename.vue";
import { type CloudChestFile } from "../models/file";
import { trashFile, updateFile, patchFile } from "../utils/filesApi";
import { type FilePatchRequest } from "../models/requestModel";
import {FileTypeCategorizer, type FileCategory} from "../utils/fileTypeCategorizer";

const props = defineProps<{
  file: CloudChestFile
}>();

const emit = defineEmits<{
  (e: "dblclick"): void;
  (e: "file-state:update", newFile: CloudChestFile): void
}>();

const file = props.file;
const fileDetailDialog = ref(false);
const renameFilePlaceholder = ref<string | undefined>(undefined);

const showFileNavigatorDialog: ((file: CloudChestFile) => void) | undefined = inject('showFileNavigatorDialog')

const categoryToMDIIcon: {[categoryName: string]: string} = {
  'document': 'mdi-file-document-outline',
  'plaintext': 'mdi-file-document-outline',
  'audio': 'mdi-music-note',
  'font': 'mdi-format-font',
  'archive': 'mdi-bookshelf',
  'other': 'mdi-file-outline',
}

const fileCategory = computed(() => FileTypeCategorizer.categorizeFile(file.FileType, file.FileName));

async function toggleFavorite(): Promise<void> {
  const fileCopy: CloudChestFile = file;
  fileCopy.IsFavorite = !file.IsFavorite;
  const isSuccessful: boolean = await updateFile(fileCopy, false);
  if (isSuccessful) {
    file.IsFavorite = !file.IsFavorite
    emit('file-state:update', file)
  }
}

async function restoreFile(): Promise<void> {
  const request: FilePatchRequest = {
    is_restore: true
  }
  await patchFile(file, request);
}

async function pruneFile(): Promise<void> {
  try {
    await trashFile(file, false);
  } catch (error) {
    console.error(error);
  }
}

async function renameFile(): Promise<void> {
  const newFilename: string = renameFilePlaceholder.value ? renameFilePlaceholder.value : file.FileName;
  const requestBody: FilePatchRequest = {
    file_name: newFilename,
  }
  await patchFile(props.file, requestBody)
  renameFilePlaceholder.value = undefined
}

async function getFileURL(): Promise<void> {
  const url: string = `/api/files/${file?.FileCode}/download`;
  window.open(url, '_blank');
}

function moveFile(): void {
  showFileNavigatorDialog?.(file);
}

const rules = {
  required: (value: string) => !!value || 'Filename cannot be empty',
};

const thumbnailURL = computed(() => {
  let url: string = `/api/files/${file.FileCode}/thumbnail`;
  if (file.DeletedAt) {
    url += '?deleted=true'
  }
  return url;
})

</script>

<template>
  <v-tooltip :text="file.FileName" location="bottom">
    <template v-slot:activator="{ props: tltpProps }">
      <v-card class="mx-auto" max-width="374" @dblclick="emit('dblclick')" @click="" v-bind="tltpProps">
        <template v-slot:loader="{ isActive }">
          <v-progress-linear :active="isActive" color="deep-purple" height="4" indeterminate></v-progress-linear>
        </template>

        <div class="tw-h-[100px] tw-flex tw-items-center tw-justify-center tw-text-4xl" v-if="!file.FileType.includes('image/') && !file.FileType.includes('video/')">
          <v-icon>{{ categoryToMDIIcon[fileCategory] }}</v-icon>
        </div>
        <v-img v-else height="100" :src="thumbnailURL" cover alt="No thumbnail"></v-img>

        <v-card-item>
          <div class="tw-flex tw-flex-row tw-h-full tw-w-full tw-items-center tw-justify-between">
          <Filename :filename="file.FileName" />
          <v-menu>
            <template v-slot:activator="{ props }">
              <v-btn density="compact" icon="mdi-dots-vertical" variant="plain" v-bind="props"></v-btn>
            </template>
            <v-list>

              <!-- RENAME FILE -->
              <v-list-item v-if="!file.DeletedAt" @click="() => {renameFilePlaceholder = file.FileName}">
                <v-icon>mdi-pencil</v-icon> Rename

                <v-dialog activator="parent" max-width="500px">
                  <template v-slot:default="{ isActive }">
                    <v-card>
                      <v-card-title title>
                        Rename {{ file.FileName }}
                      </v-card-title>

                      <v-card-item>
                        <v-text-field
                          v-model="renameFilePlaceholder"
                          label="New filename"
                          single-line
                          :rules="[rules.required]"
                        ></v-text-field>
                        <div class="tw-text-red-500">Caution: Avoid altering the file extension to prevent unexpected behavior</div>
                      </v-card-item>

                      <v-card-actions>
                        <v-btn @click="isActive.value = false">Cancel</v-btn>
                        <v-btn variant="outlined" @click="() => {
                          isActive.value = false
                          renameFile();
                        }">Rename</v-btn>
                      </v-card-actions>
                    </v-card>
                  </template>
                </v-dialog>

              </v-list-item>

              <!-- DOWNLOAD -->
              <v-list-item v-if="!file.DeletedAt" @click="getFileURL">
                <v-icon>mdi-download</v-icon> Download
              </v-list-item>

              <!-- MOVE FILE -->
              <v-list-item v-if="!file.DeletedAt" @click="moveFile">
                <v-icon>mdi-folder-arrow-right</v-icon> <span class="tw-ml-1">Move to</span>
              </v-list-item>

              <!-- FILE DETAILS DIALOG -->
              <v-list-item @click="() => { }">
                <!-- DETAILS DIALOG -->
                <v-dialog activator="parent" max-width="30rem" v-model="fileDetailDialog">
                  <template v-slot:default="{ isActive: _ }">
                    <v-card>
                      <v-card-title>
                        <div class="tw-flex tw-flex-row tw-justify-between tw-items-center tw-px-2">
                          <p>File details</p>
                          <v-btn icon="mdi-close" variant="flat" @click="fileDetailDialog = false"></v-btn>
                        </div>
                      </v-card-title>
                      <v-card-text>
                        <div v-for="(value, index) in Object.entries(
                          fileDetailFormatter(file)
                        )" :key="index" class="tw-flex tw-flex-col">
                          <div class="tw-flex tw-flex-row tw-justify-start">
                            <p class="tw-w-1/2">{{ value[0] }}</p>
                            <p class="tw-w-1/2">{{ value[1] }}</p>
                          </div>
                          <v-divider class="tw-my-2" v-if="
                            index !=
                            Object.entries(fileDetailFormatter(file)).length - 1
                          "></v-divider>
                        </div>
                      </v-card-text>
                    </v-card>
                  </template>
                </v-dialog>

                <v-icon>mdi-information-outline</v-icon> Details
              </v-list-item>

              <!-- TOGGLE FAVORITE FILES -->
              <v-list-item @click="toggleFavorite" v-if="!file.DeletedAt">
                <span v-if="!file.IsFavorite"><v-icon>mdi-star-outline</v-icon>Mark as favorite</span>
                <span v-else><v-icon>mdi-star</v-icon>Unfavorite</span>
              </v-list-item>

              <!-- DELETE FILE -->
              <v-list-item v-if="!file.DeletedAt" @click="trashFile(file, true)">
                <v-icon>mdi-trash-can</v-icon> Delete
              </v-list-item>

              <!-- RESTORE DELETED FILE -->
              <v-list-item v-else @click="restoreFile">
                <v-icon>mdi-delete-restore</v-icon> Restore
              </v-list-item>

              <!-- PERMANENTLY DELETE FILE -->
              <v-list-item v-if="file.DeletedAt" @click="() => { }">
                <v-icon>mdi-delete-forever</v-icon> Prune
                <v-dialog activator="parent" max-width="340">
                  <template v-slot:default="{ isActive }">
                    <v-card prepend-icon="mdi-alert" text="Pruning will delete the file permanently, are you sure?"
                      title="Confirmation Dialog">
                      <template v-slot:actions>
                        <v-btn class="" @click="isActive.value = false">Cancel</v-btn>
                        <v-btn class="text-red" variant="outlined" @click="() => {
                          isActive.value = false
                          pruneFile()
                        }">Delete</v-btn>
                      </template>
                    </v-card>
                  </template>
                </v-dialog>
              </v-list-item>
            </v-list>
          </v-menu>
        </div>

          <v-card-subtitle>
            <span class="me-1">{{ formatDistance(file.UpdatedAt, new Date(), { addSuffix: true }) }}</span>
          </v-card-subtitle>
        </v-card-item>
      </v-card>
    </template>
  </v-tooltip>
</template>

<script setup></script>
