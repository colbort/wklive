<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
const form = reactive({ username: "", password: "" });
const loading = ref(false);
const auth = useAuthStore(), router = useRouter(), route = useRoute();
async function submit() {
  loading.value = true;
  try { await auth.login(form.username, form.password); await router.replace(String(route.query.redirect || "/")); }
  finally { loading.value = false; }
}
</script>
<template>
  <div class="login">
    <div class="grid"></div>
    <section>
      <div class="logo">LQ</div><p class="eyebrow">WKLIVE INFRASTRUCTURE</p>
      <h1>流动性管理中心</h1><p class="sub">做市策略、外部流动性、风险与对账统一控制台</p>
      <el-form @submit.prevent="submit">
        <label>管理员账号</label><el-input v-model="form.username" size="large" placeholder="请输入账号" />
        <label>登录密码</label><el-input v-model="form.password" size="large" type="password" show-password placeholder="请输入密码" @keyup.enter="submit" />
        <el-button type="primary" size="large" :loading="loading" @click="submit">进入控制台</el-button>
      </el-form>
      <p class="hint">访问行为将记录至安全审计日志</p>
    </section>
  </div>
</template>
<style scoped>
.login { min-height: 100vh; display: grid; place-items: center; overflow: hidden; background: radial-gradient(circle at 50% 15%, #17344d, #07101d 43%); }
.grid { position: fixed; inset: 0; opacity: .13; background-image: linear-gradient(#8ba0b5 1px, transparent 1px), linear-gradient(90deg,#8ba0b5 1px,transparent 1px); background-size: 56px 56px; mask-image: linear-gradient(to bottom,black,transparent 80%); }
section { position: relative; width: 430px; padding: 42px; border: 1px solid rgba(116,150,184,.2); border-radius: 20px; background: rgba(8,19,33,.9); box-shadow: 0 34px 100px rgba(0,0,0,.45); }
.logo { display: grid; place-items: center; width: 48px; height: 48px; margin-bottom: 18px; border: 1px solid #2dd4a4; border-radius: 12px; color: #2dd4a4; font-weight: 800; }
.eyebrow { margin: 0 0 8px; color: #2dd4a4; font-size: 10px; letter-spacing: 2px; }h1 { margin: 0; color: #f2f7fb; font-size: 27px; }.sub { margin: 10px 0 30px; color: #63758d; font-size: 13px; }
label { display: block; margin: 17px 0 8px; color: #8191a6; font-size: 12px; }.el-button { width: 100%; margin-top: 26px; --el-color-primary: #19c694; }.hint { margin: 23px 0 0; text-align: center; color: #43546a; font-size: 10px; }
</style>
