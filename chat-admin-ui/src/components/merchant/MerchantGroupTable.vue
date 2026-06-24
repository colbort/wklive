<script setup lang="ts">
import { ref } from "vue";
import type { FormInstance } from "element-plus";
import type { ChatGroup } from "@/types/chat";

interface EnabledOption {
  label: string;
  value: number;
}

interface GroupForm {
  groupCode: string;
  groupName: string;
  description: string;
  enabled: number;
  sort: number;
  remark: string;
}

defineProps<{
  loading: boolean;
  groups: ChatGroup[];
  visible: boolean;
  title: string;
  dialogMode: "create" | "edit";
  groupForm: GroupForm;
  enabledOptions: EnabledOption[];
}>();

const emit = defineEmits<{
  edit: [row: ChatGroup];
  remove: [row: ChatGroup];
  search: [keyword?: string];
  create: [];
  submit: [];
  "update:visible": [visible: boolean];
}>();

const keyword = ref("");
const formRef = ref<FormInstance>();

async function submitDialog() {
  await formRef.value?.validate();
  emit("submit");
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
      @keyup.enter="emit('search', keyword)"
      @clear="emit('search')"
    />
    <el-button
      size="small"
      @click="emit('search', keyword)"
    >
      搜索
    </el-button>
    <el-button
      type="primary"
      size="small"
      @click="emit('create')"
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
            @click="emit('edit', row)"
          >
            编辑
          </el-button>
          <el-button
            link
            type="danger"
            @click="emit('remove', row)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>

  <el-dialog
    class="merchant-group-edit-dialog"
    :model-value="visible"
    :title="title"
    width="560px"
    destroy-on-close
    @update:model-value="emit('update:visible', $event)"
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
      <el-button @click="emit('update:visible', false)">
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
