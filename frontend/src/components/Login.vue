<script setup lang="ts">
import { ref } from "vue";
import { useField, useForm } from "vee-validate";
import axios from "axios";
import { useRouter } from "vue-router";

const router = useRouter();

interface LoginBody {
  email: string;
  password: string;
}

interface TokenResponse {
  token: string;
}

const { handleSubmit } = useForm({
  validationSchema: {
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
      if (/(?=.*\d).{6}/.test(value)) return true;

      return "Password must consists of 6 characters and include numbers.";
    },
  },
});

const email = useField("email");
const password = useField("password");

let visible = ref(false);
let loading = ref(false);
let currentProgMsg = ref("");

const submit = handleSubmit(async (values) => {
  const body: LoginBody = {
    email: values.email,
    password: values.password,
  };

  try {
    loading.value = true;
    const response = await axios.post<TokenResponse>(
      "http://localhost:3000/api/users/login",
      body
    );
    if (response.status == 200) {
      localStorage.setItem("token", response.data.token);
      axios.defaults.headers.common['Authorization'] = 'Bearer ' + localStorage.getItem('token');
      loading.value = false;
      currentProgMsg.value = "Login successful. Redirecting...";
      router.push("/explorer");
    } else {
      currentProgMsg.value = "Login failed. Please try again.";
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
          <h3 class="text-center text-h3 bold my-4">Welcome Back!</h3>
        </div>
      </v-card-title>
      <form @submit.prevent="submit">
        <v-card-text>
          <v-text-field
            class="tw-mb-2"
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
            LOG IN
          </v-btn>
        </v-card-actions>
      </form>
      <div class="pa-2 d-flex justify-center align-center">
        <RouterLink
          to="/signup"
          class="text-grey-lighten-2 text-decoration-none"
          >Sign Up instead?</RouterLink
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
