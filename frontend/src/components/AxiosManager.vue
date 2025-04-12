<script setup lang="ts">
import { ref } from "vue";
import AxiosRequest from "./AxiosRequest.vue";
import { useAxiosManagerStore } from "../stores/axiosManagerStore.ts";

const expand = ref<boolean>(true);

const axiosManager = useAxiosManagerStore();
</script>

<template>
    <div v-bind="$attrs" class="tw-max-h-[calc(100dvh-100px)] tw-w-72 tw-mb-4 tw-mr-4 tw-flex tw-flex-col-reverse tw-overflow-hidden">
        <v-card elevation="5">
            <v-card-actions>
                <v-card-item title="Uploads"></v-card-item>
                <v-spacer></v-spacer>
                <v-btn :icon="expand ? 'mdi-chevron-up' : 'mdi-chevron-down'" @click="expand = !expand"></v-btn>
            </v-card-actions>
            <v-divider></v-divider>
            <v-expand-transition>
                <div v-show="expand">
                    <!-- <AxiosRequest v-for="request in axiosManager.ongoingRequests" :request="request" :key="request.id" /> -->
                    <AxiosRequest v-for="request in axiosManager.ongoingRequests" :request="request" :key="request.id" />
                </div>
            </v-expand-transition>
        </v-card>
    </div>
</template>