'use client'

import { useState, useEffect, useCallback } from 'react'
import { Search, MapPin, Building, Home, AlertCircle, Plus } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import PostcodeService from '@/lib/services/postcode-service'
import type {
  SchoolSite,
  SiteArea,
  AreaBuilding,
  BuildingRoom,
  AddressSearchResult,
  AddressSelectorConfig
} from '@/lib/types/postcode'

interface AddressSelectorProps {
  value?: string // 当前选中的 postcode
  onChange: (postcode: string, fullAddress: string) => void
  config?: Partial<AddressSelectorConfig>
  className?: string
  placeholder?: string
  disabled?: boolean
}

interface AddressState {
  school?: SchoolSite
  area?: SiteArea
  building?: AreaBuilding
  room?: BuildingRoom
}

export function AddressSelector({
  value,
  onChange,
  config = {},
  className = '',
  placeholder = '请选择收件地址...',
  disabled = false
}: AddressSelectorProps) {
  // 配置默认值
  const selectorConfig: AddressSelectorConfig = {
    showSchoolSelection: true,
    allowNewAddressRequest: true,
    maxSearchResults: 10,
    enableFuzzySearch: true,
    requiredLevels: ['school', 'area', 'building', 'room'],
    ...config
  }

  // 状态管理
  const [address, setAddress] = useState<AddressState>({})
  const [schools, setSchools] = useState<SchoolSite[]>([])
  const [areas, setAreas] = useState<SiteArea[]>([])
  const [buildings, setBuildings] = useState<AreaBuilding[]>([])
  const [rooms, setRooms] = useState<BuildingRoom[]>([])
  
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<AddressSearchResult[]>([])
  const [isSearching, setIsSearching] = useState(false)
  const [showSearch, setShowSearch] = useState(false)
  
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [showNewAddressDialog, setShowNewAddressDialog] = useState(false)
  const [newAddressRequest, setNewAddressRequest] = useState('')

  // 初始化加载学校列表
  useEffect(() => {
    const loadSchools = async () => {
      try {
        setLoading(true)
        const response = await PostcodeService.getSchoolSites()
        if (response.success && response.data) {
          setSchools(response.data)
        }
      } catch (error) {
        console.error('Failed to load schools:', error)
        setError('加载学校列表失败')
      } finally {
        setLoading(false)
      }
    }

    loadSchools()
  }, [])

  // 根据现有值初始化地址
  useEffect(() => {
    if (value) {
      const initializeFromPostcode = async () => {
        try {
          const response = await PostcodeService.getAddressByPostcode(value)
          if (response.success && response.data) {
            const result = response.data
            setAddress({
              school: result.hierarchy.school,
              area: result.hierarchy.area,
              building: result.hierarchy.building,
              room: result.hierarchy.room
            })
          }
        } catch (error) {
          console.error('Failed to initialize address:', error)
        }
      }

      initializeFromPostcode()
    }
  }, [value])

  // 学校选择变化
  const handleSchoolChange = useCallback(async (schoolCode: string) => {
    const school = schools.find(s => s.code === schoolCode)
    if (!school) return

    setAddress({ school })
    setAreas([])
    setBuildings([])
    setRooms([])

    try {
      setLoading(true)
      const response = await PostcodeService.getSchoolAreas(schoolCode)
      if (response.success && response.data) {
        setAreas(response.data)
      }
    } catch (error) {
      console.error('Failed to load areas:', error)
      setError('加载片区列表失败')
    } finally {
      setLoading(false)
    }
  }, [schools])

  // 片区选择变化
  const handleAreaChange = useCallback(async (areaCode: string) => {
    if (!address.school) return

    const area = areas.find(a => a.code === areaCode)
    if (!area) return

    setAddress(prev => ({ ...prev, area }))
    setBuildings([])
    setRooms([])

    try {
      setLoading(true)
      const response = await PostcodeService.getAreaBuildings(address.school.code, areaCode)
      if (response.success && response.data) {
        setBuildings(response.data)
      }
    } catch (error) {
      console.error('Failed to load buildings:', error)
      setError('加载楼栋列表失败')
    } finally {
      setLoading(false)
    }
  }, [address.school, areas])

  // 楼栋选择变化
  const handleBuildingChange = useCallback(async (buildingCode: string) => {
    if (!address.school || !address.area) return

    const building = buildings.find(b => b.code === buildingCode)
    if (!building) return

    setAddress(prev => ({ ...prev, building }))
    setRooms([])

    try {
      setLoading(true)
      const response = await PostcodeService.getBuildingRooms(
        address.school.code,
        address.area.code,
        buildingCode
      )
      if (response.success && response.data) {
        setRooms(response.data)
      }
    } catch (error) {
      console.error('Failed to load rooms:', error)
      setError('加载房间列表失败')
    } finally {
      setLoading(false)
    }
  }, [address.school, address.area, buildings])

  // 房间选择变化
  const handleRoomChange = useCallback((roomCode: string) => {
    const room = rooms.find(r => r.code === roomCode)
    if (!room) return

    setAddress(prev => ({ ...prev, room }))
    
    // 触发完整地址选择
    const fullAddress = `${address.school?.name} ${address.area?.name} ${address.building?.name} ${room.name}`
    onChange(room.fullPostcode, fullAddress)
  }, [rooms, address.school, address.area, address.building, onChange])

  // 地址搜索
  const handleSearch = useCallback(async (query: string) => {
    if (!query.trim()) {
      setSearchResults([])
      return
    }

    setIsSearching(true)
    try {
      const response = await PostcodeService.searchAddresses(query, selectorConfig.maxSearchResults)
      if (response.success && response.data) {
        setSearchResults(response.data)
      }
    } catch (error) {
      console.error('Failed to search addresses:', error)
      setError('地址搜索失败')
    } finally {
      setIsSearching(false)
    }
  }, [selectorConfig.maxSearchResults])

  // 搜索结果选择
  const handleSearchResultSelect = (result: AddressSearchResult) => {
    setAddress({
      school: result.hierarchy.school,
      area: result.hierarchy.area,
      building: result.hierarchy.building,
      room: result.hierarchy.room
    })
    setShowSearch(false)
    setSearchQuery('')
    setSearchResults([])
    onChange(result.postcode, result.fullAddress)
  }

  // 提交新地址请求
  const handleNewAddressSubmit = async () => {
    if (!newAddressRequest.trim()) return

    try {
      await PostcodeService.submitAddressFeedback({
        type: 'new_address',
        description: newAddressRequest,
        suggestedAddress: {
          schoolCode: address.school?.code || '',
          areaCode: address.area?.code,
          buildingCode: address.building?.code,
          roomCode: '',
          name: newAddressRequest
        },
        submittedBy: 'current_user', // TODO: 从认证上下文获取
        submitterType: 'user'
      })

      setShowNewAddressDialog(false)
      setNewAddressRequest('')
      alert('新地址请求已提交，待管理员审核')
    } catch (error) {
      console.error('Failed to submit new address request:', error)
      setError('提交新地址请求失败')
    }
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {/* 搜索模式切换 */}
      <div className="flex gap-2">
        <Button
          variant={!showSearch ? "default" : "outline"}
          size="sm"
          onClick={() => setShowSearch(false)}
          disabled={disabled}
        >
          <MapPin className="w-4 h-4 mr-2" />
          分级选择
        </Button>
        <Button
          variant={showSearch ? "default" : "outline"}
          size="sm"
          onClick={() => setShowSearch(true)}
          disabled={disabled}
        >
          <Search className="w-4 h-4 mr-2" />
          搜索地址
        </Button>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* 搜索模式 */}
      {showSearch && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">地址搜索</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
              <Input
                placeholder="输入学校、楼栋或房间名称..."
                value={searchQuery}
                onChange={(e) => {
                  setSearchQuery(e.target.value)
                  handleSearch(e.target.value)
                }}
                className="pl-10"
                disabled={disabled}
              />
            </div>

            {isSearching && <div className="text-sm text-gray-500">搜索中...</div>}

            {searchResults.length > 0 && (
              <div className="space-y-2 max-h-60 overflow-y-auto">
                {searchResults.map((result, index) => (
                  <div
                    key={index}
                    className="p-3 border rounded-lg cursor-pointer hover:bg-gray-50 transition-colors"
                    onClick={() => handleSearchResultSelect(result)}
                  >
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="font-medium">{result.fullAddress}</div>
                        <div className="text-sm text-gray-500 mt-1">
                          <Badge variant="outline">{result.postcode}</Badge>
                        </div>
                      </div>
                      <div className="text-sm text-gray-400">
                        匹配度: {Math.round(result.matchScore * 100)}%
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* 分级选择模式 */}
      {!showSearch && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">分级地址选择</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* 学校选择 */}
            {selectorConfig.showSchoolSelection && (
              <div>
                <label className="text-sm font-medium mb-2 block">学校</label>
                <Select
                  value={address.school?.code || ''}
                  onValueChange={handleSchoolChange}
                  disabled={disabled || loading}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择学校" />
                  </SelectTrigger>
                  <SelectContent>
                    {schools.map((school) => (
                      <SelectItem key={school.code} value={school.code}>
                        <div className="flex items-center gap-2">
                          <Building className="w-4 h-4" />
                          {school.name}
                          <Badge variant="outline">{school.code}</Badge>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            {/* 片区选择 */}
            {address.school && (
              <div>
                <label className="text-sm font-medium mb-2 block">片区</label>
                <Select
                  value={address.area?.code || ''}
                  onValueChange={handleAreaChange}
                  disabled={disabled || loading}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择片区" />
                  </SelectTrigger>
                  <SelectContent>
                    {areas.map((area) => (
                      <SelectItem key={area.code} value={area.code}>
                        <div className="flex items-center gap-2">
                          <MapPin className="w-4 h-4" />
                          {area.name}
                          <Badge variant="outline">{area.code}</Badge>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            {/* 楼栋选择 */}
            {address.area && (
              <div>
                <label className="text-sm font-medium mb-2 block">楼栋</label>
                <Select
                  value={address.building?.code || ''}
                  onValueChange={handleBuildingChange}
                  disabled={disabled || loading}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择楼栋" />
                  </SelectTrigger>
                  <SelectContent>
                    {buildings.map((building) => (
                      <SelectItem key={building.code} value={building.code}>
                        <div className="flex items-center gap-2">
                          <Building className="w-4 h-4" />
                          {building.name}
                          <Badge variant="outline">{building.code}</Badge>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            {/* 房间选择 */}
            {address.building && (
              <div>
                <label className="text-sm font-medium mb-2 block">房间</label>
                <Select
                  value={address.room?.code || ''}
                  onValueChange={handleRoomChange}
                  disabled={disabled || loading}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择房间" />
                  </SelectTrigger>
                  <SelectContent>
                    {rooms.map((room) => (
                      <SelectItem key={room.code} value={room.code}>
                        <div className="flex items-center gap-2">
                          <Home className="w-4 h-4" />
                          {room.name}
                          <Badge variant="outline">{room.fullPostcode}</Badge>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* 当前选择显示 */}
      {address.room && (
        <Card className="bg-green-50 border-green-200">
          <CardContent className="pt-4">
            <div className="flex items-center justify-between">
              <div>
                <div className="font-medium text-green-800">已选择地址</div>
                <div className="text-sm text-green-700 mt-1">
                  {address.school?.name} {address.area?.name} {address.building?.name} {address.room?.name}
                </div>
              </div>
              <Badge className="bg-green-600">{address.room?.fullPostcode}</Badge>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 新地址请求 */}
      {selectorConfig.allowNewAddressRequest && (
        <Dialog open={showNewAddressDialog} onOpenChange={setShowNewAddressDialog}>
          <DialogTrigger asChild>
            <Button variant="outline" size="sm" disabled={disabled}>
              <Plus className="w-4 h-4 mr-2" />
              找不到地址？申请新增
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>申请新增地址</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium mb-2 block">地址描述</label>
                <Textarea
                  placeholder="请详细描述您需要的地址信息..."
                  value={newAddressRequest}
                  onChange={(e) => setNewAddressRequest(e.target.value)}
                  rows={4}
                />
              </div>
              <div className="flex gap-2 justify-end">
                <Button variant="outline" onClick={() => setShowNewAddressDialog(false)}>
                  取消
                </Button>
                <Button onClick={handleNewAddressSubmit} disabled={!newAddressRequest.trim()}>
                  提交申请
                </Button>
              </div>
            </div>
          </DialogContent>
        </Dialog>
      )}
    </div>
  )
}

export default AddressSelector