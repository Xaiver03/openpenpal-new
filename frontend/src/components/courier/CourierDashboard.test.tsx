import React from 'react'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { CourierDashboard } from './CourierDashboard'
import { useCourierStore } from '@/stores/courier-store'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'

// Mock dependencies
jest.mock('@/stores/courier-store')
jest.mock('next/navigation')
jest.mock('@/lib/api')

// Mock the store
const mockUseCourierStore = useCourierStore as jest.MockedFunction<typeof useCourierStore>
const mockUseRouter = useRouter as jest.MockedFunction<typeof useRouter>
const mockApi = api as jest.Mocked<typeof api>

describe('CourierDashboard', () => {
  const mockPush = jest.fn()
  const mockFetchTasks = jest.fn()
  const mockAcceptTask = jest.fn()
  const mockCompleteTask = jest.fn()

  const mockCourierInfo = {
    id: 'courier-1',
    level: 2,
    zone: 'PK5F',
    points: 850,
    completedTasks: 45,
    rating: 4.8,
  }

  const mockTasks = [
    {
      id: 'task-1',
      letterCode: 'LC123456',
      title: '给朋友的信',
      senderName: '张三',
      targetLocation: '北大5号楼303室',
      pickupOPCode: 'PK5F01',
      deliveryOPCode: 'PK5F03',
      status: 'pending',
      priority: 'normal',
      estimatedTime: 30,
      reward: 10,
    },
    {
      id: 'task-2',
      letterCode: 'LC789012',
      title: '生日祝福',
      senderName: '李四',
      targetLocation: '北大5号楼505室',
      pickupOPCode: 'PK5F02',
      deliveryOPCode: 'PK5F05',
      status: 'accepted',
      priority: 'urgent',
      estimatedTime: 20,
      reward: 15,
      courierID: 'courier-1',
    },
  ]

  beforeEach(() => {
    jest.clearAllMocks()

    mockUseRouter.mockReturnValue({
      push: mockPush,
      back: jest.fn(),
      forward: jest.fn(),
      refresh: jest.fn(),
      replace: jest.fn(),
      prefetch: jest.fn(),
    })

    mockUseCourierStore.mockReturnValue({
      courierInfo: mockCourierInfo,
      pendingTasks: mockTasks.filter(t => t.status === 'pending'),
      myTasks: mockTasks.filter(t => t.courierID === 'courier-1'),
      isLoading: false,
      error: null,
      fetchTasks: mockFetchTasks,
      acceptTask: mockAcceptTask,
      completeTask: mockCompleteTask,
      updateTaskStatus: jest.fn(),
      reset: jest.fn(),
    })
  })

  it('应该显示信使信息面板', () => {
    render(<CourierDashboard />)

    expect(screen.getByText('二级信使')).toBeInTheDocument()
    expect(screen.getByText('PK5F')).toBeInTheDocument()
    expect(screen.getByText('850')).toBeInTheDocument()
    expect(screen.getByText('45')).toBeInTheDocument()
    expect(screen.getByText('4.8')).toBeInTheDocument()
  })

  it('应该显示可接受的任务列表', () => {
    render(<CourierDashboard />)

    expect(screen.getByText('可接受任务')).toBeInTheDocument()
    expect(screen.getByText('给朋友的信')).toBeInTheDocument()
    expect(screen.getByText('张三')).toBeInTheDocument()
    expect(screen.getByText('北大5号楼303室')).toBeInTheDocument()
    expect(screen.getByText('10 积分')).toBeInTheDocument()
  })

  it('应该显示我的任务列表', () => {
    render(<CourierDashboard />)

    const myTasksTab = screen.getByRole('tab', { name: /我的任务/ })
    fireEvent.click(myTasksTab)

    expect(screen.getByText('生日祝福')).toBeInTheDocument()
    expect(screen.getByText('李四')).toBeInTheDocument()
    expect(screen.getByText('紧急')).toBeInTheDocument()
    expect(screen.getByText('已接受')).toBeInTheDocument()
  })

  it('应该处理接受任务', async () => {
    mockAcceptTask.mockResolvedValue({ success: true })
    const user = userEvent.setup()

    render(<CourierDashboard />)

    const acceptButton = screen.getByRole('button', { name: '接受任务' })
    await user.click(acceptButton)

    expect(mockAcceptTask).toHaveBeenCalledWith('task-1')
    
    await waitFor(() => {
      expect(mockFetchTasks).toHaveBeenCalled()
    })
  })

  it('应该显示任务详情对话框', async () => {
    const user = userEvent.setup()
    render(<CourierDashboard />)

    const taskCard = screen.getByText('给朋友的信').closest('[role="button"]')
    await user.click(taskCard!)

    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText('任务详情')).toBeInTheDocument()
    expect(screen.getByText('信件编号：')).toBeInTheDocument()
    expect(screen.getByText('LC123456')).toBeInTheDocument()
    expect(screen.getByText('取件地址：')).toBeInTheDocument()
    expect(screen.getByText('PK5F01')).toBeInTheDocument()
  })

  it('应该处理扫码取件', async () => {
    const user = userEvent.setup()
    render(<CourierDashboard />)

    // 切换到我的任务
    const myTasksTab = screen.getByRole('tab', { name: /我的任务/ })
    await user.click(myTasksTab)

    // 点击扫码取件按钮
    const scanButton = screen.getByRole('button', { name: '扫码取件' })
    await user.click(scanButton)

    // 应该打开扫码对话框
    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText('扫描二维码')).toBeInTheDocument()
  })

  it('应该显示加载状态', () => {
    mockUseCourierStore.mockReturnValue({
      ...mockUseCourierStore(),
      isLoading: true,
    })

    render(<CourierDashboard />)

    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
  })

  it('应该显示错误信息', () => {
    mockUseCourierStore.mockReturnValue({
      ...mockUseCourierStore(),
      error: '获取任务失败',
    })

    render(<CourierDashboard />)

    expect(screen.getByText('获取任务失败')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '重试' })).toBeInTheDocument()
  })

  it('应该根据信使等级显示不同的任务', () => {
    // L1信使只能看到同区域任务
    mockUseCourierStore.mockReturnValue({
      ...mockUseCourierStore(),
      courierInfo: { ...mockCourierInfo, level: 1 },
      pendingTasks: [
        {
          ...mockTasks[0],
          pickupOPCode: 'PK5F01',
          deliveryOPCode: 'PK5F03',
        },
        {
          ...mockTasks[0],
          id: 'task-3',
          pickupOPCode: 'PK3D01', // 不同区域
          deliveryOPCode: 'PK3D03',
        },
      ],
    })

    render(<CourierDashboard />)

    // 应该只显示同区域的任务
    expect(screen.getByText('PK5F01')).toBeInTheDocument()
    expect(screen.queryByText('PK3D01')).not.toBeInTheDocument()
  })

  it('应该显示晋升提示', () => {
    mockUseCourierStore.mockReturnValue({
      ...mockUseCourierStore(),
      courierInfo: {
        ...mockCourierInfo,
        points: 950, // 接近晋升
        nextLevelPoints: 1000,
      },
    })

    render(<CourierDashboard />)

    expect(screen.getByText(/还需 50 积分即可晋升/)).toBeInTheDocument()
  })

  it('应该处理批量操作', async () => {
    const user = userEvent.setup()
    render(<CourierDashboard />)

    // 选择多个任务
    const checkboxes = screen.getAllByRole('checkbox')
    await user.click(checkboxes[0])
    await user.click(checkboxes[1])

    // 批量接受按钮应该可见
    expect(screen.getByRole('button', { name: '批量接受 (2)' })).toBeInTheDocument()
  })

  it('应该刷新任务列表', async () => {
    const user = userEvent.setup()
    render(<CourierDashboard />)

    const refreshButton = screen.getByRole('button', { name: '刷新' })
    await user.click(refreshButton)

    expect(mockFetchTasks).toHaveBeenCalled()
  })

  it('应该按优先级排序任务', () => {
    mockUseCourierStore.mockReturnValue({
      ...mockUseCourierStore(),
      pendingTasks: [
        { ...mockTasks[0], priority: 'normal' },
        { ...mockTasks[0], id: 'task-urgent', priority: 'urgent' },
      ],
    })

    render(<CourierDashboard />)

    const taskElements = screen.getAllByTestId('task-card')
    expect(taskElements[0]).toHaveTextContent('紧急')
    expect(taskElements[1]).toHaveTextContent('普通')
  })
})