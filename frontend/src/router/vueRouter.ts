import Login from '../components/Login.vue';
import SignUp from '../components/SignUp.vue';
import Dashboard from '../components/Dashboard.vue';
import Files from '../components/explorer/Files.vue';
import Favorite from '../components/explorer/Favorite.vue';
import Trash from '../components/explorer/Trash.vue';
import { createWebHistory, createRouter } from 'vue-router';
import checkTokenValidation from '../utils/checkTokenValidation';

const routes = [
  { path: '/', redirect: '/login' },
  { path: '/login', component: Login, name: 'login' },
  { path: '/signup', component: SignUp, name: 'signup' },
  { path: '/explorer', redirect: '/explorer/files' },
  { 
    path: '/explorer', 
    component: Dashboard,
    children: [
      {
        path: 'files',
        component: Files,
        name: 'explorer-files',
      },
      {
        path: 'favorite',
        component: Favorite
      },
      {
        path: 'trash',
        component: Trash
      },
    ]
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

/**
 * if token is not valid and route is not login nor signup, redirect to login
 * else if token is valid and route is login or signup, redirect to explorer
 * else if token is valid and route is explorer, redirect to explorer
 */
router.beforeEach(async (to, from, next) => {
  try {
    const isValid = await checkTokenValidation();
    
    if (!isValid && to.name !== 'login' && to.name !== 'signup') {
      next({ name: 'login' });
    } else if (isValid && (to.name === 'login' || to.name === 'signup')) {
      next({ name: 'explorer-files' });
    } else {
      next();
    }
  } catch (error) {
    console.error('Router Token Check Error:', error);
    next(false);
  }
});

export default router;
