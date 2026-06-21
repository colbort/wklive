<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox, type FormInstance } from "element-plus";
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

const formRef = ref<FormInstance>();
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
  nextTick(() => formRef.value?.clearValidate());
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
  await formRef.value?.validate();

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

function statusText(status: number) {
  return statusOptions.find((item) => item.value === status)?.label || "未知";
}

function groupName(groupId: number) {
  return groups.value.find((item) => item.id === groupId)?.groupName || "-";
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
      <el-table
        v-if="activeTab === 'agents'"
        v-loading="loading"
        :data="agents"
        height="100%"
      >
        <el-table-column
          prop="agentNo"
          label="坐席编号"
          width="130"
        />
        <el-table-column
          prop="chatUserId"
          label="用户 ID"
          width="100"
        />
        <el-table-column
          label="状态"
          width="120"
        >
          <template #default="{ row }">
            <el-tag
              :type="
                row.status === 2
                  ? 'success'
                  : row.status === 3
                    ? 'warning'
                    : 'info'
              "
            >
              {{ statusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          label="分组"
          width="140"
        >
          <template #default="{ row }">
            {{ groupName(row.groupId) }}
          </template>
        </el-table-column>
        <el-table-column
          label="接待量"
          width="130"
        >
          <template #default="{ row }">
            {{ row.currentSessionCount }} / {{ row.maxSessionCount }}
          </template>
        </el-table-column>
        <el-table-column
          prop="welcomeMessage"
          label="欢迎语"
          min-width="260"
          show-overflow-tooltip
        />
        <el-table-column
          prop="remark"
          label="备注"
          width="160"
          show-overflow-tooltip
        />
        <el-table-column
          label="操作"
          width="220"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              @click="openAgentEdit(row)"
            >
              编辑
            </el-button>
            <el-dropdown @command="(status: number) => changeAgentStatus(row, status)">
              <el-button link>
                状态
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item
                    v-for="item in statusOptions"
                    :key="item.value"
                    :command="item.value"
                  >
                    {{ item.label }}
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <el-table
        v-else-if="activeTab === 'categories'"
        v-loading="loading"
        :data="categories"
        height="100%"
      >
        <el-table-column
          prop="categoryCode"
          label="分类编码"
          width="180"
        />
        <el-table-column
          prop="categoryName"
          label="分类名称"
          min-width="180"
        />
        <el-table-column
          prop="parentId"
          label="父级 ID"
          width="100"
        />
        <el-table-column
          prop="sort"
          label="排序"
          width="90"
        />
        <el-table-column
          label="状态"
          width="110"
        >
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
              {{ row.enabled === 1 ? "启用" : "禁用" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          label="备注"
          min-width="160"
          show-overflow-tooltip
        />
        <el-table-column
          label="操作"
          width="160"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              @click="openCategoryEdit(row)"
            >
              编辑
            </el-button>
            <el-button
              link
              type="danger"
              @click="removeCategory(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-table
        v-else
        v-loading="loading"
        :data="groups"
        height="100%"
      >
        <el-table-column
          prop="groupCode"
          label="分组编码"
          width="180"
        />
        <el-table-column
          prop="groupName"
          label="分组名称"
          width="160"
        />
        <el-table-column
          prop="description"
          label="描述"
          min-width="220"
          show-overflow-tooltip
        />
        <el-table-column
          prop="sort"
          label="排序"
          width="90"
        />
        <el-table-column
          label="状态"
          width="110"
        >
          <template #default="{ row }">
            <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
              {{ row.enabled === 1 ? "启用" : "禁用" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          label="备注"
          min-width="160"
          show-overflow-tooltip
        />
        <el-table-column
          label="操作"
          width="160"
          fixed="right"
        >
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              @click="openGroupEdit(row)"
            >
              编辑
            </el-button>
            <el-button
              link
              type="danger"
              @click="removeGroup(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="560px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        label-width="104px"
        :model="
          activeTab === 'agents'
            ? agentForm
            : activeTab === 'categories'
              ? categoryForm
              : groupForm
        "
      >
        <template v-if="activeTab === 'agents'">
          <el-form-item
            v-if="dialogMode === 'create'"
            label="登录账号"
            prop="username"
            :rules="[{ required: true, message: '请输入登录账号' }]"
          >
            <el-input v-model="agentForm.username" />
          </el-form-item>
          <el-form-item
            v-if="dialogMode === 'create'"
            label="登录密码"
            prop="password"
            :rules="[{ required: true, message: '请输入登录密码' }]"
          >
            <el-input
              v-model="agentForm.password"
              show-password
              type="password"
            />
          </el-form-item>
          <el-form-item
            v-if="dialogMode === 'create'"
            label="昵称"
            prop="nickname"
            :rules="[{ required: true, message: '请输入昵称' }]"
          >
            <el-input v-model="agentForm.nickname" />
          </el-form-item>
          <el-form-item
            v-if="dialogMode === 'create'"
            label="手机号"
          >
            <el-input v-model="agentForm.mobile" />
          </el-form-item>
          <el-form-item
            v-if="dialogMode === 'create'"
            label="邮箱"
          >
            <el-input v-model="agentForm.email" />
          </el-form-item>
          <el-form-item
            v-if="dialogMode === 'create'"
            label="账号状态"
          >
            <el-segmented
              v-model="agentForm.enabled"
              :options="enabledOptions"
            />
          </el-form-item>
          <el-form-item label="客服分组">
            <el-select
              v-model="agentForm.groupId"
              clearable
              class="full-input"
            >
              <el-option
                v-for="item in groups"
                :key="item.id"
                :label="item.groupName"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item
            label="最大接待"
            prop="maxSessionCount"
            :rules="[{ required: true, message: '请输入最大接待数' }]"
          >
            <el-input-number
              v-model="agentForm.maxSessionCount"
              :min="1"
              :max="999"
              class="full-input"
            />
          </el-form-item>
          <el-form-item label="欢迎语">
            <el-input
              v-model="agentForm.welcomeMessage"
              type="textarea"
              :rows="3"
            />
          </el-form-item>
          <el-form-item label="备注">
            <el-input
              v-model="agentForm.remark"
              type="textarea"
              :rows="2"
            />
          </el-form-item>
        </template>

        <template v-else-if="activeTab === 'categories'">
          <el-form-item
            v-if="dialogMode === 'create'"
            label="分类编码"
            prop="categoryCode"
            :rules="[{ required: true, message: '请输入分类编码' }]"
          >
            <el-input v-model="categoryForm.categoryCode" />
          </el-form-item>
          <el-form-item
            label="分类名称"
            prop="categoryName"
            :rules="[{ required: true, message: '请输入分类名称' }]"
          >
            <el-input v-model="categoryForm.categoryName" />
          </el-form-item>
          <el-form-item label="父级 ID">
            <el-input-number
              v-model="categoryForm.parentId"
              :min="0"
              :controls="false"
              class="full-input"
            />
          </el-form-item>
          <el-form-item label="状态">
            <el-segmented
              v-model="categoryForm.enabled"
              :options="enabledOptions"
            />
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number
              v-model="categoryForm.sort"
              :min="0"
              class="full-input"
            />
          </el-form-item>
          <el-form-item label="备注">
            <el-input
              v-model="categoryForm.remark"
              type="textarea"
              :rows="2"
            />
          </el-form-item>
        </template>

        <template v-else>
          <el-form-item
            v-if="dialogMode === 'create'"
            label="分组编码"
            prop="groupCode"
            :rules="[{ required: true, message: '请输入分组编码' }]"
          >
            <el-input v-model="groupForm.groupCode" />
          </el-form-item>
          <el-form-item
            label="分组名称"
            prop="groupName"
            :rules="[{ required: true, message: '请输入分组名称' }]"
          >
            <el-input v-model="groupForm.groupName" />
          </el-form-item>
          <el-form-item label="描述">
            <el-input
              v-model="groupForm.description"
              type="textarea"
              :rows="2"
            />
          </el-form-item>
          <el-form-item label="状态">
            <el-segmented
              v-model="groupForm.enabled"
              :options="enabledOptions"
            />
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number
              v-model="groupForm.sort"
              :min="0"
              class="full-input"
            />
          </el-form-item>
          <el-form-item label="备注">
            <el-input
              v-model="groupForm.remark"
              type="textarea"
              :rows="2"
            />
          </el-form-item>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">
          取消
        </el-button>
        <el-button
          type="primary"
          @click="submitDialog"
        >
          保存
        </el-button>
      </template>
    </el-dialog>
  </section>
</template>
