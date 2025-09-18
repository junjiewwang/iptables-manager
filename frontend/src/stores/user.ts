import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const username = ref(localStorage.getItem('username') || '')
  const isLoggedIn = ref(!!token.value)

  const login = async (loginData: { username: string; password: string }) => {
    try {
      const response = await axios.post('/api/login', loginData)
      const { access_token, username: user } = response.data
      
      token.value = access_token
      username.value = user
      isLoggedIn.value = true
      
      localStorage.setItem('token', access_token)
      localStorage.setItem('username', user)
      
      // 设置axios默认header
      axios.defaults.headers.common['Authorization'] = `Bearer ${access_token}`
      
      return true
    } catch (error) {
      console.error('Login failed:', error)
      return false
    }
  }

  const logout = () => {
    token.value = ''
    username.value = ''
    isLoggedIn.value = false
    
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    
    delete axios.defaults.headers.common['Authorization']
  }

  // 初始化时设置axios header
  if (token.value) {
    axios.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
  }

  return {
    token,
    username,
    isLoggedIn,
    login,
    logout
  }
})