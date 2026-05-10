import { useState, useEffect } from 'react'
import { activityApi } from '../services/api'
import dayjs from 'dayjs'

interface Activity {
  id: number
  title: string
  description: string
  location: string
  activity_type: string
  start_time: string
  end_time: string
  total_tickets: number
  remaining_tickets: number
  price: number
  status: string
  cover_image?: string
}

export default function Activities() {
  const [activities, setActivities] = useState<Activity[]>([])
  const [loading, setLoading] = useState(true)
  const [seckilling, setSeckilling] = useState<number | null>(null)
  const [message, setMessage] = useState('')

  useEffect(() => {
    loadActivities()
  }, [])

  const loadActivities = async () => {
    try {
      const response = await activityApi.getActivities()
      setActivities(response.data.activities)
    } catch (error) {
      console.error('Failed to load activities:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleSeckill = async (activityId: number) => {
    setSeckilling(activityId)
    setMessage('')

    try {
      const response = await activityApi.doSeckill(activityId)
      setMessage(response.data.message || '秒杀成功！')
      loadActivities()
    } catch (error: any) {
      setMessage(error.response?.data?.error || '秒杀失败')
    } finally {
      setSeckilling(null)
    }
  }

  const getStatusText = (status: string) => {
    const map: Record<string, string> = {
      draft: '未开始',
      seckill: '秒杀中',
      ongoing: '进行中',
      ended: '已结束',
      cancelled: '已取消',
    }
    return map[status] || status
  }

  const getStatusColor = (status: string) => {
    const map: Record<string, string> = {
      draft: 'bg-gray-100 text-gray-800',
      seckill: 'bg-red-100 text-red-800',
      ongoing: 'bg-green-100 text-green-800',
      ended: 'bg-gray-100 text-gray-500',
      cancelled: 'bg-gray-100 text-gray-500',
    }
    return map[status] || 'bg-gray-100 text-gray-800'
  }

  if (loading) {
    return <div className="text-center py-8">加载中...</div>
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">活动秒杀</h1>

      {message && (
        <div className="mb-6 p-4 rounded bg-blue-100 text-blue-800">
          {message}
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {activities.map((activity) => (
          <div key={activity.id} className="bg-white rounded-lg shadow overflow-hidden">
            {activity.cover_image && (
              <img
                src={activity.cover_image}
                alt={activity.title}
                className="w-full h-40 object-cover"
              />
            )}
            <div className="p-4">
              <div className="flex justify-between items-start mb-2">
                <h3 className="text-lg font-semibold">{activity.title}</h3>
                <span className={`px-2 py-1 text-xs rounded ${getStatusColor(activity.status)}`}>
                  {getStatusText(activity.status)}
                </span>
              </div>

              <p className="text-gray-500 text-sm mb-2">
                {activity.location}
              </p>

              <p className="text-gray-600 text-sm mb-4 line-clamp-2">
                {activity.description}
              </p>

              <div className="text-sm text-gray-500 mb-4">
                <p>时间: {dayjs(activity.start_time).format('YYYY-MM-DD HH:mm')}</p>
                <p>票价: ¥{activity.price.toFixed(2)}</p>
                <p className="text-red-600 font-medium">
                  剩余: {activity.remaining_tickets} / {activity.total_tickets}
                </p>
              </div>

              <button
                onClick={() => handleSeckill(activity.id)}
                disabled={
                  seckilling === activity.id ||
                  activity.status !== 'seckill' ||
                  activity.remaining_tickets <= 0
                }
                className={`w-full py-2 px-4 rounded-md font-medium ${
                  activity.status === 'seckill' && activity.remaining_tickets > 0
                    ? 'bg-red-500 text-white hover:bg-red-600'
                    : 'bg-gray-300 text-gray-500 cursor-not-allowed'
                }`}
              >
                {seckilling === activity.id
                  ? '秒杀中...'
                  : activity.status !== 'seckill'
                  ? '未开始'
                  : activity.remaining_tickets <= 0
                  ? '已售罄'
                  : '立即秒杀'}
              </button>
            </div>
          </div>
        ))}
      </div>

      {activities.length === 0 && (
        <div className="text-center py-8 text-gray-500">暂无活动</div>
      )}
    </div>
  )
}
