<script setup lang="ts">
import Filename from "./Filename.vue";
import type RequestData from "../models/requestData";
import { ref, watch, onUpdated } from "vue";
import { useAxiosManagerStore } from "../stores/axiosManagerStore.ts";

const props = defineProps<{
    request: RequestData;
}>();

const axiosManager = useAxiosManagerStore();

const UPLOADING = 'UPLOADING';
const CANCELLED = "CANCELLED";
const COMPLETED = "COMPLETED";
const FAILED = "FAILED";
type State = typeof UPLOADING | typeof CANCELLED | typeof COMPLETED | typeof FAILED | null;

const state = ref<State>(UPLOADING);
const visProgress = ref<boolean>(true);
const visIcon = ref<boolean>(false);

const iconState = ref<string>("mdi-close-circle-outline");

const iconAction = (): void => {
    switch (state.value) {
        case UPLOADING: // CANCEL REQUEST
            axiosManager.cancelRequest(props.request.id);
            iconState.value = "mdi-close-circle";
            state.value = CANCELLED;
            visIcon.value = true;
            visProgress.value = false;
            break;
        case COMPLETED:
            axiosManager.removeRequest(props.request.id);
            break;
        case CANCELLED: // RETRY REQUEST
        case FAILED: // RETRY REQUEST
        default:
            axiosManager.removeRequest(props.request.id);
            state.value = null;
            break;
    }
}


// let isHover: boolean = false;
const onHover = (): void => {
    if (visProgress.value && !visIcon.value) {
        visProgress.value = false;
        visIcon.value = true;
    }
}

const onLeave = (): void => {
    visProgress.value = true;
    visIcon.value = false;
}

watch(() => props.request.progress, () => {
    if (props.request.progress >= 100) {
        state.value = COMPLETED;
        visProgress.value = false;
        visIcon.value = true;
    }
})

watch(() => state.value, () => {
    switch(state.value) {
        case COMPLETED:
            iconState.value = "mdi-check-circle";
            break;
        case CANCELLED:
            iconState.value = "mdi-close-circle";
            break;
        case FAILED:
            iconState.value = "mdi-alret-circle";
            break;
    }
})

onUpdated(() => {
    console.log(props)
})
</script>

<template>
    <v-card-text class="tw-flex tw-flex-row tw-justify-between">
        <Filename :filename="request.filename" class="" />
        <div>
            <v-progress-circular @mouseover="onHover" v-if="state === UPLOADING && visProgress" :size="20" :model-value="request.progress" id="progress-circular" ></v-progress-circular>
            <v-icon v-else-if="state === CANCELLED || state === FAILED || state === COMPLETED || visIcon" @mouseleave="onLeave" @click="iconAction" :icon="iconState"></v-icon>
        </div>
    </v-card-text>
    <v-divider></v-divider>
</template>