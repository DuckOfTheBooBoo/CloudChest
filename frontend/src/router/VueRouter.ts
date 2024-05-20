import Login from "../components/Login.vue";
import SignUp from "../components/SignUp.vue";
import App from "../App.vue";
import { createWebHistory, createRouter } from "vue-router";
import checkTokenValidation from "../utils/checkTokenValidation";

const routes = [
  { path: "/", component: Login },
  { path: "/login", component: Login },
  { path: "/signup", component: SignUp },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// router.beforeEach(async (to, from, next) => {
//   try {
//     const isValid = await checkTokenValidation();
//     if (!isValid) {
//       return next("/login");
//     }
//   } catch (error) {
//     console.error("Router Token Check Error: ", error);
//   }
//   next();
//     next({name: 'login'})
// });

export default router;
