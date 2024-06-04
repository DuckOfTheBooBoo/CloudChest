<script setup lang="ts">
import { useRoute } from "vue-router";
import { watch, Ref, ref, onBeforeMount, onBeforeUnmount, onMounted } from "vue";
import { MinIOFile } from "../../models/file";
import { getFilesFromPath } from "../../utils/filesApi";
import File from "../File.vue";
import Folder from "../Folder.vue";
import FolderModel from "../../models/folder";

const fileList: Ref<MinIOFile[]> = ref([] as MinIOFile[]);
const folderList: Ref<FolderModel[]> = ref([] as FolderModel[]);

const route = useRoute();
const path = ref('/');

watch(() => route.params.path, (newPath, oldPath) => {
  if (newPath !== undefined) {
    path.value += `/${newPath}`;
  }
}, { immediate: true }); // Capture initial path on component mount


window.addEventListener('popstate', () => {
  console.log(path.value, route.path)
  if(path.value) {
    path.value = route.path; // Update path based on current route
  } else {
    path.value = '/';
  }
});

onBeforeUnmount(() => {
  window.removeEventListener('popstate', () => {});
});

watch(path, async () => {
  path.value = path.value.replace('//', '/')
  console.log(path.value)
  const response = await getFilesFromPath(path.value);
  fileList.value = response.files;
  folderList.value = response.folders;
})

onBeforeMount(async () => {
  path.value = (route.params.path as string).replace('root', '/');
  const response = await getFilesFromPath(path.value);
  fileList.value = response.files;
  folderList.value = response.folders;
});

onMounted(() => {
  console.log(route.fullPath)
})
</script>

<template>
  <v-container>
    <v-row>
      <v-col v-for="folder in folderList" :key="folder" :cols="2">
        <Folder :folder="folder" :parent-path="route.fullPath"/>
      </v-col>
      <v-col v-for="file in fileList" :key="file" :cols="2">
        <File :file="file" />
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped></style>
