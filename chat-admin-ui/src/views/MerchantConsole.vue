<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  createAgent,
  createCategory,
  createGroup,
  deleteCategory,
  deleteGroup,
  pageAgents,
  pageCategories,
  pageGroups,
  updateAgent,
  updateAgentStatus,
  updateCategory,
  updateGroup,
  type CreateChatAgentPayload,
  type UpdateChatAgentPayload,
  type ChatCategoryPayload,
  type ChatGroupPayload,
} from "@/api/chat";
import MerchantAgentTable from "@/components/merchant/MerchantAgentTable.vue";
import MerchantCategoryTable from "@/components/merchant/MerchantCategoryTable.vue";
import MerchantEditDialog from "@/components/merchant/MerchantEditDialog.vue";
import MerchantGroupTable from "@/components/merchant/MerchantGroupTable.vue";
import { useAuthStore } from "@/stores/auth";
import type { ChatAgent, ChatCategory, ChatGroup } from "@/types/chat";

type TabName = "agents" | "categories" | "groups";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const activeTab = computed<TabName>({
  get: () => String(route.meta.activeTab || "agents") as TabName,
  set: (value) => router.push(`/merchant/${value}`),
});

const merchantId = computed(() => auth.user?.merchantId || 0);
const loading = ref(false);
const keyword = ref("");
const agents = ref<ChatAgent[]>([]);
const categories = ref<ChatCategory[]>([]);
const groups = ref<ChatGroup[]>([]);

const dialogRef = ref<InstanceType<typeof MerchantEditDialog>>();
const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const editingId = ref(0);

const agentForm = reactive({
  username: "",
  password: "",
  nickname: "",
  mobile: "",
  email: "",
  enabled: 1,
  maxSessionCount: 10,
  groupId: undefined as number | undefined,
  welcomeMessage: "",
  remark: "",
});

const categoryForm = reactive({
  parentId: 0,
  categoryCode: "",
  categoryName: "",
  enabled: 1,
  sort: 0,
  remark: "",
});

const groupForm = reactive({
  groupCode: "",
  groupName: "",
  description: "",
  enabled: 1,
  sort: 0,
  remark: "",
});

const dialogTitle = computed(() => {
  const action = dialogMode.value === "create" ? "新增" : "编辑";
  const resource =
    activeTab.value === "agents"
      ? "坐席"
      : activeTab.value === "categories"
        ? "问题分类"
        : "客服分组";
  return `${action}${resource}`;
});

const enabledOptions = [
  { label: "启用", value: 1 },
  { label: "禁用", value: 2 },
];

const statusOptions = [
  { label: "离线", value: 1 },
  { label: "在线", value: 2 },
  { label: "忙碌", value: 3 },
  { label: "休息", value: 4 },
];

watch(
  () => activeTab.value,
  () => {
    keyword.value = "";
    loadCurrent();
  },
  { immediate: true },
);

async function loadCurrent() {
  if (!merchantId.value) return;
  loading.value = true;
  try {
    const params = {
      merchantId: merchantId.value,
      limit: 100,
      keyword: keyword.value || undefined,
    };
    if (activeTab.value === "agents") {
      agents.value = (await pageAgents(params)).data;
      if (!groups.value.length) {
        groups.value = (await pageGroups({ merchantId: merchantId.value, limit: 100 })).data;
      }
    } else if (activeTab.value === "categories") {
      categories.value = (await pageCategories(params)).data;
    } else {
      groups.value = (await pageGroups(params)).data;
    }
  } finally {
    loading.value = false;
  }
}

function resetForms() {
  editingId.value = 0;
  Object.assign(agentForm, {
    username: "",
    password: "",
    nickname: "",
    mobile: "",
    email: "",
    enabled: 1,
    maxSessionCount: 10,
    groupId: undefined,
    welcomeMessage: "",
    remark: "",
  });
  Object.assign(categoryForm, {
    parentId: 0,
    categoryCode: "",
    categoryName: "",
    enabled: 1,
    sort: 0,
    remark: "",
  });
  Object.assign(groupForm, {
    groupCode: "",
    groupName: "",
    description: "",
    enabled: 1,
    sort: 0,
    remark: "",
  });
  nextTick(() => dialogRef.value?.clearValidate());
}

function openCreate() {
  dialogMode.value = "create";
  resetForms();
  dialogVisible.value = true;
}

function openAgentEdit(row: ChatAgent) {
  dialogMode.value = "edit";
  resetForms();
  editingId.value = row.id;
  Object.assign(agentForm, {
    username: "",
    password: "",
    nickname: "",
    mobile: "",
    email: "",
    enabled: 1,
    maxSessionCount: row.maxSessionCount || 10,
    groupId: row.groupId || undefined,
    welcomeMessage: row.welcomeMessage,
    remark: row.remark,
  });
  dialogVisible.value = true;
}

function openCategoryEdit(row: ChatCategory) {
  dialogMode.value = "edit";
  resetForms();
  editingId.value = row.id;
  Object.assign(categoryForm, {
    parentId: row.parentId,
    categoryCode: row.categoryCode,
    categoryName: row.categoryName,
    enabled: row.enabled,
    sort: row.sort,
    remark: row.remark,
  });
  dialogVisible.value = true;
}

function openGroupEdit(row: ChatGroup) {
  dialogMode.value = "edit";
  resetForms();
  editingId.value = row.id;
  Object.assign(groupForm, {
    groupCode: row.groupCode,
    groupName: row.groupName,
    description: row.description,
    enabled: row.enabled,
    sort: row.sort,
    remark: row.remark,
  });
  dialogVisible.value = true;
}

async function submitDialog() {
  if (!merchantId.value) return;
  await dialogRef.value?.validate();

  if (activeTab.value === "agents") {
    if (dialogMode.value === "create") {
      const payload: CreateChatAgentPayload = {
        maxSessionCount: agentForm.maxSessionCount,
        groupId: agentForm.groupId,
        welcomeMessage: agentForm.welcomeMessage,
        remark: agentForm.remark,
        username: agentForm.username,
        password: agentForm.password,
        nickname: agentForm.nickname,
        mobile: agentForm.mobile,
        email: agentForm.email,
        enabled: agentForm.enabled,
      };
      await createAgent(payload);
    } else {
      const payload: UpdateChatAgentPayload = {
        merchantId: merchantId.value,
        maxSessionCount: agentForm.maxSessionCount,
        groupId: agentForm.groupId,
        welcomeMessage: agentForm.welcomeMessage,
        remark: agentForm.remark,
      };
      await updateAgent(editingId.value, payload);
    }
  } else if (activeTab.value === "categories") {
    const payload: ChatCategoryPayload = {
      merchantId: merchantId.value,
      parentId: categoryForm.parentId,
      categoryCode: categoryForm.categoryCode || undefined,
      categoryName: categoryForm.categoryName,
      enabled: categoryForm.enabled,
      sort: categoryForm.sort,
      remark: categoryForm.remark,
    };
    if (dialogMode.value === "create") {
      await createCategory(payload);
    } else {
      await updateCategory(editingId.value, payload);
    }
  } else {
    const payload: ChatGroupPayload = {
      merchantId: merchantId.value,
      groupCode: groupForm.groupCode || undefined,
      groupName: groupForm.groupName,
      description: groupForm.description,
      enabled: groupForm.enabled,
      sort: groupForm.sort,
      remark: groupForm.remark,
    };
    if (dialogMode.value === "create") {
      await createGroup(payload);
    } else {
      await updateGroup(editingId.value, payload);
    }
  }

  dialogVisible.value = false;
  ElMessage.success("保存成功");
  await loadCurrent();
}

async function changeAgentStatus(row: ChatAgent, status: number) {
  await updateAgentStatus(row.id, {
    merchantId: merchantId.value,
    status,
  });
  ElMessage.success("状态已更新");
  await loadCurrent();
}

async function removeCategory(row: ChatCategory) {
  await ElMessageBox.confirm(`确认删除分类「${row.categoryName}」？`, "删除确认", {
    type: "warning",
  });
  await deleteCategory(row.id, merchantId.value);
  ElMessage.success("删除成功");
  await loadCurrent();
}

async function removeGroup(row: ChatGroup) {
  await ElMessageBox.confirm(`确认删除分组「${row.groupName}」？`, "删除确认", {
    type: "warning",
  });
  await deleteGroup(row.id, merchantId.value);
  ElMessage.success("删除成功");
  await loadCurrent();
}

</script>

<template>
  <section class="console-page">
    <div class="toolbar">
      <el-tabs v-model="activeTab">
        <el-tab-pane
          label="坐席管理"
          name="agents"
        />
        <el-tab-pane
          label="问题分类"
          name="categories"
        />
        <el-tab-pane
          label="客服分组"
          name="groups"
        />
      </el-tabs>
      <div class="toolbar-actions">
        <el-input
          v-model="keyword"
          clearable
          class="search-input"
          placeholder="搜索"
          @keyup.enter="loadCurrent"
          @clear="loadCurrent"
        />
        <el-button @click="loadCurrent">
          查询
        </el-button>
        <el-button
          type="primary"
          @click="openCreate"
        >
          新增{{
            activeTab === "agents"
              ? "坐席"
              : activeTab === "categories"
                ? "分类"
                : "分组"
          }}
        </el-button>
      </div>
    </div>

    <div class="table-wrap">
      <MerchantAgentTable
        v-if="activeTab === 'agents'"
        :loading="loading"
        :agents="agents"
        :groups="groups"
        :status-options="statusOptions"
        @edit="openAgentEdit"
        @status-change="changeAgentStatus"
      />

      <MerchantCategoryTable
        v-else-if="activeTab === 'categories'"
        :loading="loading"
        :categories="categories"
        @edit="openCategoryEdit"
        @remove="removeCategory"
      />

      <MerchantGroupTable
        v-else
        :loading="loading"
        :groups="groups"
        @edit="openGroupEdit"
        @remove="removeGroup"
      />
    </div>

    <MerchantEditDialog
      ref="dialogRef"
      v-model:visible="dialogVisible"
      :title="dialogTitle"
      :active-tab="activeTab"
      :dialog-mode="dialogMode"
      :agent-form="agentForm"
      :category-form="categoryForm"
      :group-form="groupForm"
      :groups="groups"
      :enabled-options="enabledOptions"
      @submit="submitDialog"
    />
  </section>
</template>
