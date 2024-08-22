<script setup lang="ts">
import { Ref, ref, onMounted } from "vue";
const props = defineProps<{ filename: string }>();

const pClass = ref("text-body-2 tw-overflow-hidden tw-whitespace-nowrap");

const pTag: Ref<HTMLElement | null> = ref(null);
const overflowing = ref(false);

function checkOverflow() {
  if (pTag.value) {
    // console.log(pTag.value);
    if (pTag.value.scrollWidth > pTag.value.clientWidth) {
      overflowing.value = true;
      pClass.value += " text-fade-out";
    } else {
      overflowing.value = false;
      pClass.value = "text-body-2 tw-overflow-hidden tw-whitespace-nowrap";
    }
  }
}

onMounted(() => {
  checkOverflow();
  window.addEventListener("resize", checkOverflow);
});
</script>

<template>
  <p ref="pTag" :class="pClass" v-bind="$attrs">{{ props.filename }}</p>
</template>

<style scoped>
.text-fade-out {
  position: relative;
  display: inline-block;
  max-width: 100%;
}

.text-fade-out::after {
  content: "";
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 50px;
  background-image: linear-gradient(
    to right,
    rgba(255, 255, 255, 0),
    #212121
  );
}
</style>
