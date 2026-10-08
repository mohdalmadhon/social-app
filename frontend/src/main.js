import { createApp } from 'vue'
import App from './App.vue';
import { router } from './router/router.js';
import { initTheme } from './helpers/common/theme.js';
import "./styles/global.css"

initTheme();

createApp(App)
    .use(router)
    .mount('#app')