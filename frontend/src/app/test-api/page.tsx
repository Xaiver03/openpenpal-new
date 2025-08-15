'use client'

import { useState } from 'react'
import { followApi } from '@/lib/api/follow'
import { privacyApi } from '@/lib/api/privacy'

export default function TestApiPage() {
  const [results, setResults] = useState<string[]>([])
  const [token, setToken] = useState('')
  const [userId, setUserId] = useState('')

  const addResult = (result: string) => {
    setResults(prev => [...prev, `${new Date().toLocaleTimeString()}: ${result}`])
  }

  const testFollowApi = async () => {
    addResult('Testing Follow API...')
    
    try {
      // Test user suggestions (doesn't require auth)
      const suggestions = await followApi.getUserSuggestions({ limit: 5 })
      addResult(`✅ User suggestions: ${JSON.stringify(suggestions)}`)
    } catch (error: any) {
      addResult(`❌ User suggestions failed: ${error.message}`)
    }

    if (token && userId) {
      try {
        // Test follow status
        const status = await followApi.getFollowStatus(userId)
        addResult(`✅ Follow status: ${JSON.stringify(status)}`)
      } catch (error: any) {
        addResult(`❌ Follow status failed: ${error.message}`)
      }
    } else {
      addResult('⚠️ Set token and user ID to test authenticated endpoints')
    }
  }

  const testPrivacyApi = async () => {
    addResult('Testing Privacy API...')
    
    if (token) {
      try {
        // Test get privacy settings
        const settings = await privacyApi.getPrivacySettings()
        addResult(`✅ Privacy settings: ${JSON.stringify(settings)}`)
      } catch (error: any) {
        addResult(`❌ Privacy settings failed: ${error.message}`)
      }
    } else {
      addResult('⚠️ Set token to test privacy endpoints')
    }
  }

  const testDirectApi = async () => {
    addResult('Testing Direct API calls...')
    
    try {
      // Test backend health
      const response = await fetch('/api/health')
      const text = await response.text()
      addResult(`✅ Backend health (${response.status}): ${text}`)
    } catch (error: any) {
      addResult(`❌ Backend health failed: ${error.message}`)
    }

    try {
      // Test API proxy
      const response = await fetch('/api/v1/health')
      const text = await response.text()
      addResult(`✅ API v1 health (${response.status}): ${text}`)
    } catch (error: any) {
      addResult(`❌ API v1 health failed: ${error.message}`)
    }
  }

  const createTestUser = async () => {
    addResult('Creating test user...')
    
    const testUser = {
      username: `testuser_${Date.now()}`,
      password: 'Test@123456',
      email: `test${Date.now()}@example.com`,
      nickname: 'Test User',
      school: 'Test School'
    }

    try {
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(testUser)
      })
      const data = await response.json()
      
      if (response.ok) {
        addResult(`✅ User created: ${testUser.username}`)
        const extractedToken = data.data?.token || data.token
        const extractedUserId = data.data?.user?.id || data.user?.id
        if (extractedToken) {
          setToken(extractedToken)
          addResult(`✅ Token saved: ${extractedToken.substring(0, 20)}...`)
        }
        if (extractedUserId) {
          setUserId(extractedUserId.toString())
          addResult(`✅ User ID saved: ${extractedUserId}`)
        }
      } else {
        addResult(`❌ User creation failed: ${JSON.stringify(data)}`)
      }
    } catch (error: any) {
      addResult(`❌ User creation error: ${error.message}`)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-8">API Integration Test</h1>
        
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Configuration</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Auth Token</label>
              <input
                type="text"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                placeholder="Bearer token (optional)"
                className="w-full px-3 py-2 border rounded-md"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">User ID</label>
              <input
                type="text"
                value={userId}
                onChange={(e) => setUserId(e.target.value)}
                placeholder="Target user ID (optional)"
                className="w-full px-3 py-2 border rounded-md"
              />
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Test Actions</h2>
          <div className="flex flex-wrap gap-3">
            <button
              onClick={createTestUser}
              className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
            >
              Create Test User
            </button>
            <button
              onClick={testDirectApi}
              className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            >
              Test Direct API
            </button>
            <button
              onClick={testFollowApi}
              className="px-4 py-2 bg-purple-600 text-white rounded hover:bg-purple-700"
            >
              Test Follow API
            </button>
            <button
              onClick={testPrivacyApi}
              className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
            >
              Test Privacy API
            </button>
            <button
              onClick={() => setResults([])}
              className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700"
            >
              Clear Results
            </button>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">Test Results</h2>
          <div className="bg-gray-50 rounded p-4 h-96 overflow-y-auto">
            {results.length === 0 ? (
              <p className="text-gray-500">No test results yet. Click a test button above.</p>
            ) : (
              <div className="space-y-2">
                {results.map((result, index) => (
                  <div key={index} className="text-sm font-mono">
                    {result}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}