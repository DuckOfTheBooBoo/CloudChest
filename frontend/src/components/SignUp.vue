<script setup lang="ts">
import { ref } from "vue";
import { useField, useForm } from "vee-validate";
import axios from "axios";
import {useRouter} from "vue-router"

interface SignUpBody {
  first_name: string;
  last_name: string;
  email: string;
  password: string;
}

interface ErrorResponse {
  error: string
}

const router = useRouter();

const { handleSubmit } = useForm({
  validationSchema: {
    first_name(value: string) {
      if (value?.length >= 2) return true;

      return "At least 2 characters.";
    },
    last_name(value: string) {
      if (value?.length >= 2) return true;

      return "At least 2 characters.";
    },
    email(value: string) {
      if (
        /^([+\w-]+(?:\.[+\w-]+)*)@(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/.test(
          value
        )
      )
        return true;

      return "Must be a valid e-mail.";
    },
    password(value: string) {
      if (/(?=.*\d).{8}/.test(value)) return true;

      return "Password must consists of 8 characters and include numbers.";
    },
  },
});

const firstName = useField("first_name");
const lastName = useField("last_name");
const email = useField("email");
const password = useField("password");

let visible = ref(false);
let loading = ref(false);
let currentProgMsg = ref("");

const submit = handleSubmit(async (values) => {
  const body: SignUpBody = {
    first_name: values.first_name,
    last_name: values.last_name,
    email: values.email,
    password: values.password,
  };

  try {
    loading.value = true;
    const response = await axios.post(
      "http://localhost:3000/api/users/register",
      body
    );

    if (response.status == 201) {
      loading.value = false;
      currentProgMsg.value = "Sign up successful. Redirecting...";
      setTimeout(() => {
        router.push("/login");
      }, 1000);
    } else {
      const err: ErrorResponse = response.data;
      currentProgMsg.value = err.error;
    }
    
  } catch (error: Error | any) {
    console.error(error);
    currentProgMsg.value = error.response.data.error;
  }
  loading.value = false;
});
</script>

<template>
  <div class="parent-div">
    <v-card
      max-width="25rem"
      class="justify-center form-card"
      elevation="16"
      :disabled="loading"
      :loading="loading"
    >
      <template v-slot:loader="{ isActive }">
        <v-progress-linear
          :active="isActive"
          color="blue"
          height="4"
          indeterminate
        ></v-progress-linear>
      </template>
      <v-card-title primary-title>
        <div>
          <h3 class="text-center text-h3 bold my-4">HELLO!</h3>
        </div>
      </v-card-title>
      <form @submit.prevent="submit">
        <v-card-text>
          <div class="d-flex ga-2 justify-space-between d-sm-flex-row">
            <v-text-field
              min-width="50%"
              density="compact"
              v-model="firstName.value.value"
              :error-messages="firstName.errorMessage.value"
              label="First name"
              id="first_name"
              variant="outlined"
              spellcheck="false"
            ></v-text-field>
            <v-text-field
              min-width="50%"
              density="compact"
              v-model="lastName.value.value"
              :error-messages="lastName.errorMessage.value"
              label="Last name"
              id="last_name"
              variant="outlined"
              spellcheck="false"
            ></v-text-field>
          </div>
          <v-text-field
            prepend-inner-icon="mdi-email-outline"
            density="compact"
            v-model="email.value.value"
            :error-messages="email.errorMessage.value"
            label="E-mail"
            id="email"
            variant="outlined"
            spellcheck="false"
          ></v-text-field>
          <v-text-field
            :append-inner-icon="visible ? 'mdi-eye-off' : 'mdi-eye'"
            prepend-inner-icon="mdi-lock-outline"
            density="compact"
            v-model="password.value.value"
            :error-messages="password.errorMessage.value"
            label="Password"
            id="password"
            variant="outlined"
            spellcheck="false"
            @click:append-inner="visible = !visible"
          ></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-btn
            class="mb-2"
            color="blue"
            size="large"
            variant="tonal"
            type="submit"
            block
          >
            REGISTER
          </v-btn>
        </v-card-actions>
      </form>
      <div class="pa-2 d-flex justify-center align-center">
        <RouterLink to="/login" class="text-grey-lighten-2 text-decoration-none"
          >Log in instead?</RouterLink
        >
      </div>
    </v-card>
    <div class="progress-msg mt-4">
      <p class="d-block" v-show="currentProgMsg != ''">{{ currentProgMsg }}</p>
    </div>
  </div>
</template>

<style scoped>
.parent-div {
  display: flex;
  flex-direction: column;
  height: 100vh;
  justify-content: center;
  align-items: center;
}
.form-card {
  min-width: 25rem !important;
}

.progress-msg {
  min-height: 1.5rem;
}
</style>
