<template>
  <div class="products-page">
    <el-card>
      <!-- 搜索栏 -->
      <el-form :model="searchForm" inline class="search-form">
        <el-form-item label="商品名称">
          <el-input v-model="searchForm.productName" placeholder="请输入商品名称" clearable />
        </el-form-item>

        <el-form-item label="商品状态">
          <el-select v-model="searchForm.status" placeholder="全部状态" clearable style="width: 200px">
            <el-option label="下架" :value="0" />
            <el-option label="上架" :value="1" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>搜索
          </el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 操作栏 -->
      <div class="toolbar">
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>新增商品
        </el-button>
      </div>

      <!-- 数据表格 -->
      <el-table :data="productList" v-loading="loading" stripe>
        <el-table-column prop="product_id" label="商品ID" width="100" />
        <el-table-column prop="product_name" label="商品名称" />
        <el-table-column prop="original_price" label="原价" width="140">
          <template #default="{ row }">
            ¥{{ (row.original_price / 100).toFixed(2) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '上架' : '下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="200">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleUpdatePrice(row)">
              改价
            </el-button>
            <el-button link :type="row.status === 1 ? 'danger' : 'success'" @click="handleToggleStatus(row)">
              {{ row.status === 1 ? '下架' : '上架' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 新增商品对话框 -->
    <el-dialog v-model="createDialogVisible" title="新增商品" width="500px">
      <el-form :model="createForm" label-width="100px" :rules="createRules" ref="createFormRef">
        <el-form-item label="商品名称" prop="product_name">
          <el-input v-model="createForm.product_name" placeholder="请输入商品名称" />
        </el-form-item>
        <el-form-item label="原价" prop="original_price">
          <el-input-number v-model="createForm.original_price" :min="0" :precision="2" :step="0.01" style="width: 200px" />
          <span class="form-tip">元</span>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="createForm.status">
            <el-radio :label="1">上架</el-radio>
            <el-radio :label="0">下架</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitCreate" :loading="submitLoading">确定</el-button>
      </template>
    </el-dialog>

    <!-- 改价对话框 -->
    <el-dialog v-model="priceDialogVisible" title="修改价格" width="400px">
      <el-form :model="priceForm" label-width="80px">
        <el-form-item label="商品名称">
          <span>{{ priceForm.product_name }}</span>
        </el-form-item>
        <el-form-item label="新价格">
          <el-input-number v-model="priceForm.original_price" :min="0" :precision="2" :step="0.01" style="width: 200px" />
          <span class="form-tip">元</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="priceDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitUpdatePrice" :loading="submitLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getProducts, createProduct, updateProductPrice, updateProductStatus } from '@/api/products'
import type { Product } from '@/types/api'
import type { FormInstance, FormRules } from 'element-plus'

// 搜索表单
const searchForm = reactive({
  productName: '',
  status: undefined as number | undefined
})

// 表格数据
const productList = ref<Product[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)

// 对话框状态
const createDialogVisible = ref(false)
const priceDialogVisible = ref(false)
const submitLoading = ref(false)
const createFormRef = ref<FormInstance>()

// 创建表单
const createForm = reactive({
  product_name: '',
  original_price: 0,
  status: 1
})

// 创建表单校验规则
const createRules: FormRules = {
  product_name: [
    { required: true, message: '请输入商品名称', trigger: 'blur' },
    { min: 1, max: 100, message: '长度在 1 到 100 个字符', trigger: 'blur' }
  ],
  original_price: [
    { required: true, message: '请输入原价', trigger: 'blur' }
  ],
  status: [
    { required: true, message: '请选择状态', trigger: 'change' }
  ]
}

// 改价表单
const priceForm = reactive({
  product_id: 0,
  product_name: '',
  original_price: 0
})

// 获取商品列表
const fetchProducts = async () => {
  loading.value = true
  try {
    const response = await getProducts({
      page: page.value,
      pageSize: pageSize.value,
      product_name: searchForm.productName || undefined,
      status: searchForm.status
    })
    productList.value = response.data.data.list
    total.value = response.data.data.total
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  page.value = 1
  fetchProducts()
}

// 重置
const handleReset = () => {
  searchForm.productName = ''
  searchForm.status = undefined
  page.value = 1
  fetchProducts()
}

// 分页
const handleSizeChange = (val: number) => {
  pageSize.value = val
  fetchProducts()
}

const handleCurrentChange = (val: number) => {
  page.value = val
  fetchProducts()
}

// 打开新增对话框
const handleCreate = () => {
  createForm.product_name = ''
  createForm.original_price = 0
  createForm.status = 1
  createDialogVisible.value = true
}

// 提交新增
const submitCreate = async () => {
  if (!createFormRef.value) return

  await createFormRef.value.validate(async (valid) => {
    if (!valid) return

    submitLoading.value = true
    try {
      // 将元转换为分
      const data = {
        ...createForm,
        original_price: Math.round(createForm.original_price * 100)
      }
      await createProduct(data)
      ElMessage.success('创建成功')
      createDialogVisible.value = false
      fetchProducts()
    } finally {
      submitLoading.value = false
    }
  })
}

// 打开改价对话框
const handleUpdatePrice = (row: Product) => {
  priceForm.product_id = row.product_id
  priceForm.product_name = row.product_name
  priceForm.original_price = row.original_price / 100
  priceDialogVisible.value = true
}

// 提交改价
const submitUpdatePrice = async () => {
  submitLoading.value = true
  try {
    // 将元转换为分
    const priceInCents = Math.round(priceForm.original_price * 100)
    await updateProductPrice(priceForm.product_id, priceInCents)
    ElMessage.success('价格修改成功')
    priceDialogVisible.value = false
    fetchProducts()
  } finally {
    submitLoading.value = false
  }
}

// 切换上下架状态
const handleToggleStatus = async (row: Product) => {
  const newStatus = row.status === 1 ? 0 : 1
  const actionText = newStatus === 1 ? '上架' : '下架'

  try {
    await ElMessageBox.confirm(`确认${actionText}商品 "${row.product_name}"？`, '提示', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning'
    })

    await updateProductStatus(row.product_id, newStatus)
    ElMessage.success(`${actionText}成功`)
    fetchProducts()
  } catch {
    // 用户取消
  }
}

const formatDate = (dateStr: string): string => {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

onMounted(() => {
  fetchProducts()
})
</script>

<style scoped>
.search-form {
  margin-bottom: 20px;
}

.search-form :deep(.el-form-item__content) {
  min-width: 140px;
}

.search-form :deep(.el-select__wrapper) {
  min-width: 140px;
}

.toolbar {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.form-tip {
  margin-left: 8px;
  color: #666;
}
</style>
