<script setup lang="ts">
import { computed, ref } from "vue";
import type { FormInstance } from "element-plus";
import type { ChatGroup } from "@/types/chat";

type TabName = "agents" | "categories" | "groups";
type DialogMode = "create" | "edit";

interface EnabledOption {
  label: string;
  value: number;
}

interface AgentForm {
  username: string;
  password: string;
  nickname: string;
  mobile: string;
  email: string;
  enabled: number;
  maxSessionCount: number;
  groupId?: number;
  welcomeMessage: string;
  remark: string;
}

interface CategoryForm {
  parentId: number;
  categoryCode: string;
  categoryName: string;
  enabled: number;
  sort: number;
  remark: string;
}

interface GroupForm {
  groupCode: string;
  groupName: string;
  description: string;
  enabled: number;
  sort: number;
  remark: string;
}

const props = defineProps<{
  visible: boolean;
  title: string;
  activeTab: TabName;
  dialogMode: DialogMode;
  agentForm: AgentForm;
  categoryForm: CategoryForm;
  groupForm: GroupForm;
  groups: ChatGroup[];
  enabledOptions: EnabledOption[];
}>();

const emit = defineEmits<{
  submit: [];
  "update:visible": [visible: boolean];
}>();

const formRef = ref<FormInstance>();
const formModel = computed(() => {
  if (props.activeTab === "agents") return props.agentForm;
  if (props.activeTab === "categories") return props.categoryForm;
  return props.groupForm;
});

function validate() {
  return formRef.value?.validate();
}

function clearValidate() {
  formRef.value?.clearValidate();
}

defineExpose({
  validate,
  clearValidate,
});
</script>

<template>
  <el-dialog
    :model-value="visible"
    :title="title"
    width="560px"
    destroy-on-close
    @update:model-value="emit('update:visible', $event)"
  >
    <el-form
      ref="formRef"
      label-width="104px"
      :model="formModel"
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
      <el-button @click="emit('update:visible', false)">
        取消
      </el-button>
      <el-button
        type="primary"
        @click="emit('submit')"
      >
        保存
      </el-button>
    </template>
  </el-dialog>
</template>
