<script setup lang="ts">
import { useRouter, useRoute } from "vue-router";
import { Ref, ref, onBeforeMount } from "vue";
import { MinIOFile } from "../../models/file";
import { getFilesFromPath } from "../../utils/filesApi";
import File from "../File.vue";
import Folder from "../Folder.vue";
import FolderModel from "../../models/folder";

const fileDetailDialog = ref(false);
const fileList: Ref<MinIOFile[]> = ref([] as MinIOFile[]);
const folderList: Ref<FolderModel[]> = ref([] as FolderModel[]);

const path = ref("root");

const router = useRouter();
const route = useRoute();

onBeforeMount(async () => {
  const response = await getFilesFromPath(path.value);
  fileList.value = response.files;
  folderList.value = response.folders;
});
</script>

<template>
  <v-container>
    <v-row>
      <v-col v-for="folder in folderList" :key="folder" :cols="2">
        <Folder :folder="folder" />
      </v-col>
      <v-col v-for="file in fileList" :key="file" :cols="2">
        <File :file="file" />
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped></style>
