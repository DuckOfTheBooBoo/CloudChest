import { createApp } from 'vue'
import './index.css'
import App from './App.vue'
import router from './router/vueRouter'
import axios from 'axios'


// Vuetify
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import { aliases, mdi } from 'vuetify/iconsets/mdi'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'

axios.defaults.headers.common['Authorization'] = 'Bearer ' + localStorage.getItem('token');


const vuetify = createVuetify({
    components,
    directives,
    theme: {
        defaultTheme: 'dark',
    },
    icons: {
        defaultSet: 'mdi',
        aliases,
        sets: {
          mdi,
        },
    },
})

createApp(App)
    .use(router)
    .use(vuetify)
    .mount('#app')
