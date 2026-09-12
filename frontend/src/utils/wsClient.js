// ============================================================
//  可自愈的 WebSocket 客户端（自动重连 + 心跳 + 确定性清理）
//
//  解决的问题：
//   1) 旧实现只有 onopen/onmessage，一旦 onclose/onerror 就永久静默失效
//      （用户以为还连着，实际实时消息/通知再也收不到）。
//   2) WebSocket 实例被塞进 Vuex state，被 reactive() 代理后再调 close()/send()
//      会抛 "Illegal invocation"。-> 本模块把实例保存在闭包里，由调用方用普通
//      模块作用域变量持有，绝不放进 Vuex state。
//   3) 连接/定时器泄漏：stop() 会同时清掉重连定时器与心跳定时器。
//
//  重连策略：指数退避 1s → 2s → 4s ... 上限 30s，叠加 ±20% 抖动（避免多端同时重连惊群）。
//  鉴权失败（服务端 close reason = 'unauthorized'）不再重连，直接停，避免死循环。
//
//  心跳策略（heartbeat 选项）：
//   - 'ping'     ：每个周期发一帧应用层心跳，并期待服务端回帧当作 pong；
//                  连续 2 个周期 + 5s 收不到任何帧就判定连接已死，主动重连。
//                  仅适用于「服务端会把客户端帧回显」的通道（通知通道就是这样）。
//   - 'watchdog' ：不发任何帧，只检查 readyState；掉线由 onclose/onerror/readyState 触发重连。
//                  适用于「服务端不会回显、且发任何帧都会被当成业务消息」的通道（聊天通道）。
//
//  已知局限（需要后端配合才能彻底解决）：
//   浏览器 WebSocket API 没有协议层 ping()，所以无法发真正的控制帧。
//   对 'watchdog' 模式，如果 TCP 连接被“黑洞”（没有 RST/FIN，readyState 一直 OPEN），
//   客户端无法在不发业务帧的前提下探测到——要彻底解决需要后端支持 ping/pong。
// ============================================================

import { logWarn } from './log.js'

// 指数退避：基数、上限、抖动比例
const RECONNECT_BASE_DELAY = 1000
const RECONNECT_MAX_DELAY = 30000
const RECONNECT_JITTER = 0.2

// 心跳周期与 pong 宽限期
const HEARTBEAT_INTERVAL = 30000
const PONG_GRACE = HEARTBEAT_INTERVAL * 2 + 5000

/**
 * 创建一个可自动重连的 WebSocket 客户端。
 *
 * @param {Object}   options
 * @param {string|Function} options.url 连接地址；传函数则每次重连时重新求值
 *                                      （token 可能已刷新，或用户已切换），返回空值表示不再重连。
 * @param {string}   [options.scope]    日志来源标识
 * @param {Function} [options.onMessage] 收到并解析成功后的消息回调
 * @param {Function} [options.onOpen]    每次连接建立后的回调
 * @param {string}   [options.heartbeat] 'ping' | 'watchdog'，默认 'watchdog'
 * @param {Object}   [options.heartbeatPayload] 'ping' 模式发送的心跳帧内容
 * @returns {{start: Function, stop: Function, send: Function, isOpen: Function}}
 */
export function createReconnectingWebSocket(options) {
  const {
    url,
    scope = 'ws',
    onMessage = () => {},
    onOpen = () => {},
    heartbeat = 'watchdog',
    heartbeatPayload = { __heartbeat: true },
  } = options || {}

  let socket = null
  let reconnectTimer = null
  let heartbeatTimer = null
  let attempts = 0
  // 初始为 stopped：调用方必须显式 start()，避免创建即连接导致重复建连
  let stopped = true
  let lastInboundAt = 0

  function resolveUrl() {
    return typeof url === 'function' ? url() : url
  }

  function stopHeartbeat() {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer)
      heartbeatTimer = null
    }
  }

  /**
   * 关闭当前连接。
   * @param {boolean} detach 是否先摘掉事件回调（主动废弃连接时为 true，
   *                         这样不会再触发一次 onclose 重连调度）
   */
  function closeSocket(detach) {
    const current = socket
    socket = null
    if (!current) return

    if (detach) {
      current.onopen = null
      current.onmessage = null
      current.onerror = null
      current.onclose = null
    }

    try {
      current.close()
    } catch (error) {
      logWarn(scope, '关闭连接失败', error)
    }
  }

  function scheduleReconnect() {
    // 幂等：已有待执行的重连计划时不再叠加
    if (stopped || reconnectTimer) return

    attempts += 1
    const base = Math.min(RECONNECT_BASE_DELAY * Math.pow(2, attempts - 1), RECONNECT_MAX_DELAY)
    const jitter = base * RECONNECT_JITTER * (Math.random() * 2 - 1)
    const delay = Math.max(500, Math.round(base + jitter))

    logWarn(scope, `连接已断开，${delay}ms 后进行第 ${attempts} 次重连`)

    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      connect()
    }, delay)
  }

  /** 发送原始帧；返回是否真的写出（连接不可用时返回 false，不抛异常） */
  function sendRaw(payload) {
    if (!socket || socket.readyState !== WebSocket.OPEN) return false
    try {
      socket.send(typeof payload === 'string' ? payload : JSON.stringify(payload))
      return true
    } catch (error) {
      logWarn(scope, '发送消息失败', error)
      return false
    }
  }

  function startHeartbeat() {
    stopHeartbeat()
    heartbeatTimer = setInterval(() => {
      if (stopped || !socket) return

      // 连接已不在 OPEN 状态但 onclose 还没触发：主动废弃并立即重连
      if (socket.readyState !== WebSocket.OPEN) {
        logWarn(scope, `心跳发现连接状态异常(readyState=${socket.readyState})，主动重连`)
        closeSocket(true)
        scheduleReconnect()
        return
      }

      if (heartbeat === 'ping') {
        if (Date.now() - lastInboundAt > PONG_GRACE) {
          logWarn(scope, '心跳超时（长时间未收到服务端任何帧），判定连接已死，主动重连')
          closeSocket(true)
          scheduleReconnect()
          return
        }
        // 服务端会原样回显 → 回显帧刷新 lastInboundAt，等价于 pong
        sendRaw(heartbeatPayload)
      }
    }, HEARTBEAT_INTERVAL)
  }

  function connect() {
    if (stopped) return

    const target = resolveUrl()
    if (!target) {
      // 没有可用地址（通常是没有 token / 已登出）：不做无意义的重连
      logWarn(scope, '缺少连接所需信息（token/用户 id），停止重连')
      stopped = true
      return
    }

    // 保证同一时刻只保留一个连接
    closeSocket(true)

    let ws
    try {
      ws = new WebSocket(target)
    } catch (error) {
      logWarn(scope, '创建 WebSocket 失败', error)
      scheduleReconnect()
      return
    }

    socket = ws
    lastInboundAt = Date.now()

    ws.onopen = () => {
      if (socket !== ws) return
      attempts = 0
      lastInboundAt = Date.now()
      startHeartbeat()
      onOpen()
    }

    ws.onmessage = (event) => {
      if (socket !== ws) return
      lastInboundAt = Date.now()

      let payload
      try {
        payload = JSON.parse(event.data)
      } catch (error) {
        logWarn(scope, '收到无法解析的消息，已忽略', error)
        return
      }

      onMessage(payload)
    }

    ws.onerror = () => {
      // onerror 之后浏览器必定还会触发 onclose，重连统一在 onclose 中调度，避免重复重连
      logWarn(scope, 'WebSocket 出错')
    }

    ws.onclose = (event) => {
      if (socket !== ws) return
      stopHeartbeat()
      socket = null

      if (stopped) return

      // 服务端鉴权失败会主动关闭（reason=unauthorized），此时重连必然再次被拒 → 直接停
      if (event && event.reason === 'unauthorized') {
        logWarn(scope, '服务端拒绝鉴权，停止重连（请重新登录）')
        stopped = true
        return
      }

      scheduleReconnect()
    }
  }

  return {
    /** 启动连接与自动重连（重复调用无副作用） */
    start() {
      if (!stopped) return
      stopped = false
      attempts = 0
      connect()
    },
    /** 停止连接并清理全部定时器（登出/组件卸载时调用） */
    stop() {
      stopped = true
      attempts = 0
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
        reconnectTimer = null
      }
      stopHeartbeat()
      closeSocket(true)
    },
    /** 业务发送；返回布尔值，false 表示当前不可用（调用方应回退到 REST 路径） */
    send(payload) {
      return sendRaw(payload)
    },
    /** 当前是否处于可发送状态 */
    isOpen() {
      return !!socket && socket.readyState === WebSocket.OPEN
    },
  }
}

export default createReconnectingWebSocket
