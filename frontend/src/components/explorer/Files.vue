<script setup lang="ts">
import { useRouter, useRoute } from "vue-router";
import { Ref, ref, onBeforeMount } from "vue";
import { MinIOFile } from "../../models/file";
import { getAllFiles } from "../../utils/filesApi";
import { formatDistance } from "date-fns";
import fileDetailFormatter from "../../utils/fileDetailFormatter";

const fileDetailDialogActivator = ref(undefined);
const files: Ref<MinIOFile[]> = ref([] as MinIOFile[]);

const router = useRouter();
const route = useRoute();

onBeforeMount(async () => {
  files.value = await getAllFiles();
  console.log(files.value);
});
</script>

<template>
  <v-container>
    <v-row>
      <v-col v-for="file in files" :key="file.ID" :cols="2">
        <v-card max-width="10rem" class="pa-2 rounded-lg" hover @click="">
          <!-- Upper part (file name and menu) -->
          <div
            class="tw-flex tw-flex-row tw-h-full tw-mb-3 tw-w-full tw-items-center tw-flex-wrap"
          >
            <p class="text-body-2 tw-grow">{{ file.FileName }}</p>
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
                <v-list-item @click="() => {}">
                  <!-- DETAILS DIALOG -->
                  <v-dialog activator="parent" max-width="30rem">
                    <template v-slot:default="{ isActive }">
                      <v-card title="File detail">
                        <v-card-text>
                          <table>
                            <tr
                              v-for="(value, index) in Object.entries(fileDetailFormatter(file))"
                              :key="index">
                              <th
                                :class="{
                                  'tw-bg-grey-lighten-3': index % 2 == 0,
                                }">
                                <span>{{ value[0] }}:</span>
                              </th>
                              <th
                                :class="{
                                  'tw-bg-grey-lighten-3 tw-text-base': index % 2 == 1,
                                }">
                                <span>{{ value[1] }}</span>
                              </th>
                            </tr>
                          </table>
                        </v-card-text>
                      </v-card>
                    </template>
                  </v-dialog>

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
          <p class="text-caption">
            {{
              formatDistance(file.UpdatedAt, new Date(), { addSuffix: true })
            }}
          </p>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped></style>
