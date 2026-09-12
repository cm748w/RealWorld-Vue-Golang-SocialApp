// 与 frontend/src/views/Auth.vue 的真实 DOM 对齐（旧断言已失效，必失败）：
//  - 表单是两张 col-12 col-md-6 卡片，旧的 .col-5 / .col-7 早已不存在；
//  - 标题文案是 Sign In / Create Account，旧的“登录 / 注册”也不存在。
// 选择器统一走 Auth.vue 里的 data-test 属性，不依赖会随样式调整而变的 class。
describe("Auth Component Tests", () => {
    const SEL = {
        signinCol: '[data-test="auth-signin-col"]',
        signupCol: '[data-test="auth-signup-col"]',
        signinCard: '[data-test="auth-signin-card"]',
        signupCard: '[data-test="auth-signup-card"]',
        signinSubmit: '[data-test="auth-signin-submit"]',
        signupSubmit: '[data-test="auth-signup-submit"]',
    }

    // QInput 的根节点是 <label>，真正的 <input> 在内部；这里对两种落点都兼容，
    // 避免依赖 Quasar 把 $attrs 挂到哪一层的实现细节。
    const input = (name) => `[data-test="${name}"] input, input[data-test="${name}"]`

    const signinEmail = () => cy.get(input('auth-signin-email'))
    const signinPassword = () => cy.get(input('auth-signin-password'))
    const signupInputs = () => cy.get(
        [
            input('auth-signup-firstname'),
            input('auth-signup-lastname'),
            input('auth-signup-email'),
            input('auth-signup-password'),
        ].join(', ')
    )

    beforeEach(() => {
        // 应用以 /app 为基路径（vue.config.js publicPath），dev server 也挂载在 /app 下
        cy.visit('/app/Auth')
    })

    describe('Page Layout', () => {
        it('should display signin and signup cards', () => {
            cy.get(SEL.signinCard).should('be.visible')
            cy.get(SEL.signupCard).should('be.visible')
            cy.get('[data-test$="-card"]').should('have.length', 2)

            // 标题文案（以 Auth.vue 为准）
            cy.get(SEL.signinCard).should('contain', 'Sign In')
            cy.get(SEL.signupCard).should('contain', 'Create Account')
        })

        it('should have proper layout structure', () => {
            cy.get(SEL.signinCol).should('exist')
            cy.get(SEL.signupCol).should('exist')
            cy.get(SEL.signinCol).find('form').should('have.length', 1)
            cy.get(SEL.signupCol).find('form').should('have.length', 1)
        })
    })

    describe('登录表单', () => {
        it('should display all signin form elements', () => {
            cy.get(SEL.signinCard).find('input').should('have.length', 2)

            // 字段文案
            cy.get(SEL.signinCard).should('contain', '邮箱')
            cy.get(SEL.signinCard).should('contain', '密码')

            // 提交按钮
            cy.get(SEL.signinSubmit).should('be.visible').and('have.attr', 'type', 'submit')
        })

        it('should allow typing in signin inputs', () => {
            signinEmail()
                .type('test@example.com')
                .should('have.value', 'test@example.com')

            signinPassword()
                .type('password123')
                .should('have.value', 'password123')
        })

        it('should have password input type', () => {
            signinPassword().should('have.attr', 'type', 'password')
        })
    })

    describe('注册表单', () => {
        it('should display all signup form elements', () => {
            // 检查所有的输入是否存在
            signupInputs().should('have.length', 4)
            const card = cy.get(SEL.signupCard)
            card.should('contain', '名字')
            card.should('contain', '姓氏')
            card.should('contain', '邮箱')
            card.should('contain', '密码')

            // 检查按钮
            cy.get(SEL.signupSubmit).should('be.visible').and('have.attr', 'type', 'submit')
        })

        it('should allow typing in all signup inputs', () => {
            cy.get(input('auth-signup-firstname'))
                .type('John')
                .should('have.value', 'John')

            cy.get(input('auth-signup-lastname'))
                .type('Doe')
                .should('have.value', 'Doe')

            cy.get(input('auth-signup-email'))
                .type('j@example.com')
                .should('have.value', 'j@example.com')

            cy.get(input('auth-signup-password'))
                .type('password123')
                .should('have.value', 'password123')
        })

        it('should have corrent buttom colors', () => {
            // 登录（color="primary"）
            cy.get(SEL.signinSubmit).should('have.class', 'bg-primary')

            // 注册（color="positive"）
            cy.get(SEL.signupSubmit).should('have.class', 'bg-positive')
        })
    })

    describe('登录表单互动', () => {
        it('处理空输入请求', () => {
            cy.get(SEL.signinSubmit).click()
            cy.get('.q-notification').should('be.visible').and('contain', '请输入邮箱')
        })
    })

    describe('注册表单互动', () => {
        it('处理空输入请求', () => {
            cy.get(SEL.signupSubmit).click()
            cy.get('.q-notification').should('be.visible').and('contain', '请输入姓名')
        })
    })

    describe('响应式设计', () => {
        it('多端下维持布局稳定', () => {
            cy.viewport(375, 667)
            cy.get(SEL.signinCard).should('be.visible')

            cy.viewport(768, 1024)
            cy.get(SEL.signinCol).should('be.visible')
            cy.get(SEL.signupCol).should('be.visible')

            cy.viewport(1200, 800)
            cy.get(SEL.signinCard).should('be.visible')
            cy.get(SEL.signupCard).should('be.visible')
        })
    })
})
