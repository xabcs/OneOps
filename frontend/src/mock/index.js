import Mock from 'mockjs'

// 设置拦截延迟
Mock.setup({
  timeout: 0
})

// 资产分组数据
const groupData = [
  {
    label: '全部资产',
    count: 128,
    children: [
      {
        label: '生产环境',
        count: 64,
        children: [
          { label: '核心网关', count: 12 },
          { label: '数据库集群', count: 8 },
          { label: '应用服务器', count: 44 }
        ]
      },
      {
        label: '测试环境',
        count: 32,
        children: [
          { label: 'QA 节点', count: 20 },
          { label: '预发布', count: 12 }
        ]
      },
      {
        label: '开发环境',
        count: 32
      }
    ]
  }
]

// 模拟主机数据
const servers = Mock.mock({
  'list|50-100': [{
    'id': '@id',
    'name': '@word(3, 5)-@word(5, 8)-@integer(1, 99)',
    'publicIp': '@ip',
    'privateIp': '10.@integer(0, 255).@integer(0, 255).@integer(1, 254)',
    'status|1': ['online', 'offline', 'maintenance'],
    'cpu|1-100': 100,
    'memory|1-100': 100,
    'uptime': '@integer(1, 300)d @integer(1, 23)h',
    'region|1': ['shanghai', 'beijing', 'guangzhou'],
    'group|1': ['核心网关', '数据库集群', '应用服务器', 'QA 节点', '预发布', '开发环境']
  }]
})

// 模拟任务数据
const tasks = Mock.mock({
  'list|30-50': [{
    'id': 't-@integer(1000, 9999)',
    'name|1': ['全量数据库备份', '系统内核更新', '日志自动清理', 'Web 资源同步', '安全漏洞扫描', '服务自动重启'],
    'status|1': ['pending', 'running', 'completed', 'failed'],
    'type|1': ['backup', 'update', 'cleanup', 'sync', 'scan', 'restart'],
    'progress': '@integer(0, 100)',
    'createdAt': '@datetime',
    'executedAt': '@datetime'
  }]
})

// 模拟容器数据
const containers = Mock.mock({
  'list|20-40': [{
    'id': '@guid',
    'name': 'container-@word(3, 5)',
    'image': '@word(5, 8):latest',
    'status|1': ['running', 'stopped', 'restarting', 'paused'],
    'ip': '@ip',
    'ports': '80/tcp, 443/tcp',
    'cpu|1-100': 100,
    'memory|1-100': 100,
    'uptime': '@integer(1, 100)h'
  }]
})

// 模拟操作日志
const operationLogs = Mock.mock({
  'list|50-100': [{
    'id': '@id',
    'user': '@cname',
    'action|1': ['登录系统', '重启服务器', '删除任务', '修改配置', '导出数据'],
    'module|1': ['主机管理', '自动化任务', '监控中心', '系统设置'],
    'path|1': ['/api/login', '/api/servers/restart', '/api/tasks/delete', '/api/config/update', '/api/data/export'],
    'method|1': ['POST', 'GET', 'PUT', 'DELETE'],
    'params': '{"id": "@integer(1, 100)", "name": "@word(5)"}',
    'response': '{"code": 200, "message": "success", "data": {}}',
    'duration|10-2000': 100,
    'ip': '@ip',
    'status|1': ['success', 'failed'],
    'time': '@datetime'
  }]
})

// 模拟登录日志
const loginLogs = Mock.mock({
  'list|50-100': [{
    'id': '@id',
    'username': '@word(4, 8)',
    'ip': '@ip',
    'location': '@city',
    'browser|1': ['Chrome 120', 'Firefox 121', 'Safari 17', 'Edge 120'],
    'os|1': ['Windows 11', 'macOS 14', 'Linux', 'iOS 17', 'Android 14'],
    'status|1': ['success', 'failed'],
    'msg|1': ['登录成功', '密码错误', '账号锁定', '验证码错误'],
    'time': '@datetime'
  }]
})

// 保留非系统管理的 Mock 接口
Mock.mock(/\/api\/groups/, 'get', () => {
  return {
    code: 200,
    data: groupData,
    message: 'success'
  }
})

Mock.mock(/\/api\/servers/, 'get', (options) => {
  // 简单的模拟过滤逻辑
  let list = servers.list

  try {
    const url = new URL(options.url, 'http://localhost')
    const group = url.searchParams.get('group')
    const name = url.searchParams.get('name')
    const status = url.searchParams.get('status')

    if (group && group !== '全部资产') {
      list = list.filter(item => item.group === group || group.includes(item.group))
    }

    if (name) {
      list = list.filter(item =>
        item.name.toLowerCase().includes(name.toLowerCase()) ||
        item.publicIp.includes(name) ||
        item.privateIp.includes(name)
      )
    }

    if (status) {
      list = list.filter(item => item.status === status)
    }
  } catch (e) {
    console.error('Mock URL parsing error:', e)
  }

  return {
    code: 200,
    data: list,
    message: 'success'
  }
})

Mock.mock(/\/api\/servers/, 'post', (options) => {
  let data = {}
  try {
    data = JSON.parse(options.body || '{}')
  } catch (e) {
    console.error('Mock add server parse error:', e)
  }
  const newServer = {
    id: Mock.mock('@id'),
    name: data.name,
    publicIp: data.publicIp || Mock.mock('@ip'),
    privateIp: data.privateIp || '10.0.0.1',
    status: 'online',
    cpu: 0,
    memory: 0,
    uptime: '0d 0h',
    region: data.region || 'shanghai',
    group: data.group || '全部资产'
  }
  servers.list.unshift(newServer)
  return {
    code: 200,
    data: newServer,
    message: 'success'
  }
})

Mock.mock(/\/api\/tasks/, 'get', (options) => {
  let list = tasks.list

  try {
    const url = new URL(options.url, 'http://localhost')
    const name = url.searchParams.get('name')
    const status = url.searchParams.get('status')
    const type = url.searchParams.get('type')

    if (name) {
      list = list.filter(item => item.name.toLowerCase().includes(name.toLowerCase()) || item.id.includes(name))
    }

    if (status) {
      list = list.filter(item => item.status === status)
    }

    if (type) {
      list = list.filter(item => item.type === type)
    }
  } catch (e) {
    console.error('Mock URL parsing error:', e)
  }

  return {
    code: 200,
    data: list,
    message: 'success'
  }
})

Mock.mock(/\/api\/containers/, 'get', (options) => {
  let list = containers.list
  try {
    const url = new URL(options.url, 'http://localhost')
    const name = url.searchParams.get('name')
    const status = url.searchParams.get('status')

    if (name) {
      list = list.filter(item => item.name.toLowerCase().includes(name.toLowerCase()) || item.image.toLowerCase().includes(name.toLowerCase()))
    }
    if (status) {
      list = list.filter(item => item.status === status)
    }
  } catch (e) {
    console.error('Mock URL parsing error:', e)
  }
  return { code: 200, data: list, message: 'success' }
})

Mock.mock(/\/api\/logs\/operation/, 'get', (options) => {
  let list = operationLogs.list
  try {
    const url = new URL(options.url, 'http://localhost')
    const user = url.searchParams.get('user')
    const status = url.searchParams.get('status')
    const module = url.searchParams.get('module')

    if (user) {
      list = list.filter(item => item.user.includes(user))
    }
    if (status) {
      list = list.filter(item => item.status === status)
    }
    if (module) {
      list = list.filter(item => item.module === module)
    }
  } catch (e) {
    console.error('Mock URL parsing error:', e)
  }
  return { code: 200, data: list, message: 'success' }
})

Mock.mock(/\/api\/logs\/login/, 'get', (options) => {
  let list = [...loginLogs.list]
  try {
    const url = new URL(options.url, window.location.origin)
    const username = url.searchParams.get('username')
    const status = url.searchParams.get('status')
    const location = url.searchParams.get('location')

    if (username) {
      list = list.filter(item => item.username.includes(username))
    }
    if (status) {
      list = list.filter(item => item.status === status)
    }
    if (location) {
      list = list.filter(item => item.location.includes(location))
    }
  } catch (e) {
    console.error('Mock URL parsing error:', e)
  }
  return { code: 200, data: list, message: 'success' }
})

Mock.mock(/\/api\/monitoring/, 'get', () => {
  return {
    code: 200,
    data: {
      cpu: Mock.mock('@integer(10, 90)'),
      memory: Mock.mock('@integer(20, 80)'),
      disk: Mock.mock('@integer(30, 70)'),
      network: Mock.mock('@integer(1, 100)'),
      alerts: [
        { id: 1, time: Mock.mock('@datetime'), level: 'warning', source: 'Web-Server-01', message: 'CPU 使用率持续超过 80% (当前 85.4%)', status: 'unhandled' },
        { id: 2, time: Mock.mock('@datetime'), level: 'critical', source: 'DB-Master-01', message: '检测到数据库连接数异常增长，触发限流策略', status: 'unhandled' },
        { id: 3, time: Mock.mock('@datetime'), level: 'info', source: 'Log-Svc', message: '日志存储空间占用超过 70%，建议清理', status: 'unhandled' },
        { id: 4, time: Mock.mock('@datetime'), level: 'critical', source: 'Gateway-02', message: '节点心跳丢失，服务已自动切换至备用节点', status: 'unhandled' }
      ]
    },
    message: 'success'
  }
})

Mock.mock(/\/api\/images/, 'get', () => {
  return {
    code: 200,
    data: [
      { id: 1, name: 'nginx:latest', size: '132MB', tag: 'latest' },
      { id: 2, name: 'redis:6.2', size: '105MB', tag: '6.2' },
      { id: 3, name: 'mysql:8.0', size: '448MB', tag: '8.0' },
      { id: 4, name: 'node:18-alpine', size: '167MB', tag: '18-alpine' }
    ],
    message: 'success'
  }
})

Mock.mock(/\/api\/logs\/event/, 'get', () => {
  return {
    code: 200,
    data: Mock.mock({
      'list|20-40': [{
        'id': '@id',
        'level|1': ['info', 'warning', 'error'],
        'source|1': ['System', 'Network', 'Storage', 'Security'],
        'message': '@sentence(5, 10)',
        'time': '@datetime'
      }]
    }).list,
    message: 'success'
  }
})

export default Mock
