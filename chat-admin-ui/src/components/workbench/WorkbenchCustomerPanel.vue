<script setup lang="ts">
import { computed } from "vue";
import type { ChatSession } from "@/types/chat";

const props = defineProps<{
  session?: ChatSession;
}>();

const displayName = computed(
  () => props.session?.userNickname || props.session?.title || "访客",
);
</script>

<template>
  <aside class="customer-panel workbench-region">
    <div class="panel-header">
      <h2>客户信息</h2>
    </div>
    <div class="customer-card">
      <el-avatar
        :size="44"
        :src="session?.userAvatarUrl || undefined"
      >
        {{ displayName.slice(0, 1) }}
      </el-avatar>
      <div>
        <strong>{{ displayName }}</strong>
        <span>ID: {{ session?.userId || "-" }}</span>
      </div>
    </div>
    <dl class="info-list">
      <div>
        <dt>用户 ID</dt>
        <dd>{{ session?.userId || "-" }}</dd>
      </div>
      <div>
        <dt>问题分类</dt>
        <dd>{{ session?.category || "-" }}</dd>
      </div>
      <div>
        <dt>优先级</dt>
        <dd>{{ session?.priority === 3 ? "高" : "普通" }}</dd>
      </div>
      <div>
        <dt>分组 ID</dt>
        <dd>{{ session?.groupId || "-" }}</dd>
      </div>
    </dl>

    <div class="note-block">
      <h3>接待备注</h3>
      <el-input
        type="textarea"
        resize="none"
        :rows="5"
        placeholder="记录客户偏好、订单号或处理进展"
      />
    </div>
  </aside>
</template>
