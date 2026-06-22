<script setup lang="ts">
import type { ChatSession } from "@/types/chat";

defineProps<{
  session?: ChatSession;
}>();
</script>

<template>
  <aside class="customer-panel workbench-region">
    <div class="panel-header">
      <h2>客户信息</h2>
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
