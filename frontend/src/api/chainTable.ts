import axios from 'axios'

// 创建axios实例
const api = axios.create({
    baseURL: '/api',
    timeout: 10000,
})

// 请求拦截器
api.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('token')
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    },
    (error) => {
        return Promise.reject(error)
    }
)

// 响应拦截器
api.interceptors.response.use(
    (response) => {
        return response.data
    },
    (error) => {
        if (error.response?.status === 401) {
            localStorage.removeItem('token')
            window.location.href = '/login'
        }
        return Promise.reject(error)
    }
)

// 获取五链四表聚合数据
export const getChainTableData = async (interfaceName?: string) => {
    const params = interfaceName ? {interface: interfaceName} : {}
    return await api.get('/chain-table-data', {params})
}

// 获取网络接口列表
export const getNetworkInterfaces = async () => {
    return await api.get('/network/interfaces')
}

// 获取指定接口的规则统计
export const getInterfaceRuleStats = async (interfaceName: string) => {
    return await api.get(`/network/interfaces/${interfaceName}/rules`)
}

// 获取链的详细信息
export const getChainDetails = async (tableName: string, chainName: string) => {
    return await api.get(`/tables/${tableName}/chains/${chainName}/verbose`)
}

// 获取表的详细信息
export const getTableDetails = async (tableName: string) => {
    return await api.get(`/tables/${tableName}`)
}