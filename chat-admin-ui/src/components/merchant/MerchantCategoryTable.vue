<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox, type FormInstance } from "element-plus";
import {
  createCategory,
  deleteCategory,
  pageCategories,
  updateCategory,
  type ChatCategoryPayload,
} from "@/api/chat";
import { useAuthStore } from "@/stores/auth";
import type { ChatCategory } from "@/types/chat";

const auth = useAuthStore();
const merchantId = computed(() => auth.user?.merchantId || 0);

const enabledOptions = [
  { label: "启用", value: 1 },
  { label: "禁用", value: 2 },
];
const loading = ref(false);
const categories = ref<ChatCategory[]>([]);
const keyword = ref("");
const formRef = ref<FormInstance>();
const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const editingId = ref(0);

const categoryForm = reactive({
  parentId: 0,
  categoryCode: "",
  categoryName: "",
  enabled: 1,
  sort: 0,
  remark: "",
});

const dialogTitle = computed(
  () => `${dialogMode.value === "create" ? "新增" : "编辑"}问题分类`,
);

onMounted(() => {
  void loadCurrent();
});

async function loadCurrent(searchKeyword = keyword.value) {
  if (!merchantId.value) return;
  loading.value = true;
  try {
    categories.value = (
      await pageCategories({
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
  Object.assign(categoryForm, {
    parentId: 0,
    categoryCode: "",
    categoryName: "",
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

function openEdit(row: ChatCategory) {
  dialogMode.value = "edit";
  resetForm();
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

async function submitDialog() {
  if (!merchantId.value) return;
  await formRef.value?.validate();

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

  dialogVisible.value = false;
  ElMessage.success("保存成功");
  await loadCurrent();
}

async function removeCategory(row: ChatCategory) {
  await ElMessageBox.confirm(
    `确认删除分类「${row.categoryName}」？`,
    "删除确认",
    {
      type: "warning",
    },
  );
  await deleteCategory(row.id, merchantId.value);
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
  <div class="table-panel merchant-category-table-panel">
    <el-table
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
            @click="openEdit(row)"
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
  </div>

  <el-dialog
    class="merchant-category-edit-dialog"
    v-model="dialogVisible"
    :title="dialogTitle"
    width="560px"
    destroy-on-close
  >
    <el-form
      ref="formRef"
      label-width="104px"
      :model="categoryForm"
    >
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
  .merchant-category-table-panel {
    width: 100%;
    min-width: 0;
    overflow: hidden;
  }

  .merchant-category-table-panel :deep(.el-table) {
    width: 100% !important;
    min-width: 0 !important;
  }

  .merchant-category-table-panel :deep(.el-table__inner-wrapper),
  .merchant-category-table-panel :deep(.el-table__body-wrapper),
  .merchant-category-table-panel :deep(.el-scrollbar) {
    min-width: 0;
  }
}
</style>

<style>
@media (max-width: 760px) {
  .merchant-category-edit-dialog {
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

  .merchant-category-edit-dialog .el-dialog__header {
    flex: 0 0 auto;
    margin-right: 0;
    padding: 16px 18px 10px;
  }

  .merchant-category-edit-dialog .el-dialog__body {
    flex: 1 1 auto;
    min-height: 0;
    overflow: auto;
    padding: 8px 18px 12px;
  }

  .merchant-category-edit-dialog .el-dialog__footer {
    flex: 0 0 auto;
    padding: 12px 18px 16px;
    border-top: 1px solid #e6e9ef;
  }

  .merchant-category-edit-dialog .el-form-item {
    margin-bottom: 16px;
  }

  .merchant-category-edit-dialog .el-form-item__label {
    width: 88px !important;
    padding-right: 10px;
  }

  .merchant-category-edit-dialog .el-form-item__content {
    min-width: 0;
  }

  .merchant-category-edit-dialog .el-input-number,
  .merchant-category-edit-dialog .el-segmented {
    width: 100%;
  }

  .merchant-category-edit-dialog .el-dialog__footer .el-button {
    min-width: 96px;
  }
}
</style>
