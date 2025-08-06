'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { AddressSelector } from '@/components/postcode/AddressSelector'

export default function TestPostcodePage() {
  const [selectedPostcode, setSelectedPostcode] = useState('')
  const [selectedAddress, setSelectedAddress] = useState('')

  const handleAddressChange = (postcode: string, fullAddress: string) => {
    setSelectedPostcode(postcode)
    setSelectedAddress(fullAddress)
    console.log('Selected:', { postcode, fullAddress })
  }

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-4xl mx-auto space-y-6">
        {/* 页面标题 */}
        <Card>
          <CardHeader>
            <CardTitle className="text-2xl font-bold text-center">
              OpenPenPal Postcode 系统测试
            </CardTitle>
          </CardHeader>
        </Card>

        {/* 地址选择器测试 */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>地址选择器</CardTitle>
            </CardHeader>
            <CardContent>
              <AddressSelector
                value={selectedPostcode}
                onChange={handleAddressChange}
                placeholder="请选择收件地址..."
              />
            </CardContent>
          </Card>

          {/* 选择结果展示 */}
          <Card>
            <CardHeader>
              <CardTitle>选择结果</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-2">Postcode 编码</label>
                  {selectedPostcode ? (
                    <Badge variant="default" className="text-lg px-4 py-2">
                      {selectedPostcode}
                    </Badge>
                  ) : (
                    <div className="text-gray-500">未选择</div>
                  )}
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">完整地址</label>
                  {selectedAddress ? (
                    <div className="p-3 bg-gray-100 rounded-lg">
                      {selectedAddress}
                    </div>
                  ) : (
                    <div className="text-gray-500">未选择</div>
                  )}
                </div>

                {selectedPostcode && (
                  <div>
                    <label className="block text-sm font-medium mb-2">编码解析</label>
                    <div className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <span>学校编码:</span>
                        <Badge variant="outline">{selectedPostcode.substring(0, 2)}</Badge>
                      </div>
                      <div className="flex justify-between">
                        <span>片区编码:</span>
                        <Badge variant="outline">{selectedPostcode.substring(2, 3)}</Badge>
                      </div>
                      <div className="flex justify-between">
                        <span>楼栋编码:</span>
                        <Badge variant="outline">{selectedPostcode.substring(3, 4)}</Badge>
                      </div>
                      <div className="flex justify-between">
                        <span>房间编码:</span>
                        <Badge variant="outline">{selectedPostcode.substring(4, 6)}</Badge>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* API 测试 */}
        <Card>
          <CardHeader>
            <CardTitle>API 测试</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <h4 className="font-medium mb-2">可用测试编码:</h4>
                <div className="space-y-2">
                  <Badge 
                    variant="outline" 
                    className="cursor-pointer hover:bg-gray-100"
                    onClick={() => handleAddressChange('PK5F3D', '北京大学 第五片区 F栋宿舍 3D宿舍')}
                  >
                    PK5F3D - 北京大学
                  </Badge>
                  <Badge 
                    variant="outline" 
                    className="cursor-pointer hover:bg-gray-100"
                    onClick={() => handleAddressChange('PK5F2A', '北京大学 第五片区 F栋宿舍 2A宿舍')}
                  >
                    PK5F2A - 北京大学
                  </Badge>
                  <Badge 
                    variant="outline" 
                    className="cursor-pointer hover:bg-gray-100"
                    onClick={() => handleAddressChange('QH1C2E', '清华大学 第一片区 C栋宿舍 2E宿舍')}
                  >
                    QH1C2E - 清华大学
                  </Badge>
                </div>
              </div>

              <div>
                <h4 className="font-medium mb-2">搜索测试关键词:</h4>
                <div className="space-y-1 text-sm">
                  <div>• "北京大学" - 搜索学校</div>
                  <div>• "F栋" - 搜索楼栋</div>
                  <div>• "3D宿舍" - 搜索房间</div>
                  <div>• "PK5" - 搜索编码前缀</div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* 说明文档 */}
        <Card>
          <CardHeader>
            <CardTitle>系统说明</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="prose prose-sm max-w-none">
              <h4>Postcode 编码规则 (6位)</h4>
              <ul>
                <li><strong>第1-2位</strong>: 学校站点码 (如: PK=北京大学, QH=清华大学)</li>
                <li><strong>第3位</strong>: 片区码 (如: 5=第五片区)</li>
                <li><strong>第4位</strong>: 楼栋码 (如: F=F栋)</li>
                <li><strong>第5-6位</strong>: 房间号 (如: 3D=3D宿舍)</li>
              </ul>

              <h4>权限对应关系</h4>
              <ul>
                <li><strong>四级信使</strong>: 管理学校级别 (PK, QH)</li>
                <li><strong>三级信使</strong>: 管理片区级别 (PK5, QH1)</li>
                <li><strong>二级信使</strong>: 管理楼栋级别 (PK5F, QH1C)</li>
                <li><strong>一级信使</strong>: 负责具体投递 (PK5F3D, QH1C2E)</li>
              </ul>

              <h4>使用场景</h4>
              <ul>
                <li><strong>用户端</strong>: 写信时选择收件地址</li>
                <li><strong>信使端</strong>: 根据编码权限接收投递任务</li>
                <li><strong>管理端</strong>: 管理地址结构和权限分配</li>
              </ul>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}