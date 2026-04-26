<template>
  <div class="order-detail">
    <el-page-header @back="goBack" title="订单详情" />

    <el-card class="detail-card" v-loading="loading">
      <!-- 订单基本信息 -->
      <div class="section">
        <h3 class="section-title">基本信息</h3>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="订单号">{{ order?.orderNo }}</el-descriptions-item>
          <el-descriptions-item label="用户ID">{{ order?.userId }}</el-descriptions-item>
          <el-descriptions-item label="订单状态">
            <el-tag :type="getStatusType(order?.status)">
              {{ getStatusText(order?.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="支付状态">
            <el-tag :type="order?.payStatus === 1 ? 'success' : 'info'">
              {{ order?.payStatus === 1 ? '已支付' : '未支付' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="支付方式">{{ order?.payMethod || '-' }}</el-descriptions-item>
          <el-descriptions-item label="支付时间">
            {{ order?.payTime ? formatDate(order.payTime) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="订单金额" :span="2">
            <span class="amount">¥{{ order?.totalAmount.toFixed(2) }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ order?.createdAt ? formatDate(order.createdAt) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="备注" :span="3">{{ order?.remark || '-' }}</el-descriptions-item>
        </el-descriptions>
      </div>

      <!-- 商品明细 -->
      <div class="section">
        <h3 class="section-title">商品明细</h3>
        <el-table :data="order?.items" stripe>
          <el-table-column prop="productId" label="商品ID" width="100" />
          <el-table-column prop="productName" label="商品名称" />
          <el-table-column prop="unitPrice" label="单价" width="120">
            <template #default="{ row }">
              ¥{{ row.unitPrice.toFixed(2) }}
            </template>
          </el-table-column>
          <el-table-column prop="quantity" label="数量" width="100" />
          <el-table-column prop="totalPrice" label="小计" width="120">
            <template #default="{ row }">
              ¥{{ row.totalPrice.toFixed(2) }}
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getOrderDetail } from '@/api/orders'
import type { Order, OrderStatus } from '@/types/api'

const route = useRoute()
const router = useRouter()

const order = ref<Order | null>(null)
const loading = ref(false)

const fetchOrderDetail = async () => {
  const id = route.params.id as string
  if (!id) return

  loading.value = true
  try {
    const response = await getOrderDetail(id)
    order.value = response.data.data
  } catch (error) {
    ElMessage.error('获取订单详情失败')
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.back()
}

const getStatusType = (status?: OrderStatus): string => {
  if (status === undefined) return 'info'
  const map: Record<number, string> = {
    0: 'warning',
    1: 'primary',
    2: 'info',
    3: 'info',
    4: 'success',
    5: 'danger'
  }
  return map[status] || 'info'
}

const getStatusText = (status?: OrderStatus): string => {
  if (status === undefined) return '未知'
  const map: Record<number, string> = {
    0: '待支付',
    1: '已支付',
    2: '处理中',
    3: '已发货',
    4: '已完成',
    5: '已取消'
  }
  return map[status] || '未知'
}

const formatDate = (dateStr: string): string => {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

onMounted(() => {
  fetchOrderDetail()
})
</script>

<style scoped>
.order-detail {
  padding: 0;
}

.detail-card {
  margin-top: 20px;
}

.section {
  margin-bottom: 30px;
}

.section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 15px;
  padding-left: 10px;
  border-left: 4px solid #409EFF;
}

.amount {
  color: #F56C6C;
  font-weight: bold;
  font-size: 16px;
}
</style>
