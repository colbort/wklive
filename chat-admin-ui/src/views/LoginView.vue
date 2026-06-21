<script setup lang="ts">
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();
const username = ref("");
const password = ref("");
const googleCode = ref("");
const loading = ref(false);

async function submit() {
  loading.value = true;
  try {
    await auth.loginWithPassword({
      username: username.value,
      password: password.value,
      googleCode: googleCode.value || undefined,
    });
    ElMessage.success("登录成功");
    const redirect =
      typeof route.query.redirect === "string"
        ? route.query.redirect
        : auth.homePath;
    router.replace(redirect);
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : "登录失败");
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="login-page">
    <section class="login-panel">
      <div class="login-copy">
        <div class="page-kicker">
          Chat Admin
        </div>
        <h1>客服工作台</h1>
        <p>登录后会根据账号身份进入商户管理台或坐席接待台。</p>
      </div>

      <el-form
        label-position="top"
        class="login-form"
        @submit.prevent="submit"
      >
        <el-form-item label="账号">
          <el-input
            v-model="username"
            autocomplete="username"
            placeholder="请输入账号"
          />
        </el-form-item>
        <el-form-item label="密码">
          <el-input
            v-model="password"
            type="password"
            autocomplete="current-password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="Google 验证码">
          <el-input
            v-model="googleCode"
            inputmode="numeric"
            placeholder="未开启可不填"
          />
        </el-form-item>
        <el-button
          type="primary"
          native-type="submit"
          class="login-button"
          :loading="loading"
        >
          登录
        </el-button>
      </el-form>
    </section>
  </div>
</template>
