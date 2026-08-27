// ============================================================
//  Timing helpers — debounce & throttle
//  Framework-free, works with Vue Options API (methods bound to `this`).
// ============================================================

/**
 * 防抖 debounce
 * 在一段静默时间后才执行；适用于“输入/触发频繁、只关心最后一次结果”的场景，
 * 例如搜索框输入、WebSocket 连续通知触发的刷新。
 *
 * @param {Function} fn      被包裹的函数
 * @param {number}   wait    静默等待毫秒数
 * @param {Object}   opts    { leading: 是否先立即触发一次, trailing: 是否在静默结束后触发 }
 * @returns {Function}       带 this 绑定的防抖函数
 */
export function debounce(fn, wait = 300, { leading = false, trailing = true } = {}) {
  let timer = null
  let lastArgs = null
  let lastThis = null

  const invoke = () => {
    fn.apply(lastThis, lastArgs)
    lastArgs = null
    lastThis = null
  }

  return function debounced(...args) {
    lastArgs = args
    lastThis = this

    if (leading && timer === null) {
      invoke()
    }

    if (timer) {
      clearTimeout(timer)
    }

    timer = setTimeout(() => {
      timer = null
      if (trailing) {
        invoke()
      }
    }, wait)
  }
}

/**
 * 节流 throttle
 * 在固定时间窗口内最多执行一次；适用于“点击/发送等需要限制频率”的场景，
 * 避免快速连点造成请求重复。
 *
 * @param {Function} fn      被包裹的函数
 * @param {number}   wait    时间窗口毫秒数
 * @param {Object}   opts    { leading: 是否在窗口起始立即触发, trailing: 是否在窗口结束后补一次 }
 * @returns {Function}       带 this 绑定的节流函数
 */
export function throttle(fn, wait = 300, { leading = true, trailing = true } = {}) {
  let last = 0
  let timer = null

  return function throttled(...args) {
    const now = Date.now()
    const context = this
    const remaining = wait - (now - last)

    // Keep at most one trailing call pending.
    if (timer !== null) {
      return
    }

    if (remaining <= 0) {
      if (leading) {
        last = now
        fn.apply(context, args)
      } else if (trailing) {
        last = now
        timer = setTimeout(() => {
          timer = null
          last = Date.now()
          fn.apply(context, args)
        }, wait)
      }
    } else if (trailing) {
      timer = setTimeout(() => {
        timer = null
        last = Date.now()
        fn.apply(context, args)
      }, remaining)
    }
  }
}
