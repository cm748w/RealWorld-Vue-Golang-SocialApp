// ============================================================
//  统一的前端日志出口（极简实现）
//
//  背景：store / 组件里原先散落 51 处 console.log/console.error 作为错误处理，
//  生产环境会把这些内部细节打到浏览器控制台（信息泄露 + 噪音），但不能简单删掉
//  ——删掉后 catch 块会变空，既丢掉了可诊断性，也过不了 eslint 的 no-empty。
//
//  策略：生产构建静默、开发环境输出 warn。
//  之所以生产包里能“彻底消失”，是因为这里直接写 `process.env.NODE_ENV !== 'production'`
//  判断：vue-cli 在构建时用 DefinePlugin 把它替换成字面量，terser 随后判定该分支恒假并
//  整段删除（实测 dist/js/app.*.js 里搜不到任何 console.*）。
//  注意：不要把条件提前抽成模块级常量，那样 terser 不会把常量内联进函数体，
//  console.warn 会残留在产物里（此坑已实测踩过）。
// ============================================================

/**
 * 记录非致命错误/异常上下文；生产环境为 no-op。
 *
 * @param {string} scope 来源标识，建议用 '模块.动作' 形式，例如 'Posts.createPost'
 * @param {...any} args  错误对象或附加上下文（仅开发环境输出）
 */
export function logWarn(scope, ...args) {
  if (process.env.NODE_ENV !== 'production') {
    // eslint-disable-next-line no-console
    console.warn(`[${scope}]`, ...args)
  }
}

export default logWarn
