import { createRouter, createWebHashHistory } from 'vue-router'
import Home from '@/components/Home'

let routes = [
  {
    path: '/',
    name: 'home',
    component: Home
  }
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
