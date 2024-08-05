<script setup lang="ts">
import { ref } from "vue";
import { formatDistance } from "date-fns";
import { fileDetailFormatter } from "../utils/fileDetailFormatter";
import Filename from "./Filename.vue";
import { CloudChestFile } from "../models/file";
import { trashFile, updateFile } from "../utils/filesApi";
const props = defineProps<{
  file: CloudChestFile
}>();

const file = props.file;
const fileDetailDialog = ref(false);

async function toggleFavorite(): Promise<void> {
  const fileCopy: CloudChestFile = file;
  fileCopy.IsFavorite = !file.IsFavorite;
  const isSuccessful: boolean = await updateFile(fileCopy, false);
  if (isSuccessful) {
    file.IsFavorite = !file.IsFavorite
  }
}

async function restoreFile(): Promise<void> {
  const isSuccessful: boolean = await updateFile(file, true);
  if (isSuccessful) {
    file.DeletedAt = null;
  }
}

async function pruneFile(): Promise<void> {
  try {
    await trashFile(file, false);
  } catch (error) {
    console.error(error);
  }
}
</script>

<template>
  <v-tooltip :text="file.FileName" location="bottom">
    <template v-slot:activator="{ props: tltpProps }">
      <v-card :disabled="loading" :loading="loading" class="mx-auto" max-width="374" @click="" v-bind="tltpProps">
        <template v-slot:loader="{ isActive }">
          <v-progress-linear :active="isActive" color="deep-purple" height="4" indeterminate></v-progress-linear>
        </template>

        <v-img height="100" src="https://picsum.photos/id/11/100/60" cover alt="No thumbnail"></v-img>

        <v-card-item>
          <div class="tw-flex tw-flex-row tw-h-full tw-w-full tw-items-center tw-justify-between">
          <Filename :filename="file.FileName" />
          <v-menu>
            <template v-slot:activator="{ props }">
              <v-btn density="compact" icon="mdi-dots-vertical" variant="plain" v-bind="props"></v-btn>
            </template>
            <v-list>
              <v-list-item @click="console.log('Download')">
                <v-icon>mdi-download</v-icon> Download
              </v-list-item>
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
              <v-list-item @click="toggleFavorite" v-if="!file.DeletedAt">
                <span v-if="!file.IsFavorite"><v-icon>mdi-star-outline</v-icon>Mark as favorite</span>
                <span v-else><v-icon>mdi-star</v-icon>Unfavorite</span>
              </v-list-item>
              <v-list-item v-if="!file.DeletedAt" @click="trashFile(file, true)">
                <v-icon>mdi-trash-can</v-icon> Delete
              </v-list-item>
              <v-list-item v-else @click="restoreFile">
                <v-icon>mdi-delete-restore</v-icon> Restore
              </v-list-item>
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
