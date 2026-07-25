<template>
  <div class="profile-stage">
    <div class="profile-card">
      <div class="head">
        <h1>个人资料</h1>
        <p class="sub">完善你的展示信息，仅你本人可编辑</p>
      </div>

      <el-form :model="form" :rules="rules" ref="formRef" label-position="top" @submit.prevent="onSubmit">
        <el-form-item label="账号">
          <el-input :model-value="auth.user?.username" disabled />
        </el-form-item>

        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="form.nickname" placeholder="展示名（必填）" size="large" />
        </el-form-item>

        <el-form-item label="头像">
          <div class="avatar-row">
            <div
              v-for="p in presets"
              :key="p.id"
              class="preset"
              :class="{ active: form.avatar === p.id }"
              :style="{ background: `linear-gradient(135deg, ${p.from}, ${p.to})` }"
              @click="form.avatar = p.id"
            >{{ initial }}</div>
            <div v-if="isImageRef(form.avatar)" class="preset url">
              <img :src="form.avatar" alt="avatar" />
            </div>
          </div>
          <div class="avatar-actions">
            <el-button size="small" :loading="uploadLoading" @click="pickFile">上传图片</el-button>
            <input ref="fileInput" type="file" accept="image/png,image/jpeg,image/webp" hidden @change="onFile" />
            <span v-if="uploadErr" class="upload-err">{{ uploadErr }}</span>
          </div>
          <el-input v-model="form.avatar" placeholder="或粘贴头像图片 URL（留空则用上方预设 / 首字母）" size="large" class="avatar-input" />
        </el-form-item>

        <el-form-item label="联系方式" prop="contact">
          <el-input v-model="form.contact" placeholder="电话 / 微信 / 邮箱（必填）" size="large" />
        </el-form-item>

        <el-form-item label="个人简介">
          <el-input v-model="form.bio" type="textarea" :rows="3" maxlength="500" show-word-limit placeholder="一句话介绍自己" />
        </el-form-item>

        <el-button type="primary" size="large" class="submit" :loading="loading" @click="onSubmit">保存资料</el-button>
        <p v-if="err" class="err">{{ err }}</p>
        <p v-if="ok" class="ok">{{ ok }}</p>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { uploadAvatar } from '../api/request'

const auth = useAuthStore()
const formRef = ref(null)
const loading = ref(false)
const err = ref('')
const ok = ref('')

const presets = [
  { id: 'preset:blue', from: '#5EA1FF', to: '#A07BFF' },
  { id: 'preset:cyan', from: '#3DD9EB', to: '#5EA1FF' },
  { id: 'preset:sun', from: '#FFB547', to: '#FF6B6B' }
]

const initial = computed(() => {
  const u = auth.user?.nickname || auth.user?.username || ''
  return u ? u.charAt(0).toUpperCase() : '?'
})
const isImageRef = (v) =>
  typeof v === 'string' &&
  (/^https?:\/\//.test(v) || v.startsWith('/avatars/') || v.startsWith('data:'))

const fileInput = ref(null)
const uploadLoading = ref(false)
const uploadErr = ref('')

function pickFile() {
  fileInput.value?.click()
}

async function onFile(e) {
  const f = e.target.files?.[0]
  if (!f) return
  uploadErr.value = ''
  if (!/image\/(png|jpeg|webp)/.test(f.type)) {
    uploadErr.value = '仅支持 PNG / JPG / WEBP'
    e.target.value = ''
    return
  }
  if (f.size > 2 << 20) {
    uploadErr.value = '图片不能超过 2MB'
    e.target.value = ''
    return
  }
  uploadLoading.value = true
  try {
    const res = await uploadAvatar(f)
    if (!res || res.code !== 0) throw new Error((res && res.message) || '上传失败')
    form.avatar = res.data.url
  } catch (err) {
    uploadErr.value = (err && err.message) || '上传失败'
  } finally {
    uploadLoading.value = false
    e.target.value = ''
  }
}

const form = reactive({ nickname: '', avatar: '', contact: '', bio: '' })

const rules = {
  nickname: [{ required: true, message: '昵称为必填项', trigger: 'blur' }],
  contact: [{ required: true, message: '联系方式为必填项', trigger: 'blur' }]
}

onMounted(async () => {
  await auth.fetchProfile()
  const u = auth.user || {}
  form.nickname = u.nickname || ''
  form.avatar = u.avatar || ''
  form.contact = u.contact || ''
  form.bio = u.bio || ''
})

async function onSubmit() {
  err.value = ''
  ok.value = ''
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch (e) {
    return // 校验未过，不提交
  }
  loading.value = true
  try {
    await auth.saveProfile({
      nickname: form.nickname.trim(),
      avatar: form.avatar.trim(),
      contact: form.contact.trim(),
      bio: form.bio.trim()
    })
    ok.value = '资料已保存'
    ElMessage.success('资料已保存')
  } catch (e) {
    const msg = (e && e.message) || '保存失败'
    err.value = /status code 500|Network Error|timeout/i.test(msg)
      ? '保存失败：无法连接服务器，请确认 Go 后端(:8080)已启动'
      : msg
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.profile-stage {
  min-height: 100vh;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 48px 16px;
  background: radial-gradient(1200px 600px at 50% -10%, #16203c 0%, #0b1020 60%);
}
.profile-card {
  width: 460px;
  max-width: 100%;
  padding: 32px 30px;
  border-radius: 16px;
  background: rgba(19, 26, 48, 0.72);
  border: 1px solid rgba(61, 217, 235, 0.18);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(14px);
}
.head { text-align: center; margin-bottom: 20px; }
.head h1 { font-size: 18px; margin: 0 0 4px; color: #eaf2ff; }
.head .sub { font-size: 13px; color: #8fa3c8; margin: 0; }
.avatar-row { display: flex; gap: 10px; margin-bottom: 10px; align-items: center; }
.preset {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: #fff;
  font-weight: 700;
  cursor: pointer;
  border: 2px solid transparent;
  &.active {
    border-color: #3DD9EB;
    box-shadow: 0 0 0 3px rgba(61, 217, 235, 0.25);
  }
  &.url {
    padding: 0;
    overflow: hidden;
    img { width: 100%; height: 100%; border-radius: 50%; object-fit: cover; }
  }
}
.avatar-input { margin-bottom: 4px; }
.avatar-actions { display: flex; align-items: center; gap: 10px; margin-bottom: 10px; }
.upload-err { color: #ff6b6b; font-size: 12px; }
.submit { width: 100%; margin-top: 8px; }
.err { color: #ff6b6b; font-size: 13px; text-align: center; margin: 12px 0 0; }
.ok { color: #7DD96E; font-size: 13px; text-align: center; margin: 12px 0 0; }
</style>
