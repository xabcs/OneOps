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
    'ip': '@ip',
    'status|1': ['success', 'failed'],
    'time': '@datetime'
  }]
})

// 系统管理数据
const menus = [
  { id: 1, name: '仪表盘概览', icon: 'House', path: '/', permission: 'menu:home', sort: 1, status: 1 },
  { id: 2, name: '资产管理', icon: 'Monitor', path: '/servers', permission: 'menu:servers', sort: 2, status: 1, children: [
    { id: 21, name: '主机管理', icon: 'Monitor', path: '/servers', permission: 'menu:servers', sort: 1, status: 1 },
    { id: 22, name: '添加资产', icon: 'Plus', path: '/servers/add', permission: 'menu:servers', sort: 2, status: 1 }
  ]},
  { id: 3, name: '自动化任务', icon: 'Timer', path: '/tasks', permission: 'menu:tasks', sort: 3, status: 1, children: [
    { id: 31, name: '任务列表', icon: 'Timer', path: '/tasks', permission: 'menu:tasks', sort: 1, status: 1 },
    { id: 32, name: '新建任务', icon: 'Plus', path: '/tasks/create', permission: 'menu:tasks', sort: 2, status: 1 }
  ]},
  { id: 4, name: '监控中心', icon: 'DataLine', path: '/monitoring', permission: 'menu:monitoring', sort: 4, status: 1, children: [
    { id: 41, name: '系统监控', icon: 'DataLine', path: '/monitoring', permission: 'menu:monitoring', sort: 1, status: 1 },
    { id: 42, name: '告警管理', icon: 'Bell', path: '/monitoring/alerts', permission: 'menu:monitoring', sort: 2, status: 1 }
  ]},
  { id: 5, name: '系统管理', icon: 'Setting', path: '/system', permission: 'menu:system', sort: 5, status: 1, children: [
    { id: 51, name: '菜单管理', icon: 'Menu', path: '/system/menus', permission: 'menu:system:menus', sort: 1, status: 1 },
    { id: 52, name: '角色管理', icon: 'UserFilled', path: '/system/roles', permission: 'menu:system:roles', sort: 2, status: 1 },
    { id: 53, name: '用户管理', icon: 'User', path: '/system/users', permission: 'menu:system:users', sort: 3, status: 1 }
  ]}
]

const roles = [
  { id: 1, name: '超级管理员', code: 'admin', description: '拥有系统所有权限', menuIds: [1, 2, 21, 22, 3, 31, 32, 4, 41, 42, 5, 51, 52, 53] },
  { id: 2, name: '运维工程师', code: 'ops', description: '负责主机和任务管理', menuIds: [1, 2, 21, 22, 3, 31, 32, 4, 41, 42] },
  { id: 3, name: '审计员', code: 'auditor', description: '仅拥有查看权限', menuIds: [1, 4, 41] }
]

const users = [
  { id: 1, username: 'admin', nickname: '超级管理员', avatar: '', roleIds: [1], status: 'active', email: 'admin@example.com', createdAt: '2024-01-01' },
  { id: 2, username: 'ops_user', nickname: '运维小王', avatar: '', roleIds: [2], status: 'active', email: 'ops@example.com', createdAt: '2024-02-15' },
  { id: 3, username: 'audit_user', nickname: '审计老李', avatar: '', roleIds: [3], status: 'active', email: 'audit@example.com', createdAt: '2024-03-20' }
]

// 接口拦截
Mock.mock(/\/api\/login/, 'post', (options) => {
  let body = {}
  try {
    body = JSON.parse(options.body || '{}')
  } catch (e) {
    console.error('Mock login parse error:', e)
  }
  const { username, password } = body
  const user = users.find(u => u.username === username)
  
  if (user) {
    // 模拟登录成功
    const userRoles = roles.filter(r => user.roleIds.includes(r.id))
    const isAdmin = userRoles.some(r => r.code === 'admin')
    
    // 如果是 admin，拥有所有菜单 ID
    const allMenuIds = isAdmin 
      ? menus.flatMap(m => [m.id, ...(m.children ? m.children.map(c => c.id) : [])])
      : [...new Set(userRoles.flatMap(r => r.menuIds || []))]
    
    // 递归获取并过滤菜单树
    const getMenuTree = (menuList, targetIds) => {
      return menuList
        .filter(m => (isAdmin || targetIds.includes(m.id)) && m.status === 1)
        .sort((a, b) => a.sort - b.sort)
        .map(m => {
          const node = { ...m }
          if (m.children) {
            node.children = getMenuTree(m.children, targetIds)
            if (node.children.length === 0) delete node.children
          }
          return node
        })
    }

    const menuTree = getMenuTree(menus, allMenuIds)
    const getPermissions = (tree) => {
      let perms = []
      tree.forEach(m => {
        perms.push(m.permission)
        if (m.children) perms = perms.concat(getPermissions(m.children))
      })
      return [...new Set(perms)]
    }
    const permissions = getPermissions(menuTree)
    if (isAdmin) permissions.push('*:*:*')

    return {
      code: 200,
      success: true,
      token: 'mock-token-' + Mock.mock('@guid'),
      user: {
        ...user,
        roleNames: userRoles.map(r => r.name),
        menuTree,
        permissions
      },
      message: '登录成功'
    }
  }
  return {
    code: 401,
    success: false,
    message: '用户名或密码错误'
  }
})

Mock.mock(/\/api\/register/, 'post', () => {
  return {
    code: 200,
    success: true,
    message: '注册成功，请联系管理员审核'
  }
})

Mock.mock(/\/api\/user\/info/, 'get', () => {
  // 模拟当前登录用户为 admin
  const user = users[0]
  const userRoles = roles.filter(r => user.roleIds.includes(r.id))
  const isAdmin = userRoles.some(r => r.code === 'admin')
  
  // 如果是 admin，拥有所有菜单 ID
  const allMenuIds = isAdmin 
    ? menus.flatMap(m => [m.id, ...(m.children ? m.children.map(c => c.id) : [])])
    : [...new Set(userRoles.flatMap(r => r.menuIds || []))]
  
  // 递归获取并过滤菜单树
  const getMenuTree = (menuList, targetIds) => {
    return menuList
      .filter(m => (isAdmin || targetIds.includes(m.id)) && m.status === 1)
      .sort((a, b) => a.sort - b.sort)
      .map(m => {
        const node = { ...m }
        if (m.children) {
          node.children = getMenuTree(m.children, targetIds)
          if (node.children.length === 0) delete node.children
        }
        return node
      })
  }

  const menuTree = getMenuTree(menus, allMenuIds)
  
  // 获取扁平化的权限列表
  const getPermissions = (tree) => {
    let perms = []
    tree.forEach(m => {
      perms.push(m.permission)
      if (m.children) {
        perms = perms.concat(getPermissions(m.children))
      }
    })
    return [...new Set(perms)]
  }

  const permissions = getPermissions(menuTree)
  if (isAdmin) permissions.push('*:*:*')

  return {
    code: 200,
    data: {
      ...user,
      roleNames: userRoles.map(r => r.name),
      menuTree: menuTree,
      permissions: permissions
    },
    message: 'success'
  }
})

Mock.mock(/\/api\/system\/menus/, 'get', () => {
  // 排序逻辑
  const sortMenus = (list) => {
    return list.sort((a, b) => a.sort - b.sort).map(item => {
      if (item.children) {
        item.children = sortMenus(item.children)
      }
      return item
    })
  }
  const safeDeepCopy = (obj) => {
    try {
      return JSON.parse(JSON.stringify(obj))
    } catch (e) {
      console.error('Mock data contains circular structure:', e)
      return obj // Fallback to original object if stringification fails
    }
  }
  return { code: 200, data: sortMenus(safeDeepCopy(menus)), message: 'success' }
})

Mock.mock(/\/api\/system\/menus/, 'post', (options) => {
  let data = {}
  try {
    data = JSON.parse(options.body || '{}')
  } catch (e) {
    console.error('Mock add menu parse error:', e)
  }
  const newMenu = {
    ...data,
    id: Mock.mock('@integer(100, 999)'),
    status: data.status ?? 1,
    sort: data.sort ?? 1
  }
  menus.push(newMenu)
  return { code: 200, data: newMenu, message: 'success' }
})

Mock.mock(/\/api\/system\/menus\/\d+/, 'put', (options) => {
  const id = parseInt(options.url.split('/').pop())
  let data = {}
  try {
    data = JSON.parse(options.body || '{}')
  } catch (e) {
    console.error('Mock update menu parse error:', e)
  }
  
  const updateMenu = (list) => {
    for (let i = 0; i < list.length; i++) {
      if (list[i].id === id) {
        list[i] = { ...list[i], ...data }
        return true
      }
      if (list[i].children && updateMenu(list[i].children)) return true
    }
    return false
  }
  
  updateMenu(menus)
  return { code: 200, data: null, message: 'success' }
})

Mock.mock(/\/api\/system\/menus\/\d+/, 'delete', (options) => {
  const id = parseInt(options.url.split('/').pop())
  const deleteMenu = (list) => {
    const index = list.findIndex(m => m.id === id)
    if (index !== -1) {
      list.splice(index, 1)
      return true
    }
    for (let item of list) {
      if (item.children && deleteMenu(item.children)) return true
    }
    return false
  }
  deleteMenu(menus)
  return { code: 200, data: null, message: 'success' }
})
Mock.mock(/\/api\/system\/roles/, 'get', () => ({ code: 200, data: roles, message: 'success' }))
Mock.mock(/\/api\/system\/roles/, 'post', (options) => {
  let data = {}
  try {
    data = JSON.parse(options.body || '{}')
  } catch (e) {
    console.error('Mock add role parse error:', e)
  }
  const newRole = {
    ...data,
    id: roles.length > 0 ? Math.max(...roles.map(r => r.id)) + 1 : 1,
    menuIds: data.menuIds || []
  }
  roles.push(newRole)
  return { code: 200, data: newRole, message: 'success' }
})
Mock.mock(/\/api\/system\/roles\/\d+/, 'put', (options) => {
  const id = parseInt(options.url.split('/').pop())
  let data = {}
  try {
    data = JSON.parse(options.body || '{}')
  } catch (e) {
    console.error('Mock update role parse error:', e)
  }
  const index = roles.findIndex(r => r.id === id)
  if (index !== -1) {
    roles[index] = { ...roles[index], ...data }
  }
  return { code: 200, data: null, message: 'success' }
})
Mock.mock(/\/api\/system\/roles\/\d+/, 'delete', (options) => {
  const id = parseInt(options.url.split('/').pop())
  const index = roles.findIndex(r => r.id === id)
  if (index !== -1) {
    roles.splice(index, 1)
  }
  return { code: 200, data: null, message: 'success' }
})

Mock.mock(/\/api\/system\/users/, 'get', () => ({ code: 200, data: users, message: 'success' }))
Mock.mock(/\/api\/system\/users/, 'post', (options) => {
  let data = {}
  try {
    data = JSON.parse(options.body || '{}')
  } catch (e) {
    console.error('Mock add user parse error:', e)
  }
  const newUser = {
    ...data,
    id: users.length > 0 ? Math.max(...users.map(u => u.id)) + 1 : 1,
    createdAt: new Date().toISOString().split('T')[0]
  }
  users.push(newUser)
  return { code: 200, data: newUser, message: 'success' }
})
Mock.mock(/\/api\/system\/users\/\d+/, 'put', (options) => {
  const id = parseInt(options.url.split('/').pop())
  let data = {}
  try {
    data = JSON.parse(options.body || '{}')
  } catch (e) {
    console.error('Mock update user parse error:', e)
  }
  const index = users.findIndex(u => u.id === id)
  if (index !== -1) {
    users[index] = { ...users[index], ...data }
  }
  return { code: 200, data: null, message: 'success' }
})
Mock.mock(/\/api\/system\/users\/\d+/, 'delete', (options) => {
  const id = parseInt(options.url.split('/').pop())
  const index = users.findIndex(u => u.id === id)
  if (index !== -1) {
    users.splice(index, 1)
  }
  return { code: 200, data: null, message: 'success' }
})

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
