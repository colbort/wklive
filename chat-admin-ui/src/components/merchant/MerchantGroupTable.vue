<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox, type FormInstance } from "element-plus";
import {
  createGroup,
  deleteGroup,
  pageGroups,
  updateGroup,
  type ChatGroupPayload,
} from "@/api/chat";
import { useAuthStore } from "@/stores/auth";
import type { ChatGroup } from "@/types/chat";

const auth = useAuthStore();
const merchantId = computed(() => auth.user?.merchantId || 0);

const enabledOptions = [
  { label: "启用", value: 1 },
  { label: "禁用", value: 2 },
];
const loading = ref(false);
const groups = ref<ChatGroup[]>([]);
const keyword = ref("");
const formRef = ref<FormInstance>();
const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const editingId = ref(0);

const groupForm = reactive({
  groupCode: "",
  groupName: "",
  description: "",
  enabled: 1,
  sort: 0,
  remark: "",
});

const dialogTitle = computed(
  () => `${dialogMode.value === "create" ? "新增" : "编辑"}客服分组`,
);

onMounted(() => {
  void loadCurrent();
});

async function loadCurrent(searchKeyword = keyword.value) {
  if (!merchantId.value) return;
  loading.value = true;
  try {
    groups.value = (
      await pageGroups({
        merchantId: merchantId.value,
        limit: 100,
        keyword: searchKeyword || undefined,
      })
    ).data;
  } finally {
    loading.value = false;
  }
}

function resetForm() {
  editingId.value = 0;
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
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: ChatGroup) {
  dialogMode.value = "edit";
  resetForm();
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

  dialogVisible.value = false;
  ElMessage.success("保存成功");
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
  <div
    class="table-actions"
    style="width: 100%;"
  >
    <el-input
      v-model="keyword"
      clearable
      placeholder="搜索坐席"
      size="small"
      @keyup.enter="loadCurrent(keyword)"
      @clear="loadCurrent('')"
    />
    <el-button
      size="small"
      @click="loadCurrent(keyword)"
    >
      搜索
    </el-button>
    <el-button
      type="primary"
      size="small"
      @click="openCreate"
    >
      新增坐席
    </el-button>
  </div>
  <div class="table-panel merchant-group-table-panel">
    <el-table
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
            @click="openEdit(row)"
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
    class="merchant-group-edit-dialog"
    v-model="dialogVisible"
    :title="dialogTitle"
    width="560px"
    destroy-on-close
  >
    <el-form
      ref="formRef"
      label-width="104px"
      :model="groupForm"
    >
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
</template>

<style scoped>
.table-panel {
  display: grid;
  flex: 1 1 auto;
  gap: 12px;
  min-height: 0;
  overflow: hidden;
  border: 1px solid #e6e9ef;
  border-radius: 8px;
  background: #fff;
}

.table-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: nowrap;
}

.table-actions :deep(.el-input) {
  flex: 1;
  min-width: 0;
}

.table-actions :deep(.el-button) {
  flex: none;
}

@media (max-width: 760px) {
  .merchant-group-table-panel {
    width: 100%;
    min-width: 0;
    overflow: hidden;
  }

  .merchant-group-table-panel :deep(.el-table) {
    width: 100% !important;
    min-width: 0 !important;
  }

  .merchant-group-table-panel :deep(.el-table__inner-wrapper),
  .merchant-group-table-panel :deep(.el-table__body-wrapper),
  .merchant-group-table-panel :deep(.el-scrollbar) {
    min-width: 0;
  }
}
</style>

<style>
@media (max-width: 760px) {
  .merchant-group-edit-dialog {
    position: absolute;
    right: 0;
    bottom: 0;
    left: 0;
    display: flex;
    width: 100% !important;
    max-height: 88dvh;
    flex-direction: column;
    margin: 0 !important;
    border-radius: 16px 16px 0 0;
  }

  .merchant-group-edit-dialog .el-dialog__header {
    flex: 0 0 auto;
    margin-right: 0;
    padding: 16px 18px 10px;
  }

  .merchant-group-edit-dialog .el-dialog__body {
    flex: 1 1 auto;
    min-height: 0;
    overflow: auto;
    padding: 8px 18px 12px;
  }

  .merchant-group-edit-dialog .el-dialog__footer {
    flex: 0 0 auto;
    padding: 12px 18px 16px;
    border-top: 1px solid #e6e9ef;
  }

  .merchant-group-edit-dialog .el-form-item {
    margin-bottom: 16px;
  }

  .merchant-group-edit-dialog .el-form-item__label {
    width: 88px !important;
    padding-right: 10px;
  }

  .merchant-group-edit-dialog .el-form-item__content {
    min-width: 0;
  }

  .merchant-group-edit-dialog .el-input-number,
  .merchant-group-edit-dialog .el-segmented {
    width: 100%;
  }

  .merchant-group-edit-dialog .el-dialog__footer .el-button {
    min-width: 96px;
  }
}
</style>
