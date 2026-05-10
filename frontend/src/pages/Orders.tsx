import { useState, useEffect } from 'react'
import { orderApi } from '../services/api'
import dayjs from 'dayjs'

interface Order {
  id: number
  order_no: string
  order_type: string
  status: string
  total_amount: number
  payment_deadline?: string
  paid_at?: string
  cancelled_at?: string
  cancel_reason?: string
  created_at: string
  items: OrderItem[]
}

interface OrderItem {
  id: number
  resource_id?: number
  activity_id?: number
  ticket_count: number
  unit_price: number
  slot_date?: string
  start_time?: string
  end_time?: string
}

export default function Orders() {
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState<number | null>(null)
  const [message, setMessage] = useState('')

  useEffect(() => {
    loadOrders()
  }, [])

  const loadOrders = async () => {
    try {
      const response = await orderApi.getOrders()
      setOrders(response.data.orders)
    } catch (error) {
      console.error('Failed to load orders:', error)
    } finally {
      setLoading(false)
    }
  }

  const handlePay = async (orderId: number) => {
    setActionLoading(orderId)
    setMessage('')

    try {
      await orderApi.payOrder(orderId)
      setMessage('支付成功！')
      loadOrders()
    } catch (error: any) {
      setMessage(error.response?.data?.error || '支付失败')
    } finally {
      setActionLoading(null)
    }
  }

  const handleCancel = async (orderId: number) => {
    setActionLoading(orderId)
    setMessage('')

    try {
      await orderApi.cancelOrder(orderId)
      setMessage('取消成功！')
      loadOrders()
    } catch (error: any) {
      setMessage(error.response?.data?.error || '取消失败')
    } finally {
      setActionLoading(null)
    }
  }

  const getStatusText = (status: string) => {
    const map: Record<string, string> = {
      pending: '待支付',
      confirmed: '已确认',
      paid: '已支付',
      cancelled: '已取消',
      no_show: '已爽约',
      completed: '已完成',
    }
    return map[status] || status
  }

  const getStatusColor = (status: string) => {
    const map: Record<string, string> = {
      pending: 'bg-yellow-100 text-yellow-800',
      confirmed: 'bg-blue-100 text-blue-800',
      paid: 'bg-green-100 text-green-800',
      cancelled: 'bg-gray-100 text-gray-800',
      no_show: 'bg-red-100 text-red-800',
      completed: 'bg-gray-100 text-gray-500',
    }
    return map[status] || 'bg-gray-100 text-gray-800'
  }

  if (loading) {
    return <div className="text-center py-8">加载中...</div>
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">我的订单</h1>

      {message && (
        <div className="mb-6 p-4 rounded bg-blue-100 text-blue-800">
          {message}
        </div>
      )}

      <div className="space-y-4">
        {orders.map((order) => (
          <div key={order.id} className="bg-white rounded-lg shadow p-4">
            <div className="flex justify-between items-start mb-4">
              <div>
                <p className="text-sm text-gray-500">订单号: {order.order_no}</p>
                <p className="text-sm text-gray-500">
                  创建时间: {dayjs(order.created_at).format('YYYY-MM-DD HH:mm:ss')}
                </p>
              </div>
              <span className={`px-3 py-1 text-sm rounded ${getStatusColor(order.status)}`}>
                {getStatusText(order.status)}
              </span>
            </div>

            <div className="border-t pt-4">
              {order.items.map((item) => (
                <div key={item.id} className="mb-2">
                  {item.resource_id && (
                    <p className="text-gray-700">
                      空间预订: {item.slot_date} {item.start_time} - {item.end_time}
                    </p>
                  )}
                  {item.activity_id && (
                    <p className="text-gray-700">活动门票 x {item.ticket_count}</p>
                  )}
                  <p className="text-gray-500 text-sm">
                    单价: ¥{item.unit_price.toFixed(2)} × {item.ticket_count}
                  </p>
                </div>
              ))}
            </div>

            <div className="border-t pt-4 flex justify-between items-center">
              <div className="text-lg font-semibold">
                总计: ¥{order.total_amount.toFixed(2)}
              </div>

              <div className="flex space-x-2">
                {order.status === 'pending' && (
                  <>
                    <button
                      onClick={() => handlePay(order.id)}
                      disabled={actionLoading === order.id}
                      className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600 disabled:opacity-50"
                    >
                      {actionLoading === order.id ? '支付中...' : '立即支付'}
                    </button>
                    <button
                      onClick={() => handleCancel(order.id)}
                      disabled={actionLoading === order.id}
                      className="bg-gray-300 text-gray-700 px-4 py-2 rounded hover:bg-gray-400 disabled:opacity-50"
                    >
                      取消订单
                    </button>
                  </>
                )}

                {order.status === 'paid' && order.payment_deadline && (
                  <p className="text-sm text-gray-500">
                    支付时间: {dayjs(order.paid_at).format('YYYY-MM-DD HH:mm:ss')}
                  </p>
                )}

                {order.status === 'cancelled' && order.cancel_reason && (
                  <p className="text-sm text-red-500">取消原因: {order.cancel_reason}</p>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      {orders.length === 0 && (
        <div className="text-center py-8 text-gray-500">暂无订单</div>
      )}
    </div>
  )
}
