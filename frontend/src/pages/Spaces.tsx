import { useState, useEffect } from 'react'
import { spaceApi } from '../services/api'

interface Space {
  id: number
  type: 'academic' | 'sports'
  name: string
  code: string
  capacity: number
  location: string
  status: string
}

interface TimeSlot {
  id: number
  date: string
  start_time: string
  end_time: string
  status: string
}

export default function Spaces() {
  const [spaces, setSpaces] = useState<Space[]>([])
  const [, setLoading] = useState(true)
  const [selectedSpace, setSelectedSpace] = useState<Space | null>(null)
  const [selectedDate, setSelectedDate] = useState('')
  const [slots, setSlots] = useState<TimeSlot[]>([])
  const [selectedSlot, setSelectedSlot] = useState<TimeSlot | null>(null)
  const [booking, setBooking] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    loadSpaces()
  }, [])

  const loadSpaces = async () => {
    try {
      const response = await spaceApi.getSpaces()
      setSpaces(response.data.spaces)
    } catch (error) {
      console.error('Failed to load spaces:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadSlots = async (spaceId: number, date: string) => {
    try {
      const response = await spaceApi.getSlots(spaceId, date)
      setSlots(response.data.slots)
    } catch (error) {
      console.error('Failed to load slots:', error)
    }
  }

  const handleSpaceClick = (space: Space) => {
    setSelectedSpace(space)
    setSelectedSlot(null)
    setMessage('')
    if (selectedDate) {
      loadSlots(space.id, selectedDate)
    }
  }

  const handleDateChange = (date: string) => {
    setSelectedDate(date)
    setSelectedSlot(null)
    setMessage('')
    if (selectedSpace) {
      loadSlots(selectedSpace.id, date)
    }
  }

  const handleBooking = async () => {
    if (!selectedSpace || !selectedSlot || !selectedDate) return

    setBooking(true)
    setMessage('')

    try {
      const response = await spaceApi.createBooking({
        resource_id: selectedSpace.id,
        slot_id: selectedSlot.id,
        date: selectedDate,
      })
      setMessage(response.data.message || '预订成功！')
      setSelectedSlot(null)
      loadSlots(selectedSpace.id, selectedDate)
    } catch (error: any) {
      setMessage(error.response?.data?.error || '预订失败')
    } finally {
      setBooking(false)
    }
  }

  const today = new Date().toISOString().split('T')[0]

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">空间预订</h1>

      {/* Space List */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        {spaces.map((space) => (
          <div
            key={space.id}
            onClick={() => handleSpaceClick(space)}
            className={`bg-white p-6 rounded-lg shadow cursor-pointer transition-all ${
              selectedSpace?.id === space.id ? 'ring-2 ring-primary-500' : ''
            }`}
          >
            <div className="flex justify-between items-start mb-2">
              <h3 className="text-lg font-semibold">{space.name}</h3>
              <span
                className={`px-2 py-1 text-xs rounded ${
                  space.type === 'academic'
                    ? 'bg-blue-100 text-blue-800'
                    : 'bg-green-100 text-green-800'
                }`}
              >
                {space.type === 'academic' ? '学术空间' : '体育设施'}
              </span>
            </div>
            <p className="text-gray-500 text-sm mb-2">编码: {space.code}</p>
            <p className="text-gray-500 text-sm">容量: {space.capacity}人</p>
            <p className="text-gray-500 text-sm">位置: {space.location}</p>
          </div>
        ))}
      </div>

      {/* Booking Panel */}
      {selectedSpace && (
        <div className="bg-white p-6 rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">
            预订: {selectedSpace.name}
          </h2>

          <div className="mb-4">
            <label htmlFor="date" className="block text-sm font-medium text-gray-700 mb-1">
              选择日期
            </label>
            <input
              type="date"
              id="date"
              min={today}
              value={selectedDate}
              onChange={(e) => handleDateChange(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
            />
          </div>

          {selectedDate && (
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                选择时间
              </label>
              <div className="grid grid-cols-4 gap-2">
                {slots.map((slot) => (
                  <button
                    key={slot.id}
                    onClick={() => setSelectedSlot(slot)}
                    disabled={slot.status !== 'available'}
                    className={`p-2 text-sm rounded ${
                      selectedSlot?.id === slot.id
                        ? 'bg-primary-500 text-white'
                        : slot.status === 'available'
                        ? 'bg-gray-100 hover:bg-gray-200'
                        : 'bg-gray-300 text-gray-500 cursor-not-allowed'
                    }`}
                  >
                    {slot.start_time} - {slot.end_time}
                  </button>
                ))}
              </div>
            </div>
          )}

          {selectedSlot && (
            <button
              onClick={handleBooking}
              disabled={booking}
              className="w-full bg-primary-600 text-white py-2 px-4 rounded-md hover:bg-primary-700 disabled:opacity-50"
            >
              {booking ? '预订中...' : '确认预订'}
            </button>
          )}

          {message && (
            <div className="mt-4 p-4 rounded bg-blue-100 text-blue-800">
              {message}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
