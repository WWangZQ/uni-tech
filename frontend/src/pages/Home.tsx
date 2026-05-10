import { Link } from 'react-router-dom'

export default function Home() {
  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">欢迎使用智约校园</h1>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Link
          to="/spaces"
          className="bg-white p-6 rounded-lg shadow hover:shadow-lg transition-shadow"
        >
          <h2 className="text-xl font-semibold mb-2">空间预订</h2>
          <p className="text-gray-600">预订会议室、体育场馆等校园空间</p>
        </Link>

        <Link
          to="/activities"
          className="bg-white p-6 rounded-lg shadow hover:shadow-lg transition-shadow"
        >
          <h2 className="text-xl font-semibold mb-2">活动秒杀</h2>
          <p className="text-gray-600">抢订讲座、演唱会、比赛等热门活动门票</p>
        </Link>

        <Link
          to="/orders"
          className="bg-white p-6 rounded-lg shadow hover:shadow-lg transition-shadow"
        >
          <h2 className="text-xl font-semibold mb-2">我的订单</h2>
          <p className="text-gray-600">查看和管理您的所有订单</p>
        </Link>
      </div>
    </div>
  )
}
